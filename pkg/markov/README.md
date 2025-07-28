# Sarracenia Markov Library

[![AGPLv3 License](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/CTAG07/Sarracenia)](https://goreportcard.com/report/github.com/CTAG07/Sarracenia)
[![Go Reference](https://pkg.go.dev/badge/github.com/CTAG07/Sarracenia/pkg/markov.svg)](https://pkg.go.dev/github.com/CTAG07/Sarracenia/pkg/markov)

A robust, performant, and database-backed Markov chain library for Go.

This library provides a comprehensive toolkit for training, managing, and generating text from Markov models. It is designed for high-quality environments, with a focus on performance, data integrity, and a flexible, developer-friendly API.

## Features

*   **🗄️ Database-Backed:** All model data is stored in a SQLite database, allowing for persistence, large datasets, and management of multiple models.
*   **⚡ High-Performance Training:** Utilizes transactions and batch inserts to train models from large text streams efficiently.
*   **🔌 Extensible Tokenizer:** Comes with a sensible default tokenizer, but provides a `Tokenizer` interface to let you define custom rules for splitting text (e.g., for different languages or formats).
*   **🔥 Advanced Generation:** Goes beyond basic random selection with support for `temperature` and `top-K` sampling, giving you fine-grained control over the creativity and coherence of generated text.
*   **🌊 Streaming API:** Generate text token-by-token through a channel, perfect for real-time applications (like chatbots) or generating long sequences without high memory usage.
*   **🛠️ Full Model Lifecycle:** Easily `Insert`, `Remove`, `Export` to JSON, `Import` from JSON, and `Prune` models to manage their entire lifecycle.
*   **🔒 Robust and Safe:** Employs database transactions for all write operations to guarantee data consistency and uses `context.Context` throughout for cancellation and deadline propagation.

## Installation

```sh
go get github.com/CTAG07/Sarracenia/pkg/markov
```

## Quick Start

Here is a complete example of setting up a database, training a model, and generating text.

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/CTAG07/Sarracenia/pkg/markov"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

func main() {
	// For this example, we use an in-memory SQLite DB.
	// For persistence, use a file path: "file:my_markov.db?cache=shared"
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// 1. Setup the database schema. This is only needed once.
	if err := markov.SetupSchema(db); err != nil {
		log.Fatalf("Failed to set up schema: %v", err)
	}

	// 2. Create a tokenizer and a generator.
	tokenizer := markov.NewDefaultTokenizer()
	generator, err := markov.NewGenerator(db, tokenizer)
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}
	defer generator.Close() // Releases prepared statements

	// Optional: Set a logger to see debug output.
	generator.SetLogger(log.New(os.Stdout, "markov: ", log.LstdFlags))

	ctx := context.Background()

	// 3. Define and insert a new model.
	// The order is the number of preceding tokens to consider (2 is a good start).
	model := markov.ModelInfo{
		Name:  "example-model",
		Order: 2,
	}
	if err := generator.InsertModel(ctx, model); err != nil {
		log.Fatalf("Failed to insert model: %v", err)
	}

	// 4. Train the model on some text data.
	trainingData := `A house is not a home. A home is not a house. I am a fish. You are not a fish.`
	err = generator.Train(ctx, model, strings.NewReader(trainingData))
	if err != nil {
		log.Fatalf("Training failed: %v", err)
	}
	fmt.Println("Training complete!")

	// 5. Generate a new sentence!
	generatedText, err := generator.Generate(ctx, model,
		markov.WithMaxLength(20),
		markov.WithTemperature(0.8), // A little less random than default
	)
	if err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	fmt.Printf("Generated Text: %s\n", generatedText)
	// Possible output: Generated Text: You are not a fish.
}
```

## Advanced Usage

### Streaming Generation

For real-time applications, use `GenerateStream` to get a channel of tokens.

```go
tokenChan, err := generator.GenerateStream(ctx, model, markov.WithMaxLength(50))
if err != nil {
    log.Fatal(err)
}

fmt.Print("Streaming: ")
for token := range tokenChan {
    if token.EOC {
        fmt.Print(generator.Tokenizer().EOC())
        break
    }
    fmt.Print(token.Text + generator.Tokenizer().Separator())
}
fmt.Println()
```

### Model Management

Export a model to a JSON file for backup or sharing.

```go
file, err := os.Create("my_model.json")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

err = generator.ExportModel(ctx, model, file)
if err != nil {
    log.Fatal(err)
}
```

## Benchmarks

The results below were captured on the following system:

*   **CPU:** 13th Gen Intel(R) Core(TM) i9-13905H
*   **OS:** Windows 11
*   **Go:** 1.24.5

The models were trained on a corpus containing the following files from the Go standard library source:

*   `src/net/http/server.go`
*   `src/go/parser/parser.go`
*   `src/encoding/json/encode.go`

*Note: Your results may vary based on your hardware.*

### Generation Performance

| Benchmark                          | Time/Op | Mem/Op | Allocs/Op |
|:-----------------------------------|:--------|:-------|:----------|
| **Generate/Simple**                | 6.52 ms | 721 KB | 27,451    |
| **GenerateStream/Simple**          | 6.88 ms | 695 KB | 26,466    |
| **Generate/WithTopK**              | 7.19 ms | 747 KB | 28,499    |
| **GenerateStream/WithTopK**        | 7.38 ms | 704 KB | 26,816    |
| **Generate/WithTemp**              | 6.99 ms | 908 KB | 28,928    |
| **GenerateStream/WithTemp**        | 7.79 ms | 947 KB | 30,143    |
| **Generate/WithTempAndTopK**       | 7.17 ms | 780 KB | 29,696    |
| **GenerateStream/WithTempAndTopK** | 7.88 ms | 782 KB | 29,782    |

### Training & Maintenance Performance

Note: (Order #) means the trained model uses # tokens as context when deciding on the next one.

| Benchmark           | Time/Op | Processed/Sec | Mem/Op  | Allocs/Op |
|:--------------------|:--------|---------------|:--------|:----------|
| **Train (Order 1)** | 451 ms  | 0.56MB        | 62.4 MB | 1,743,442 |
| **Train (Order 2)** | 654 ms  | 0.43MB        | 79.9 MB | 2,188,133 |
| **Train (Order 3)** | 1.06 s  | 0.39MB        | 88.9 MB | 2,394,120 |
| **Train (Order 4)** | 1.07 s  | 0.37MB        | 91.0 MB | 2,446,817 |
| **Train (Order 5)** | 1.14 s  | 0.36MB        | 92.3 MB | 2,464,441 |
| **VocabularyPrune** | 2.03 ms | N/A           | 366 KB  | 6,475     |

### How to Run Benchmarks

To run the benchmarks on your own machine, navigate to the package directory and run the following command:

```sh
cd pkg/markov
go test -bench . -benchmem
```

## Database

This library is designed and optimized for **SQLite**. Because of this, it uses some statements specific to SQLite. While it could be adapted for other SQL databases like PostgreSQL or MySQL, it would require modifying the prepared statements in `generator.go` and `train.go`.

## License

This library is part of the Sarracenia project and is licensed under the AGPLv3. See the [Project Readme](https://github.com/CTAG07/Sarracenia/blob/main/README.md) for details on alternative licensing.