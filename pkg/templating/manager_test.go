package templating

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CTAG07/Sarracenia/pkg/markov"
	_ "github.com/mattn/go-sqlite3"
)

// TestMain runs once before all tests in this package. It handles the one-time
// initialization of the global wordList from a temporary file.
func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "sarracenia-global-test-")
	if err != nil {
		log.Fatalf("failed to create temp dir for TestMain: %v", err)
	}

	// Using the correct filename "wordlist.txt" as seen in manager.go
	wordsPath := filepath.Join(tempDir, "wordlist.txt")
	if err := os.WriteFile(wordsPath, []byte("test\nword\nhello\nglobal"), 0644); err != nil {
		log.Fatalf("failed to write global wordlist.txt: %v", err)
	}

	// InitWordList needs the full path to the file.
	if err := InitWordList(wordsPath); err != nil {
		log.Fatalf("failed to init global word list: %v", err)
	}

	exitCode := m.Run()
	_ = os.RemoveAll(tempDir)
	os.Exit(exitCode)
}

// setupTestManager creates a TemplateManager for a single test's scope.
// It relies on the globally initialized wordList and correctly sets up a Markov model.
func setupTestManager(tb testing.TB) *TemplateManager {
	tb.Helper() // Use tb instead of t

	dataDir := tb.TempDir() // Use tb instead of t
	templatesPath := filepath.Join(dataDir, "templates")
	if err := os.Mkdir(templatesPath, 0755); err != nil {
		tb.Fatalf("failed to create templates dir: %v", err)
	}

	dummyWordsPath := filepath.Join(dataDir, "wordlist.txt")
	if err := os.WriteFile(dummyWordsPath, []byte("dummy"), 0644); err != nil {
		tb.Fatalf("failed to write dummy wordlist.txt: %v", err)
	}

	dummyTmplPath := filepath.Join(templatesPath, "dummy.tmpl.html")
	if err := os.WriteFile(dummyTmplPath, []byte(`{{define "dummy.tmpl.html"}}Hello{{end}}`), 0644); err != nil {
		tb.Fatalf("failed to write dummy template: %v", err)
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=memory&cache=shared", tb.Name()))
	if err != nil {
		tb.Fatalf("failed to open in-memory db: %v", err)
	}
	tb.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	if err = markov.SetupSchema(db); err != nil {
		tb.Fatalf("failed to setup markov schema: %v", err)
	}
	tokenizer := markov.NewDefaultTokenizer()
	markovGen, err := markov.NewGenerator(db, tokenizer)
	if err != nil {
		tb.Fatalf("failed to create markov generator: %v", err)
	}
	tb.Cleanup(func() { markovGen.Close() })
	model := markov.ModelInfo{Name: "test_model", Order: 2}
	if err = markovGen.InsertModel(ctx, model); err != nil {
		tb.Fatalf("failed to insert markov model: %v", err)
	}
	model, err = markovGen.GetModelInfo(ctx, model.Name)
	if err != nil {
		tb.Fatalf("failed to get markov model: %v", err)
	}
	trainingData := "one two three four. one two three five."
	if err = markovGen.Train(ctx, model, strings.NewReader(trainingData)); err != nil {
		tb.Fatalf("failed to train markov model: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	config := DefaultConfig()
	config.MarkovEnabled = true
	tm, err := NewTemplateManager(logger, markovGen, config, dataDir)
	if err != nil {
		tb.Fatalf("NewTemplateManager failed: %v", err)
	}
	return tm
}

func TestNewTemplateManager(t *testing.T) {
	if len(wordList) == 0 {
		t.Fatal("global wordList should be initialized by TestMain, but is empty")
	}

	tm := setupTestManager(t)
	if tm == nil {
		t.Fatal("NewTemplateManager returned nil manager")
	}
	if len(tm.templateNames) == 0 {
		t.Error("manager should have loaded at least one template on init")
	}
	if _, ok := tm.markovModels["test_model"]; !ok {
		t.Error("manager should have loaded markov models on init, but 'test_model' is missing")
	}
}

func TestManager_Refresh(t *testing.T) {
	tm := setupTestManager(t)
	initialCount := len(tm.templateNames)

	newTmplPath := filepath.Join(tm.templateDir, "new.tmpl.html")
	if err := os.WriteFile(newTmplPath, []byte(`New Content`), 0644); err != nil {
		t.Fatalf("failed to write new template: %v", err)
	}

	if err := tm.Refresh(); err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	if len(tm.templateNames) != initialCount+1 {
		t.Errorf("expected %d templates after refresh, got %d", initialCount+1, len(tm.templateNames))
	}
}

func TestManager_Execute(t *testing.T) {
	tm := setupTestManager(t)
	var buf bytes.Buffer
	err := tm.Execute(&buf, "dummy.tmpl.html", nil)
	if err != nil {
		t.Fatalf("Execute failed for valid template: %v", err)
	}
	if buf.String() != "Hello" {
		t.Errorf("expected output 'Hello', got '%s'", buf.String())
	}

	err = tm.Execute(&buf, "nonexistent.tmpl.html", nil)
	if err == nil {
		t.Fatal("expected an error for non-existent template, but got nil")
	}
	// html/template returns a specific error message format
	expectedErrString := `html/template: "nonexistent.tmpl.html" is undefined`
	if !strings.Contains(err.Error(), expectedErrString) {
		t.Errorf("error message mismatch: got '%v', expected to contain '%s'", err, expectedErrString)
	}
}

func TestManager_GetRandomTemplate(t *testing.T) {
	tm := setupTestManager(t)
	name := tm.GetRandomTemplate()
	if name != "dummy.tmpl.html" {
		t.Errorf("GetRandomTemplate returned unexpected name '%s'", name)
	}
}

func TestManager_SetConfig(t *testing.T) {
	tm := setupTestManager(t)
	newConfig := DefaultConfig()
	newConfig.MaxNestDivs = 99
	newConfig.PathWhitelist = []string{"/test", "/path"}
	tm.SetConfig(newConfig)

	if tm.config.MaxNestDivs != 99 {
		t.Errorf("SetConfig failed to update MaxNestDivs: expected 99, got %d", tm.config.MaxNestDivs)
	}
	if _, ok := tm.whitelistMap["/test"]; !ok {
		t.Error("SetConfig failed to update whitelistMap")
	}
}

// setupBenchmarkTemplate is a helper to create and load a specific template for a benchmark.
func setupBenchmarkTemplate(b *testing.B, tm *TemplateManager, name, content string) {
	b.Helper()
	templatePath := filepath.Join(tm.templateDir, name)
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		b.Fatalf("failed to write benchmark template %s: %v", name, err)
	}
	if err := tm.Refresh(); err != nil {
		b.Fatalf("failed to refresh after writing template %s: %v", name, err)
	}
}

// BenchmarkExecute_Simple measures the cost of common, low-overhead functions.
func BenchmarkExecute_Simple(b *testing.B) {
	tm := setupTestManager(b)
	content := `<h1>{{randomWord}}</h1><p>{{randomSentence 5}}</p>`
	setupBenchmarkTemplate(b, tm, "simple_funcs.tmpl.html", content)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tm.Execute(io.Discard, "simple_funcs.tmpl.html", nil)
	}
}

// BenchmarkExecute_Styling isolates the performance of generating CSS classes, IDs, and inline styles.
func BenchmarkExecute_Styling(b *testing.B) {
	tm := setupTestManager(b)
	content := `<div id="{{randomId "pfx" 8}}" class="{{randomClasses 5}}" style="{{randomInlineStyle 5}}"></div>`
	setupBenchmarkTemplate(b, tm, "styling.tmpl.html", content)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tm.Execute(io.Discard, "styling.tmpl.html", nil)
	}
}

// BenchmarkExecute_Structure measures the performance of generating large, complex HTML structures.
func BenchmarkExecute_Structure(b *testing.B) {
	tm := setupTestManager(b)
	content := `{{nestDivs 15}} {{randomComplexTable 10 10}}`
	setupBenchmarkTemplate(b, tm, "structure.tmpl.html", content)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tm.Execute(io.Discard, "structure.tmpl.html", nil)
	}
}

// BenchmarkExecute_DataGeneration measures the cost of generating structured data like forms and JSON.
func BenchmarkExecute_DataGeneration(b *testing.B) {
	tm := setupTestManager(b)
	content := `{{randomForm 10 5}} <script type="application/json">{{randomJSON 4 5 10}}</script>`
	setupBenchmarkTemplate(b, tm, "data_gen.tmpl.html", content)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tm.Execute(io.Discard, "data_gen.tmpl.html", nil)
	}
}

// BenchmarkExecute_Markov isolates the performance of querying the database-backed
// Markov generator. This is a critical metric.
func BenchmarkExecute_Markov(b *testing.B) {
	tm := setupTestManager(b)
	content := `<h1>{{markovSentence "test_model" 15}}</h1><p>{{markovParagraphs "test_model" 2 3 5 10 20}}</p>`
	setupBenchmarkTemplate(b, tm, "markov.tmpl.html", content)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tm.Execute(io.Discard, "markov.tmpl.html", nil)
	}
}

// BenchmarkExecute_CPUIntensive measures the server-side cost of the most expensive
// tarpit functions designed to generate complex, computationally-heavy content.
func BenchmarkExecute_CPUIntensive(b *testing.B) {
	tm := setupTestManager(b)
	content := `{{randomSVG "fractal" 10}} {{jsInteractiveContent "div" "secret" 1000}}`
	setupBenchmarkTemplate(b, tm, "cpu_intensive.tmpl.html", content)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tm.Execute(io.Discard, "cpu_intensive.tmpl.html", nil)
	}
}
