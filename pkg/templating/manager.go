package templating

import (
	"bufio"
	"context"
	"github.com/CTAG07/Sarracenia/pkg/markov"
	"html/template"
	"io"
	"log/slog"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sync"
)

var (
	wordCount     int
	wordList      []string
	loadWordsOnce sync.Once
	loadErr       error
)

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

type TemplateManager struct {
	logger        *slog.Logger
	config        TemplateConfig
	whitelistMap  map[string]struct{}
	markovGen     *markov.Generator
	markovModels  map[string]markov.ModelInfo
	templates     *template.Template
	templateNames []string
	funcMap       template.FuncMap
	templateDir   string
	mu            sync.RWMutex
}

func NewTemplateManager(logger *slog.Logger, markovGen *markov.Generator, config TemplateConfig, dataDir string) (*TemplateManager, error) {

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

func (tm *TemplateManager) SetConfig(config TemplateConfig) {
	tm.config = config
	for _, path := range config.PathWhitelist {
		tm.whitelistMap[path] = struct{}{}
	}
}

func (tm *TemplateManager) Refresh() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	filePattern := filepath.Join(tm.templateDir, "*.tmpl.html")
	tm.logger.Info("Loading template files...")

	parsedFiles, err := template.New("").Funcs(tm.funcMap).ParseGlob(filePattern)
	if err != nil {
		tm.logger.Error("failed to parse template files", "error", err)
		return err
	}

	var names []string
	for _, t := range parsedFiles.Templates() {
		// By default, there is a root template with no name. We don't want to execute this
		if t.Name() != "" && t.Name() != parsedFiles.Name() {
			names = append(names, t.Name())
		}
	}

	if len(names) == 0 {
		tm.logger.Warn("No template files found matching pattern", "pattern", filePattern)
	}

	tm.templates = parsedFiles
	tm.templateNames = names
	tm.logger.Info("Loaded template files", "count", len(parsedFiles.Templates()))

	if tm.config.MarkovEnabled {
		tm.logger.Info("Loading markov models...")
		models, err := tm.markovGen.GetModelInfos(context.Background())
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

func (tm *TemplateManager) Execute(w io.Writer, name string, data interface{}) error {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.templates.ExecuteTemplate(w, name, data)
}

func (tm *TemplateManager) GetRandomTemplate() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.templateNames[rand.IntN(len(tm.templateNames))]
}
