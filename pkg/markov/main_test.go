package markov

import (
	"context"
	"database/sql"
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a new in-memory SQLite database and a Generator for testing.
// It uses t.Cleanup to ensure resources are released.
func setupTestDB(t *testing.T) (*sql.DB, *Generator) {
	dbFile := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite3", dbFile+"?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=-4000")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := SetupSchema(db); err != nil {
		t.Fatalf("failed to set up schema: %v", err)
	}

	tokenizer := NewDefaultTokenizer()
	g, err := NewGenerator(db, tokenizer)
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}
	t.Cleanup(g.Close)

	return db, g
}

// setupTestDBWithTraining is a convenience helper that also trains a default model.
func setupTestDBWithTraining(t *testing.T) (context.Context, *Generator, ModelInfo) {
	_, g := setupTestDB(t)
	ctx := context.Background()
	modelInfo := ModelInfo{Name: "test_model", Order: 2}

	if err := g.InsertModel(ctx, modelInfo); err != nil {
		t.Fatalf("setup: InsertModel() failed: %v", err)
	}
	modelInfo, err := g.GetModelInfo(ctx, modelInfo.Name)
	if err != nil {
		t.Fatalf("setup: GetModelInfo() failed: %v", err)
	}
	trainingData := "one fish two fish. red fish blue fish."
	if err := g.Train(ctx, modelInfo, strings.NewReader(trainingData)); err != nil {
		t.Fatalf("setup: Train() failed: %v", err)
	}
	return ctx, g, modelInfo
}

// setupTestDBBench creates a database for benchmarking.
func setupTestDBBench(b *testing.B) (*sql.DB, *Generator) {
	dbFile := filepath.Join(b.TempDir(), "bench.db")
	db, err := sql.Open("sqlite3", dbFile+"?_journal_mode=WAL&_synchronous=OFF&_cache_size=-16000&_mmap_size=268435456")
	if err != nil {
		b.Fatalf("failed to open database: %v", err)
	}
	b.Cleanup(func() { _ = db.Close() })

	if err := SetupSchema(db); err != nil {
		b.Fatalf("failed to set up schema: %v", err)
	}

	tokenizer := NewDefaultTokenizer()
	g, err := NewGenerator(db, tokenizer)
	if err != nil {
		b.Fatalf("NewGenerator() error = %v", err)
	}
	b.Cleanup(g.Close)

	return db, g
}

var (
	benchmarkCorpus string
	corpusOnce      sync.Once
)

// createBenchmarkCorpus reads Go source files to create a corpus for benchmarking.
func createBenchmarkCorpus() string {
	corpusOnce.Do(func() {
		var sb strings.Builder
		goRoot := build.Default.GOROOT
		filesToRead := []string{
			filepath.Join(goRoot, "src/net/http/server.go"),
			filepath.Join(goRoot, "src/go/parser/parser.go"),
			filepath.Join(goRoot, "src/encoding/json/encode.go"),
		}

		for _, file := range filesToRead {
			content, err := os.ReadFile(file)
			if err != nil {
				benchmarkCorpus = "this is a fallback corpus for benchmarking. it is not very long but will prevent a crash. "
				return
			}
			sb.Write(content)
			sb.WriteString("\n")
		}
		benchmarkCorpus = sb.String()
	})
	return benchmarkCorpus
}
