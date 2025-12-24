# Sarracenia Markov Library

[![Go Reference](https://pkg.go.dev/badge/github.com/CTAG07/Sarracenia/pkg/markov.svg)](https://pkg.go.dev/github.com/CTAG07/Sarracenia/pkg/markov)
[![Go Version](https://img.shields.io/github/go-mod/go-version/CTAG07/Sarracenia)](https://golang.org)
[![Part of Sarracenia](https://img.shields.io/badge/Part%20of-Sarracenia-8b5cf6)](https://github.com/CTAG07/Sarracenia)
[![AGPLv3 License](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

A high-performance, persistent Markov chain library for Go, backed by SQLite. Designed for production environments requiring reliable text generation, efficient storage of large datasets, and transactional safety.

## Installation

```sh
go get github.com/CTAG07/Sarracenia/pkg/markov
```

## Quick Start

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/CTAG07/Sarracenia/pkg/markov"
	_ "modernc.org/sqlite" // Or github.com/mattn/go-sqlite3
)

func main() {
	// Open a database connection
	db, _ := sql.Open("sqlite", "file:markov.db?_journal_mode=WAL&_busy_timeout=5000")
	defer db.Close()

	// Initialize Schema (Run once)
	if err := markov.SetupSchema(db); err != nil {
		log.Fatal(err)
	}

	// Create Generator
	gen, _ := markov.NewGenerator(db, markov.NewDefaultTokenizer())
	defer gen.Close()

	ctx := context.Background()

	// Define a Model
	model := markov.ModelInfo{Name: "demo", Order: 2}
	_ = gen.InsertModel(ctx, model)

	// Train
	corpus := "The quick brown fox jumps over the lazy dog."
	if err := gen.Train(ctx, model, strings.NewReader(corpus)); err != nil {
		log.Fatal(err)
	}

	// Generate
	text, _ := gen.Generate(ctx, model, markov.WithMaxLength(10))
	fmt.Println(text)
}
```

## Advanced Usage

### Streaming Generation

For real-time applications or large outputs, use `GenerateStream` to receive tokens via a channel as they are generated.

```go
tokenChan, err := generator.GenerateStream(ctx, model, markov.WithMaxLength(50))
if err != nil {
    log.Fatal(err)
}

for token := range tokenChan {
    if token.EOC {
        break
    }
    fmt.Print(token.Text + generator.Tokenizer().Separator())
}
```

### Model Management

Models can be exported to JSON for backup or portability.

```go
file, err := os.Create("model_backup.json")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

err = generator.ExportModel(ctx, model, file)
if err != nil {
    log.Fatal(err)
}
```

## Concurrency Note

**Training Limitation:**
Due to the write-intensive nature of Markov chain training and the single-writer locking model of SQLite, **only one model can be trained at a time**.

While read operations (generation) are concurrent and non-blocking in WAL mode, attempting to run multiple `Train()` jobs simultaneously on the same database file will likely result in `SQLITE_BUSY` (database locked) errors.

## Benchmarks

Benchmarks performed on an Intel Core i9-13905H (Windows 11, Go 1.24.5) using a corpus from the Go standard library.

### Generation Performance

| Benchmark                 | Time/Op | Mem/Op | Allocs/Op |
|:--------------------------|:--------|:-------|:----------|
| `Generate/Simple`         | 6.52 ms | 721 KB | 27,451    |
| `GenerateStream/Simple`   | 6.88 ms | 695 KB | 26,466    |
| `Generate/WithTopK`       | 7.19 ms | 747 KB | 28,499    |
| `GenerateStream/WithTopK` | 7.38 ms | 704 KB | 26,816    |
| `Generate/WithTemp`       | 6.99 ms | 908 KB | 28,928    |

### Training Performance

Note: (Order #) indicates the number of preceding tokens used as context.

| Benchmark         | Time/Op | Processed/Sec | Mem/Op  | Allocs/Op |
|:------------------|:--------|:--------------|:--------|:----------|
| `Train (Order 1)` | 451 ms  | 0.56 MB       | 62.4 MB | 1.74M     |
| `Train (Order 2)` | 654 ms  | 0.43 MB       | 79.9 MB | 2.18M     |
| `Train (Order 3)` | 1.06 s  | 0.39 MB       | 88.9 MB | 2.39M     |
| `Train (Order 4)` | 1.07 s  | 0.37MB        | 91.0 MB | 2.44M     |
| `Train (Order 5)` | 1.14 s  | 0.36MB        | 92.3 MB | 2.46M     |
| `VocabularyPrune` | 2.03 ms | N/A           | 366 KB  | 6,475     |

### Running Benchmarks

```sh
cd pkg/markov
go test -bench . -benchmem
```

## Database Compatibility

This library is optimized for **SQLite**. It utilizes specific SQL dialect features for performance. Porting to PostgreSQL or MySQL would require modifying the prepared statements in `generator.go` and `train.go`.

## License

Licensed under the **AGPLv3**. See the [Sarracenia README](https://github.com/CTAG07/Sarracenia) for alternative licensing options.