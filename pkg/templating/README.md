# Sarracenia Templating Engine

[![Go Reference](https://pkg.go.dev/badge/github.com/amenyxia/Sarracenia/pkg/templating.svg)](https://pkg.go.dev/github.com/amenyxia/Sarracenia/pkg/templating)
[![Go Version](https://img.shields.io/github/go-mod/go-version/amenyxia/Sarracenia)](https://golang.org)
[![Part of Sarracenia](https://img.shields.io/badge/Part%20of-Sarracenia-8b5cf6)](https://github.com/amenyxia/Sarracenia)
[![MIT License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A high-performance, extensible Go templating engine designed for generating complex, dynamic, and obfuscated web
content.

Built for the Sarracenia tarpit, this engine specializes in creating plausible-looking but randomly generated HTML
structures, integrated directly with Markov text generation sources.

## Installation

```sh
go get github.com/amenyxia/Sarracenia/pkg/templating
```

## Quick Start

```go
package main

import (
	"bytes"
	"log/slog"
	"os"

	"github.com/amenyxia/Sarracenia/pkg/templating"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
    // Initialize Config
	cfg := templating.DefaultConfig()
	
    // Create Manager (Pass nil for generator if Markov features aren't needed)
	tm, _ := templating.NewTemplateManager(logger, nil, cfg, "./data")

	// Render a template by name
	var buf bytes.Buffer
	if err := tm.Execute(&buf, "page.tmpl.html", nil); err != nil {
		panic(err)
	}
    
    // Output: <html>...random content...</html>
}
```

## Configuration

The `TemplateConfig` struct controls the behavior and safety limits of the templating engine.

| Key                          | Description                                                                             | Default         |
|:-----------------------------|:----------------------------------------------------------------------------------------|:----------------|
| `markov_enabled`             | Controls whether `markov` functions use the generator. Falls back to `random` if false. | `true`          |
| `markov_separator`           | Separator used by the markov tokenizer.                                                 | `""`            |
| `markov_eoc`                 | End-of-chain marker used by the markov tokenizer.                                       | `""`            |
| `markov_split_regex`         | Regex for splitting tokens in the markov tokenizer.                                     | `""`            |
| `markov_eoc_regex`           | Regex for detecting EOC tokens.                                                         | `""`            |
| `markov_separator_exc_regex` | Regex for tokens that should not have a separator prefix.                               | `""`            |
| `markov_eoc_exc_regex`       | Regex for tokens that should not have an EOC suffix.                                    | `""`            |
| `path_whitelist`             | URL paths considered safe; excluded from random link generation.                        | `[]`            |
| `min_subpaths`               | Minimum number of segments in generated URL paths.                                      | `1`             |
| `max_subpaths`               | Maximum number of segments in generated URL paths.                                      | `5`             |
| `max_json_depth`             | Hard limit on recursion depth for `randomJSON`.                                         | `8`             |
| `max_nest_divs`              | Hard limit on recursion depth for `nestDivs`.                                           | `50`            |
| `max_table_rows`             | Maximum rows for `randomComplexTable`.                                                  | `100`           |
| `max_table_cols`             | Maximum columns for `randomComplexTable`.                                               | `50`            |
| `max_form_fields`            | Maximum fields for `randomForm`.                                                        | `75`            |
| `max_style_rules`            | Maximum complex CSS rules for `randomStyleBlock`.                                       | `200`           |
| `max_css_vars`               | Maximum interdependent CSS variables for `randomCSSVars`.                               | `100`           |
| `max_svg_elements`           | Complexity limit for `randomSVG`.                                                       | `7`             |
| `max_js_content_size`        | Maximum content size (bytes) for `jsInteractiveContent`.                                | `1048576` (1MB) |
| `max_js_waste_cycles`        | Maximum waste loop iterations for `jsInteractiveContent`.                               | `1,000,000`     |

## Template Function Reference

The engine exposes a wide range of custom functions to templates.

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

| Signature                  | Description                                                                  |
|:---------------------------|:-----------------------------------------------------------------------------|
| `randomLink`               | Generates a plausible, root-relative URL path, avoiding the `PathWhitelist`. |
| `randomQueryLink keyCount` | Generates a random path and appends `keyCount` random query parameters.      |

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

| Signature                                 | Description                                                                                                                                     |
|:------------------------------------------|:------------------------------------------------------------------------------------------------------------------------------------------------|
| `randomStyleBlock type count`             | Generates a `<style>` block with `count` complex/nested CSS rules. Capped by `MaxStyleRules`.                                                   |
| `randomCSSVars count`                     | Generates a `<style>` block with a chain of interdependent CSS custom properties. Capped by `MaxCssVars`.                                       |
| `randomSVG type complexity`               | Generates a complex inline SVG. `complexity` controls detail. Capped by `MaxSvgElements`. Types: `"fractal"`, `"filters"`                       |
| `jsInteractiveContent tag content cycles` | Generates a JS element that decodes `content` after running a CPU waste loop for `cycles` iterations. Capped by `MaxJsContentSize/WasteCycles`. |

## Advanced Usage: The Composition Pattern

The engine uses a strict file-naming convention to distinguish between stand-alone pages and reusable components.

* `*.tmpl.html`: **Main Templates.** These are complete pages (e.g., `page.tmpl.html`).
* `*.part.html`: **Partials.** These are reusable components (e.g., `layout.part.html`, `_header.part.html`). They
  should not be rendered directly and should only be used within a complete template.

### Example Architecture

**1. Base Layout (`data/templates/layout.part.html`)**

```html
{{define "layout.part.html"}}
<!DOCTYPE html>
<html>
<body>
    {{template "content" .}}
</body>
</html>
{{end}}
```

**2. Main Page (`data/templates/page.tmpl.html`)**

```html
{{/* Extend the layout */}}
{{template "layout.part.html" .}}

{{/* Define the content block */}}
{{define "content"}}
    <h1>{{markovSentence "tech-model" 10}}</h1>
{{end}}
```

## Benchmarks

The results below were captured on the following system and provide a performance profile for various content generation
categories.

* **CPU:** 13th Gen Intel(R) Core(TM) i9-13905H
* **OS:** Windows 11
* **Go:** 1.24.5

The templates used for the benchmarks are as follows:

```html
{{/* BenchmarkExecute_Simple */}}
<h1>{{randomWord}}</h1><p>{{randomSentence 5}}</p>

{{/* BenchmarkExecute_Styling */}}
<div id="{{randomId " pfx" 8}}" class="{{randomClasses 5}}" style="{{randomInlineStyle 5}}"></div>

{{/* BenchmarkExecute_CPUIntensive */}}
{{randomSVG "fractal" 10}} {{jsInteractiveContent "div" "secret" 1000}}

{{/* BenchmarkExecute_Structure */}}
{{nestDivs 15}} {{randomComplexTable 10 10}}

{{/* BenchmarkExecute_DataGeneration */}}
{{randomForm 10 5}}
<script type="application/json">{{randomJSON 4 5 10}}</script>

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

### Running Benchmarks

```sh
cd pkg/templating
go test -bench . -benchmem
```