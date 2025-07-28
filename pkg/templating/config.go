package templating

// TemplateConfig holds all configuration options for the templating engine.
type TemplateConfig struct {
	// MarkovEnabled controls whether markov models will be used for content generation or not
	MarkovEnabled bool

	// PathWhitelist is a list of URL paths that are considered safe and should not
	// be used for randomly generated links, to avoid collisions with real endpoints.
	PathWhitelist []string

	// MinSubpaths defines the minimum number of segments in a generated URL path.
	MinSubpaths int

	// MaxSubpaths defines the maximum number of segments in a generated URL path.
	MaxSubpaths int

	// MaxJSONDepth sets a hard upper limit on the recursion depth for randomJSON.
	MaxJSONDepth int

	// MaxNestDivs sets a hard upper limit on the depth for the nestDivs function.
	// This prevents templates from requesting a depth that could crash a browser.
	MaxNestDivs int

	// MaxTableRows sets the maximum number of rows for randomComplexTable.
	MaxTableRows int

	// MaxTableCols sets the maximum number of columns for randomComplexTable.
	MaxTableCols int

	// MaxFormFields sets the maximum number of fields for randomForm.
	MaxFormFields int

	// MaxStyleRules sets the maximum number of complex CSS rules for randomStyleBlock.
	MaxStyleRules int

	// MaxCssVars sets the maximum number of interdependent CSS custom properties for randomCSSVars.
	MaxCssVars int

	// MaxSvgElements sets a general limit for SVG complexity (e.g., recursion depth).
	MaxSvgElements int

	// MaxJsContentSize sets the maximum size of the content encoded in jsInteractiveContent.
	MaxJsContentSize int

	// MaxJsWasteCycles sets the maximum number of iterations for the CPU waste loop.
	MaxJsWasteCycles int
}

// DefaultConfig returns a TemplateConfig with safe default values.
// PathWhiteList is empty by default, and as such the default config assumes
// that every path will lead to the tarpit.
func DefaultConfig() TemplateConfig {
	return TemplateConfig{
		MarkovEnabled:    false,
		PathWhitelist:    []string{},
		MinSubpaths:      1,
		MaxSubpaths:      5,
		MaxJSONDepth:     8,
		MaxNestDivs:      50,
		MaxTableRows:     100,
		MaxTableCols:     50,
		MaxFormFields:    75,
		MaxStyleRules:    200,
		MaxCssVars:       100,
		MaxSvgElements:   7,
		MaxJsContentSize: 1048576, // 1MB
		MaxJsWasteCycles: 1_000_000,
	}
}
