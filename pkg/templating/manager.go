package templating

import (
	"bufio"
	"context"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/amenyxia/Sarracenia/pkg/markov"
)

var (
	wordCount     int
	wordList      []string
	loadWordsOnce sync.Once
	loadErr       error
)

// InitWordList loads the global dictionary of words from a file at the given path.
// It is designed to be called once at application startup. It uses a sync.Once
// to ensure the word list is loaded only a single time, making subsequent calls
// a no-op. An error is returned if the file cannot be read.
func InitWordList(path string) error {
	loadWordsOnce.Do(func() {
		var words []string
		file, err := os.Open(path)
		if err != nil {
			loadErr = err
			wordList = []string{"fallback"}
			return
		}
		defer func(file *os.File) {
			_ = file.Close()
		}(file)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			words = append(words, scanner.Text())
		}

		if err = scanner.Err(); err != nil {
			loadErr = err
			wordList = []string{"fallback"}
			return
		}

		wordList = words
	})
	wordCount = len(wordList)
	return loadErr
}

// TemplateManager is the central controller for the templating engine.
// It manages the template set, configuration, function map, and connections
// to other services like the Markov generator. It is responsible for loading,
// parsing, and executing templates in a concurrent-safe manner.
// All methods are concurrent-safe.
type TemplateManager struct {
	logger         *slog.Logger
	config         *TemplateConfig
	whitelistMap   map[string]struct{}
	markovGen      *markov.Generator
	markovModels   map[string]markov.ModelInfo
	templates      *template.Template
	cleanTemplates *template.Template
	templateNames  []string
	funcMap        template.FuncMap
	templateDir    string
	mu             sync.RWMutex
}

// NewTemplateManager creates, initializes, and returns a new TemplateManager.
// It requires a logger, an optional Markov generator (can be nil if config.MarkovEnabled
// is false), a configuration, and the path to the data directory which must contain
// a "templates" subdirectory and a "wordlist.txt" file. It performs an initial
// Refresh to load all templates and models.
func NewTemplateManager(logger *slog.Logger, markovGen *markov.Generator, config *TemplateConfig, dataDir string) (*TemplateManager, error) {

	var wordListFile, templateDir string

	wordListFile = filepath.Join(dataDir, "wordlist.txt")
	templateDir = filepath.Join(dataDir, "templates")

	if err := InitWordList(wordListFile); err != nil {
		return nil, err
	}

	// Do note that this uses a default config automatically,
	// custom configs must be injected in via SetConfig
	tm := &TemplateManager{
		logger:       logger,
		markovGen:    markovGen,
		templateDir:  templateDir,
		config:       config,
		whitelistMap: map[string]struct{}{},
	}
	tm.funcMap = tm.makeFuncMap()

	if err := tm.Refresh(); err != nil {
		return nil, err
	}

	logger.Info("Template manager initialized")
	return tm, nil
}

func (tm *TemplateManager) makeFuncMap() template.FuncMap {
	return template.FuncMap{
		// Content Generation (from funcs_content.go)
		"markovSentence":   tm.markovSentence,
		"markovParagraphs": tm.markovParagraphs,
		"randomWord":       randomWord,
		"randomSentence":   randomSentence,
		"randomParagraphs": randomParagraphs,
		"randomString":     randomString,
		"randomDate":       randomDate,
		"randomJSON":       tm.randomJSON,

		// Structure & Composition (from funcs_structure.go)
		"randomForm":           tm.randomForm,
		"randomDefinitionData": randomDefinitionData,
		"nestDivs":             tm.nestDivs,
		"randomComplexTable":   tm.randomComplexTable,

		// Styling (from funcs_styling.go)
		"randomColor":       randomColor,
		"randomId":          randomId,
		"randomClasses":     randomClasses,
		"randomCSSStyle":    randomCSSStyle,
		"randomInlineStyle": randomInlineStyle,

		// Link & Navigation (from funcs_links.go)
		"randomLink":      tm.randomLink,
		"randomQueryLink": tm.randomQueryLink,

		// Logic & Control (from funcs_logic.go)
		"repeat":       repeat,
		"list":         list,
		"randomChoice": randomChoice,
		"randomInt":    randomInt,

		// Simple (from funcs_simple.go)
		"add":   add,
		"sub":   sub,
		"div":   div,
		"mult":  mult,
		"max":   max,
		"min":   min,
		"mod":   mod,
		"inc":   inc,
		"dec":   dec,
		"and":   and,
		"or":    or,
		"not":   not,
		"isSet": isSet,

		// Computationally Expensive (from funcs_expensive.go)
		"randomStyleBlock":     tm.randomStyleBlock,
		"randomCSSVars":        tm.randomCSSVars,
		"randomSVG":            tm.randomSVG,
		"jsInteractiveContent": tm.jsInteractiveContent,
	}
}

// SetConfig applies a new configuration to the TemplateManager. This allows for
// changes to the engine's behavior, such as updating safety limits or
// the path whitelist, without needing to restart the application.
func (tm *TemplateManager) SetConfig(config *TemplateConfig) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.config = config
	for _, path := range config.PathWhitelist {
		tm.whitelistMap[path] = struct{}{}
	}
	if tm.config.MarkovEnabled {
		var opts []markov.Option
		if tm.config.MarkovSeparator != "" {
			opts = append(opts, markov.WithSeparator(tm.config.MarkovSeparator))
		}
		if tm.config.MarkovEoc != "" {
			opts = append(opts, markov.WithEOC(tm.config.MarkovEoc))
		}
		if tm.config.MarkovSplitRegex != "" {
			opts = append(opts, markov.WithSeparatorRegex(tm.config.MarkovSplitRegex))
		}
		if tm.config.MarkovEocRegex != "" {
			opts = append(opts, markov.WithEOCRegex(tm.config.MarkovEocRegex))
		}
		if tm.config.MarkovSeparatorExcRegex != "" {
			opts = append(opts, markov.WithSeparatorExcRegex(tm.config.MarkovSeparatorExcRegex))
		}
		if tm.config.MarkovEocExcRegex != "" {
			opts = append(opts, markov.WithEOCExcRegex(tm.config.MarkovEocExcRegex))
		}
		tm.markovGen.SetTokenizer(markov.NewDefaultTokenizer(opts...))
	}
}

// Refresh reloads all templates from the filesystem and, if enabled, refreshes
// the list of available Markov models from the database. This function allows for
// updates to templates and models without restarting the application.
func (tm *TemplateManager) Refresh() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	filePattern := filepath.Join(tm.templateDir, "*.tmpl.html")
	tm.logger.Info("Loading template files...")

	parsedFiles, err := template.New("").Funcs(tm.funcMap).ParseGlob(filePattern)
	var names []string
	if err != nil {
		if !strings.Contains(err.Error(), "pattern matches no files") {
			tm.logger.Error("failed to parse template files", "error", err)
			return err
		} else {
			// No template files, so we have to create the object without any
			parsedFiles = template.New("").Funcs(tm.funcMap)
			names = []string{}
		}
	} else {
		for _, t := range parsedFiles.Templates() {
			// By default, there is a root template with no name. We don't want to execute this
			if strings.Contains(t.Name(), ".tmpl.html") {
				names = append(names, t.Name())
			}
		}
	}

	filePattern = filepath.Join(tm.templateDir, "*.part.html")
	tm.logger.Info("Loading partial files...")

	newParsedFiles, err := parsedFiles.ParseGlob(filePattern)
	if err != nil {
		if !strings.Contains(err.Error(), "pattern matches no files") {
			tm.logger.Error("failed to parse partial files", "error", err)
			return err
		} else {
			newParsedFiles = parsedFiles
		}
	}
	// We skip the for loop here because templateNames is only for full templates

	if len(names) == 0 {
		tm.logger.Warn("No template files found matching pattern", "pattern", filePattern)
	}

	tm.templates = newParsedFiles
	tm.templateNames = names
	tm.logger.Info("Loaded template and partial files", "count", len(parsedFiles.Templates())-1) // Subtract one for the root template

	// Create a clean clone for string executions after all parsing is complete.
	tm.cleanTemplates, err = tm.templates.Clone()
	if err != nil {
		tm.logger.Error("failed to create a clean clone of templates", "error", err)
		return err
	}

	if tm.config.MarkovEnabled {
		tm.logger.Info("Loading markov models...")
		var models map[string]markov.ModelInfo
		models, err = tm.markovGen.GetModelInfos(context.Background())
		if err != nil {
			tm.logger.Error("failed to load markov models", "error", err)
			return err
		}

		tm.markovModels = make(map[string]markov.ModelInfo)

		for _, model := range models {
			tm.markovModels[model.Name] = model
		}
		tm.logger.Info("Loaded markov models", "count", len(tm.markovModels))
	}

	return nil
}

// Execute renders a specific template by name, writing the output to the provided io.Writer.
// The `data` argument is passed to the template and can be used to provide context or
// dynamic values.
func (tm *TemplateManager) Execute(w io.Writer, name string, data interface{}) error {
	if name == "" {
		return nil
	}
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.templates.ExecuteTemplate(w, name, data)
}

// GetRandomTemplate returns the name of a randomly selected template from the set
// of loaded full templates. This is the primary mechanism for serving varied and
// unpredictable pages to web scrapers.
func (tm *TemplateManager) GetRandomTemplate() string {
	if len(tm.templateNames) == 0 {
		return ""
	}
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.templateNames[rand.IntN(len(tm.templateNames))]
}

// GetConfig returns a copy of the current configuration.
// This mainly exists for concurrency-safety reasons.
func (tm *TemplateManager) GetConfig() TemplateConfig {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return *tm.config
}

// GetTemplateNames returns a slice of the loaded template names.
// This mainly exists for concurrency-safety reasons, and because
// it returns the names of partial templates as well.
func (tm *TemplateManager) GetTemplateNames() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	var names []string
	for _, t := range tm.templates.Templates() {
		// By default, there is a root template with no name. We don't want to return this in the list
		if strings.Contains(t.Name(), ".html") {
			names = append(names, t.Name())
		}
	}
	return names
}

// GetTemplateDir returns the template dir that the TemplateManager uses.
// This mainly exists for concurrency-safety reasons as well.
func (tm *TemplateManager) GetTemplateDir() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.templateDir
}

// ExecuteTemplateString parses and executes a raw template string using the manager's function map.
// This is ideal for testing or previewing templates without saving them to disk.
func (tm *TemplateManager) ExecuteTemplateString(w io.Writer, content string, data interface{}) error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// Clone the clean, unexecuted template set to avoid race conditions and execution state issues.
	tempSet, err := tm.cleanTemplates.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone clean templates for string execution: %w", err)
	}

	// Parse the user-provided content string into this fresh clone.
	t, err := tempSet.Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse string template: %w", err)
	}

	// Execute the temporary template.
	return t.Execute(w, data)
}
