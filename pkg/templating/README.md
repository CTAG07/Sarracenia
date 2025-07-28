
# Sarracenia Templating Engine

[![AGPLv3 License](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/CTAG07/Sarracenia)](https://goreportcard.com/report/github.com/CTAG07/Sarracenia)
[![Go Reference](https://pkg.go.dev/badge/github.com/CTAG07/Sarracenia/pkg/templating.svg)](https://pkg.go.dev/github.com/CTAG07/Sarracenia/pkg/templating)

A high-performance, configurable, and extensible Go templating engine designed for generating complex and dynamic web pages.

This library provides a comprehensive toolkit for building varied and intricate HTML content on the fly. It is built on the principle of server-side efficiency, allowing for the generation of sophisticated pages with minimal overhead. It is a ground-up, professional rewrite designed for stability, safety, and extensibility.

## Features

*   **📈 Highly Parameterized Functions:** All generation functions accept parameters (like `count`, `depth`, `complexity`) enabling precise, dynamic control over the generated output directly from within your templates.
*   **🧠 Multi-Model Content Generation:** (Optionally) leverages the `Sarracenia/pkg/markov` library to generate plausible, thematic text from multiple, distinct Markov models stored in a database.
*   **🕸️ Complex Structure Generation:** Includes a suite of functions for creating intricate HTML structures, such as deeply nested DOMs (`nestDivs`) and complex, irregular tables (`randomComplexTable`).
*   **⚙️ Client-Side Content Rendering:** Provides functions to render content via JavaScript, using randomized obfuscation schemes to protect the source content and require client-side script execution.
*   **🛠️ Configurable & Safe:** All computationally intensive functions are governed by safety limits in a `TemplateConfig` struct, preventing templates from accidentally overloading the server or a client's browser.
*   **🚀 Built for Performance:** Engineered for minimal server-side load, ensuring efficient generation even under high concurrency.
*   **🧩 Component-Based Templates:** Fully supports Go's native template composition (`{{template "..."}}`, `{{define "..."}}`), enabling the creation of clean, maintainable, and reusable layouts and partials.

## Template Function Library (Macro Reference)

### Content (`funcs_content.go`)
| Signature                                                                        | Description                                                                                       |
|:---------------------------------------------------------------------------------|:--------------------------------------------------------------------------------------------------|
| `markovSentence modelName maxLength`                                             | Generates a thematic sentence from the specified Markov model.                                    |
| `markovParagraphs modelName count minSentences maxSentences minLength maxLength` | Generates paragraphs of thematic text from the specified Markov model.                            |
| `randomWord`                                                                     | Returns a single random word from the loaded dictionary.                                          |
| `randomSentence length`                                                          | Generates a nonsensical sentence of a given length.                                               |
| `randomParagraphs count minSentences maxSentences minLength maxLength`           | Generates a nonsensical set of paragraphs with lengths in the range of `minLength` to `maxLength` |
| `randomString type length`                                                       | Generates a random string. Types: `username`, `email`, `uuid`, `hex`, `alphanum`.                 |
| `randomDate layout start end`                                                    | Generates a random, formatted date within a range from three strings.                             |
| `randomJSON depth elements len`                                                  | Generates a random, nested JSON object string. Capped by `MaxJSONDepth`.                          |

### Structure (`funcs_structure.go`)
| Signature                                   | Description                                                                            |
|:--------------------------------------------|:---------------------------------------------------------------------------------------|
| `randomForm count styleCount`               | Generates a `<form>` with `count` varied input fields. Capped by `MaxFormFields`.      |
| `randomDefinitionData count sentenceLength` | Returns a slice of `{Term, Def}` structs for building `<dl>` lists.                    |
| `nestDivs depth`                            | Generates `depth` deeply nested `<div>` elements. Capped by `MaxNestDivs`.             |
| `randomComplexTable rows cols`              | Generates an irregular `<table>` with random `colspan`. Capped by `MaxTableRows/Cols`. |

### Styling (`funcs_styling.go`)
| Signature                 | Description                                                                    |
|:--------------------------|:-------------------------------------------------------------------------------|
| `randomColor`             | Returns a random hex color code string (e.g., `#a1f6b3`).                      |
| `randomId prefix length`  | Generates a random HTML ID string with a prefix.                               |
| `randomClasses count`     | Returns a space-separated string of `count` random, utility-style class names. |
| `randomCSSStyle count`    | Returns a string of `count` random CSS property declarations.                  |
| `randomInlineStyle count` | Returns a complete `style="..."` attribute with `count` random properties.     |

### Links & Navigation (`funcs_links.go`)
| Signature                       | Description                                                                  |
|:--------------------------------|:-----------------------------------------------------------------------------|
| `randomLink`                    | Generates a plausible, root-relative URL path, avoiding the `PathWhitelist`. |
| `randomQueryLink keyCount`      | Generates a random path and appends `keyCount` random query parameters.      |

### Logic & Control (`funcs_logic.go`)
| Signature              | Description                                                 |
|:-----------------------|:------------------------------------------------------------|
| `repeat count`         | Returns a slice for use with `range` to loop `count` times. |
| `list item1 item2 ...` | Returns a slice from the provided arguments.                |
| `randomChoice slice`   | Returns a random item from a slice.                         |
| `randomInt min max`    | Returns a random integer in `[min, max)`.                   |

### Simple Math & Logic (`funcs_simple.go`)
| Signature           | Description                                                                     |
|:--------------------|:--------------------------------------------------------------------------------|
| `add a b`           | Returns `a+b`                                                                   |
| `sub a b`           | Returns `a-b`                                                                   |
| `div a b`           | Returns `a/b` (Or 0 if `b=0`)                                                   |
| `mult a b`          | Returns `a*b`                                                                   |
| `mod a b`           | Returns `a%b` (Or 0 if `b=0`)                                                   |
| `max a b`           | Returns the maximum of `a` and `b`                                              |
| `min a b`           | Returns the minimum of `a` and `b`                                              |
| `inc i`             | Increments `i` by one                                                           |
| `dec i`             | Decrements `i` by one                                                           |
| `and arg1 arg2 ...` | Returns `true` if all of the bool args are true                                 |
| `or arg1 arg2 ...`  | Returns `true` if any of the bool args are true                                 |
| `not arg`           | Returns `!arg`                                                                  |
| `isSet value`       | Returns `true` if `value` is not its "zero" value (not `nil`, `""`, `0`, etc.). |

### Computationally Expensive (`funcs_expensive.go`)
| Signature                                 | Description                                                                                                                                             |
|:------------------------------------------|:--------------------------------------------------------------------------------------------------------------------------------------------------------|
| `randomStyleBlock type count`             | Generates a `<style>` block with `count` complex/nested CSS rules. Capped by `MaxStyleRules`.                                                           |
| `randomCSSVars count`                     | Generates a `<style>` block with a chain of interdependent CSS custom properties. Capped by `MaxCssVars`.                                               |
| `randomSVG type complexity`               | Generates a complex inline SVG. `complexity` controls detail. Capped by `MaxSvgElements`. Types: `"fractal"`, `"filters"`                               |
| `jsInteractiveContent tag content cycles` | Generates a JS-powered element that decodes `content` after running a CPU waste loop for `cycles` iterations. Capped by `MaxJsContentSize/WasteCycles`. |

## Installation

```sh
go get github.com/CTAG07/Sarracenia/pkg/templating
```

## Quick Start

Here is a minimal example of setting up and using the `TemplateManager` to render a template.

```go
package main

import (
	"bytes"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/CTAG07/Sarracenia/pkg/templating"
)

func main() {
	// Assumes a logger and a dataDir are already set up.
	var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	var dataDir = "./data" // Directory containing templates/, words.txt, etc.

	// 1. Create a config and the TemplateManager.
	tplConfig := templating.DefaultConfig()
	tplConfig.MarkovEnabled = false // Markov functions are disabled for this example.
	// We pass 'nil' for the markov generator as it's not needed when MarkovEnabled is false.
	tm, err := templating.NewTemplateManager(logger, nil, tplConfig, dataDir)
	if err != nil {
		log.Fatalf("Failed to create template manager: %v", err)
	}

	// 2. Execute a template by name.
	var output bytes.Buffer
	// Assumes a "page.tmpl.html" exists in data/templates. See example below.
	err = tm.Execute(&output, "page.tmpl.html", nil)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	fmt.Println(output.String())
}
```

**Example `data/templates/page.tmpl.html`:**
```html
<h1>{{randomSentence 12}}</h1>
<ul>
    {{range repeat 5}}
    <li>Item: {{randomWord}}</li>
    {{end}}
</ul>
```

## Advanced Usage: Layouts and Partials

The engine fully supports Go's native template composition, which enables the creation of reusable layouts and partials that can receive data.

**Go Application Code:**
```go
// In your main application, assuming 'tm' is an initialized TemplateManager
// with Markov features enabled and a valid generator instance.
pageData := map[string]any{
"PageTitle": "My Dynamic Page",
"Items":     []string{"Alpha", "Beta", "Gamma"},
"TextModel": "tech-docs-model", // Specify which model to use for generation.
}
// Execute the page template, passing in the data map.
err := tm.Execute(&output, "page.tmpl.html", pageData)
// ... handle error
```

**`data/templates/layout.tmpl.html`**
```html
<!DOCTYPE html>
<html>
<head>
    <title>{{.PageTitle}}</title>
</head>
<body>
    {{/* The "." passes the pageData down to the header partial */}}
    {{template "_header.tmpl.html" .}}
    <main>
        {{template "content" .}}
    </main>
</body>
</html>
```

**`data/templates/_header.tmpl.html` (Reusable Partial)**
```html
{{define "_header.tmpl.html"}}
<header>
    <h1>{{.PageTitle}}</h1>
    <nav>...</nav>
</header>
{{end}}
```

**`data/templates/page.tmpl.html` (Content Page)**
```html
{{/* Use the layout as the base, passing all data down */}}
{{template "layout.tmpl.html" .}}

{{/* Define the "content" block that the layout will render */}}
{{define "content"}}
    {{/* This call now uses the specific model passed in the data */}}
    <p>{{markovSentence .TextModel 25}}</p>
    <ul>
        {{range .Items}}
        <li>{{.}}</li>
        {{end}}
    </ul>
{{end}}
```

## Benchmarks

The results below were captured on the following system and provide a performance profile for various content generation categories.

*   **CPU:** 13th Gen Intel(R) Core(TM) i9-13905H
*   **OS:** Windows 11
*   **Go:** 1.24.5

The templates used for the benchmarks are as follows:

```html
{{/* BenchmarkExecute_Simple */}}
<h1>{{randomWord}}</h1><p>{{randomSentence 5}}</p>

{{/* BenchmarkExecute_Styling */}}
<div id="{{randomId "pfx" 8}}" class="{{randomClasses 5}}" style="{{randomInlineStyle 5}}"></div>

{{/* BenchmarkExecute_CPUIntensive */}}
{{randomSVG "fractal" 10}} {{jsInteractiveContent "div" "secret" 1000}}

{{/* BenchmarkExecute_Structure */}}
{{nestDivs 15}} {{randomComplexTable 10 10}}

{{/* BenchmarkExecute_DataGeneration */}}
{{randomForm 10 5}} <script type="application/json">{{randomJSON 4 5 10}}</script>

{{/* BenchmarkExecute_Markov */}}
<h1>{{markovSentence "test_model" 15}}</h1><p>{{markovParagraphs "test_model" 2 3 5 10 20}}</p>
```

### Template Execution Performance

| Benchmark                  | Time/Op   | Mem/Op  | Allocs/Op |
|:---------------------------|:----------|:--------|:----------|
| **Execute/Simple**         | 1.70 µs   | 723 B   | 27        |
| **Execute/Styling**        | 4.49 µs   | 1.9 KB  | 75        |
| **Execute/CPUIntensive**   | 6.09 µs   | 3.0 KB  | 80        |
| **Execute/Structure**      | 21.37 µs  | 17.2 KB | 460       |
| **Execute/DataGeneration** | 77.57 µs  | 51.9 KB | 792       |
| **Execute/Markov**         | 278.84 µs | 70.8 KB | 2,373     |

### How to Run Benchmarks

To run these benchmarks on your own machine, navigate to the package directory and use the following command:

```sh
go test -bench . -benchmem ./pkg/templating
```

## License

This library is part of the Sarracenia project and is licensed under the AGPLv3. See the [Project Readme](https://github.com/CTAG07/Sarracenia/blob/main/README.md) for details on alternative licensing.