package templating

import (
	"fmt"
	"html/template"
	"math/rand/v2"
	"strconv"
	"strings"
)

var (
	// cssPropertyGenerators holds a map of CSS properties to functions that generate plausible values.
	cssPropertyGenerators map[string]func() string

	// cssPropertyKeys holds a slice of the map keys for fast random selection.
	cssPropertyKeys []string

	// classPrefixes is a pre-defined list for generating random utility-style class names.
	classPrefixes = []string{"p", "m", "w", "h", "bg", "text", "border", "flex", "grid"}
)

// See https://www.w3schools.com/CSSref/index.php for the list of CSS properties I used
func init() {
	// Helper functions
	randomLength := func(min, max int, units ...string) string {
		if len(units) == 0 {
			units = []string{"px", "em", "%", "rem", "vh", "vw"}
		}
		return strconv.Itoa(rand.IntN(max-min+1)+min) + units[rand.IntN(len(units))]
	}
	randomShorthand := func(min, max int) string {
		count := rand.IntN(4) + 1
		values := make([]string, count)
		for i := range values {
			values[i] = randomLength(min, max, "px", "%", "em")
		}
		return strings.Join(values, " ")
	}
	randomFloat := func(min, max float64) string {
		return fmt.Sprintf("%.2f", min+rand.Float64()*(max-min))
	}
	randomAngle := func() string {
		return randomKeyword([]string{strconv.Itoa(rand.IntN(361)) + "deg", randomFloat(0, 6.28) + "rad"})
	}
	randomBorderStyle := func() string {
		return randomKeyword([]string{"solid", "dotted", "dashed", "double", "groove", "ridge", "inset", "outset"})
	}
	randomBorderWidth := func() string { return randomLength(1, 12, "px") }
	randomTime := func() string {
		return fmt.Sprintf("%.2fs", rand.Float64()*2)
	}

	// This must be the definition of insanity
	cssPropertyGenerators = map[string]func() string{
		// A
		"accent-color": randomColor,
		"align-content": func() string {
			return randomKeyword([]string{"flex-start", "flex-end", "center", "space-between", "space-around", "stretch"})
		},
		"align-items": func() string {
			return randomKeyword([]string{"stretch", "flex-start", "flex-end", "center", "baseline"})
		},
		"align-self": func() string {
			return randomKeyword([]string{"auto", "flex-start", "flex-end", "center", "baseline", "stretch"})
		},
		"animation": func() string {
			return fmt.Sprintf("%s %s %s %s", randomHexString(8), randomFloat(0.2, 2)+"s", randomKeyword([]string{"ease-in-out", "linear"}), randomKeyword([]string{"infinite", "1", "3"}))
		},
		"aspect-ratio": func() string { return randomKeyword([]string{"auto", "1 / 1", "16 / 9", "4 / 3"}) },

		// B
		"backdrop-filter": func() string {
			return fmt.Sprintf("blur(%s) contrast(%s)", randomLength(0, 10, "px"), randomFloat(0.5, 2.0))
		},
		"backface-visibility":   func() string { return randomKeyword([]string{"visible", "hidden"}) },
		"background-attachment": func() string { return randomKeyword([]string{"scroll", "fixed", "local"}) },
		"background-blend-mode": func() string {
			return randomKeyword([]string{"normal", "multiply", "screen", "overlay", "darken", "lighten", "color-dodge"})
		},
		"background-clip":  func() string { return randomKeyword([]string{"border-box", "padding-box", "content-box", "text"}) },
		"background-color": randomColor,
		"background-image": func() string {
			return fmt.Sprintf("linear-gradient(%s, %s, %s)", randomAngle(), randomColor(), randomColor())
		},
		"background-origin":   func() string { return randomKeyword([]string{"padding-box", "border-box", "content-box"}) },
		"background-position": func() string { return fmt.Sprintf("%s %s", randomLength(0, 100, "%"), randomLength(0, 100, "%")) },
		"background-repeat": func() string {
			return randomKeyword([]string{"repeat", "no-repeat", "repeat-x", "repeat-y", "space", "round"})
		},
		"background-size":     func() string { return randomKeyword([]string{"auto", "cover", "contain"}) },
		"border":              func() string { return fmt.Sprintf("%s %s %s", randomBorderWidth(), randomBorderStyle(), randomColor()) },
		"border-top":          func() string { return fmt.Sprintf("%s %s %s", randomBorderWidth(), randomBorderStyle(), randomColor()) },
		"border-right":        func() string { return fmt.Sprintf("%s %s %s", randomBorderWidth(), randomBorderStyle(), randomColor()) },
		"border-bottom":       func() string { return fmt.Sprintf("%s %s %s", randomBorderWidth(), randomBorderStyle(), randomColor()) },
		"border-left":         func() string { return fmt.Sprintf("%s %s %s", randomBorderWidth(), randomBorderStyle(), randomColor()) },
		"border-top-color":    randomColor,
		"border-right-color":  randomColor,
		"border-bottom-color": randomColor,
		"border-left-color":   randomColor,
		"border-top-style":    randomBorderStyle,
		"border-right-style":  randomBorderStyle,
		"border-bottom-style": randomBorderStyle,
		"border-left-style":   randomBorderStyle,
		"border-top-width":    randomBorderWidth,
		"border-right-width":  randomBorderWidth,
		"border-bottom-width": randomBorderWidth,
		"border-left-width":   randomBorderWidth,
		"border-collapse":     func() string { return randomKeyword([]string{"separate", "collapse"}) },
		"border-spacing":      func() string { return randomLength(0, 15, "px") },
		"border-radius":       func() string { return randomLength(0, 50, "%", "px") },
		"border-inline":       func() string { return fmt.Sprintf("%s %s %s", randomBorderWidth(), randomBorderStyle(), randomColor()) },
		"border-block":        func() string { return fmt.Sprintf("%s %s %s", randomBorderWidth(), randomBorderStyle(), randomColor()) },
		"bottom":              func() string { return randomKeyword([]string{"auto", randomLength(-50, 150, "px", "%")}) },
		"box-shadow": func() string {
			return fmt.Sprintf("%s %s %s %s %s %s", randomLength(-8, 8, "px"), randomLength(-8, 8, "px"), randomLength(0, 20, "px"), randomLength(0, 12, "px"), randomColor(), randomKeyword([]string{"", "inset"}))
		},
		"box-sizing": func() string { return randomKeyword([]string{"content-box", "border-box"}) },

		// C
		"caret-color": randomColor,
		"clear":       func() string { return randomKeyword([]string{"none", "left", "right", "both"}) },
		"clip-path": func() string {
			return randomKeyword([]string{"circle(50%)", "ellipse(25% 40%)", "inset(10% 20% 30% 10%)", "polygon(50% 0%, 100% 50%, 50% 100%, 0% 50%)"})
		},
		"color":        randomColor,
		"column-count": func() string { return strconv.Itoa(rand.IntN(4) + 1) },
		"column-rule": func() string {
			return fmt.Sprintf("%s %s %s", randomLength(1, 10, "px"), randomKeyword([]string{"solid", "dotted", "dashed"}), randomColor())
		},
		"column-span": func() string { return randomKeyword([]string{"none", "all"}) },
		"contain": func() string {
			return randomKeyword([]string{"none", "strict", "content", "size", "layout", "style", "paint"})
		},
		"content": func() string { return randomKeyword([]string{"normal", "none", "' '", "open-quote"}) },
		"cursor": func() string {
			return randomKeyword([]string{"pointer", "default", "wait", "text", "move", "help", "not-allowed", "crosshair", "zoom-in"})
		},

		// D
		"display": func() string {
			return randomKeyword([]string{"block", "inline", "inline-block", "flex", "grid", "inline-flex", "table", "none"})
		},

		// F
		"filter": func() string {
			return fmt.Sprintf("blur(%dpx) brightness(%.1f) contrast(%d%%) saturate(%d%%) hue-rotate(%s)", rand.IntN(10), rand.Float64()*1.5+0.5, rand.IntN(151)+50, rand.IntN(201), randomAngle())
		},
		"flex-basis":     func() string { return randomKeyword([]string{"auto", "content", randomLength(10, 50, "%", "px")}) },
		"flex-direction": func() string { return randomKeyword([]string{"row", "row-reverse", "column", "column-reverse"}) },
		"flex-grow":      func() string { return strconv.Itoa(rand.IntN(5)) },
		"flex-shrink":    func() string { return strconv.Itoa(rand.IntN(5)) },
		"flex-wrap":      func() string { return randomKeyword([]string{"nowrap", "wrap", "wrap-reverse"}) },
		"float":          func() string { return randomKeyword([]string{"none", "left", "right"}) },
		"font-family": func() string {
			return randomKeyword([]string{`"Arial", sans-serif`, `"Georgia", serif`, `"Courier New", monospace`})
		},
		"font-size":    func() string { return randomLength(12, 48, "px", "em", "rem") },
		"font-style":   func() string { return randomKeyword([]string{"normal", "italic", "oblique"}) },
		"font-variant": func() string { return randomKeyword([]string{"normal", "small-caps"}) },
		"font-weight":  func() string { return randomKeyword([]string{"normal", "bold", "100", "400", "700", "900"}) },

		// G
		"gap":                   func() string { return randomShorthand(0, 40) },
		"grid-auto-flow":        func() string { return randomKeyword([]string{"row", "column", "dense", "row dense"}) },
		"grid-template-columns": func() string { return fmt.Sprintf("repeat(%d, 1fr)", rand.IntN(5)+1) },

		// I
		"image-rendering": func() string { return randomKeyword([]string{"auto", "crisp-edges", "pixelated"}) },
		"isolation":       func() string { return randomKeyword([]string{"auto", "isolate"}) },

		// J
		"justify-content": func() string {
			return randomKeyword([]string{"flex-start", "flex-end", "center", "space-between", "space-around", "space-evenly"})
		},
		"justify-items": func() string { return randomKeyword([]string{"start", "end", "center", "stretch"}) },

		// L
		"left":           func() string { return randomKeyword([]string{"auto", randomLength(-50, 150, "px", "%")}) },
		"letter-spacing": func() string { return randomLength(-2, 10, "px", "em") },
		"line-height":    func() string { return randomKeyword([]string{"normal", randomFloat(1, 2.5)}) },
		"list-style":     func() string { return randomKeyword([]string{"disc", "circle", "square", "decimal", "none", "inside"}) },

		// M
		"margin":        func() string { return randomShorthand(0, 100) },
		"margin-inline": func() string { return randomShorthand(0, 100) },
		"margin-block":  func() string { return randomShorthand(0, 100) },
		"max-height":    func() string { return randomKeyword([]string{"none", randomLength(100, 500, "px", "vh")}) },
		"max-width":     func() string { return randomKeyword([]string{"none", randomLength(200, 1000, "px", "vw")}) },
		"mix-blend-mode": func() string {
			return randomKeyword([]string{"normal", "multiply", "screen", "overlay", "darken", "lighten"})
		},

		// O
		"object-fit":      func() string { return randomKeyword([]string{"fill", "contain", "cover", "none", "scale-down"}) },
		"object-position": func() string { return fmt.Sprintf("%s %s", randomLength(0, 100, "%"), randomLength(0, 100, "%")) },
		"opacity":         func() string { return randomFloat(0.1, 1.0) },
		"order":           func() string { return strconv.Itoa(rand.IntN(11) - 5) },
		"outline": func() string {
			return fmt.Sprintf("%s %s %s", randomLength(1, 8, "px"), randomKeyword([]string{"solid", "dotted", "dashed"}), randomColor())
		},
		"outline-offset":      func() string { return randomLength(0, 15, "px") },
		"overflow":            func() string { return randomKeyword([]string{"visible", "hidden", "scroll", "auto"}) },
		"overflow-wrap":       func() string { return randomKeyword([]string{"normal", "break-word"}) },
		"overscroll-behavior": func() string { return randomKeyword([]string{"auto", "contain", "none"}) },

		// P
		"padding":        func() string { return randomShorthand(0, 100) },
		"padding-inline": func() string { return randomShorthand(0, 100) },
		"padding-block":  func() string { return randomShorthand(0, 100) },
		"perspective":    func() string { return randomLength(500, 2000, "px") },
		"pointer-events": func() string { return randomKeyword([]string{"auto", "none"}) },
		"position":       func() string { return randomKeyword([]string{"static", "relative", "absolute", "fixed", "sticky"}) },

		// R
		"resize": func() string { return randomKeyword([]string{"none", "both", "horizontal", "vertical"}) },
		"right":  func() string { return randomKeyword([]string{"auto", randomLength(-50, 150, "px", "%")}) },

		// S
		"scroll-behavior":   func() string { return randomKeyword([]string{"auto", "smooth"}) },
		"scroll-margin":     func() string { return randomShorthand(0, 20) },
		"scroll-padding":    func() string { return randomShorthand(0, 20) },
		"scroll-snap-align": func() string { return randomKeyword([]string{"none", "start", "end", "center"}) },
		"scroll-snap-type":  func() string { return randomKeyword([]string{"none", "x mandatory", "y proximity", "block mandatory"}) },

		// T
		"table-layout": func() string { return randomKeyword([]string{"auto", "fixed"}) },
		"text-align":   func() string { return randomKeyword([]string{"left", "right", "center", "justify"}) },
		"text-decoration": func() string {
			return fmt.Sprintf("%s %s %s", randomKeyword([]string{"none", "underline", "overline", "line-through"}), randomKeyword([]string{"solid", "wavy", "dotted"}), randomColor())
		},
		"text-overflow": func() string { return randomKeyword([]string{"clip", "ellipsis"}) },
		"text-shadow": func() string {
			return fmt.Sprintf("%s %s %s %s", randomLength(-5, 5, "px"), randomLength(-5, 5, "px"), randomLength(0, 10, "px"), randomColor())
		},
		"text-transform": func() string { return randomKeyword([]string{"none", "capitalize", "uppercase", "lowercase"}) },
		"top":            func() string { return randomKeyword([]string{"auto", randomLength(-50, 150, "px", "%")}) },
		"transform": func() string {
			return fmt.Sprintf("rotate(%s) scale(%.2f) skewX(%s) translateX(%s)", randomAngle(), rand.Float64()*1.5+0.5, randomAngle(), randomLength(-50, 50, "px"))
		},
		"transform-origin": func() string {
			return fmt.Sprintf("%s %s", randomKeyword([]string{"center", "top", "left"}), randomKeyword([]string{"", "bottom", "right"}))
		},
		"transition": func() string {
			return fmt.Sprintf("%s %s %s %s", randomKeyword([]string{"all", "opacity", "transform", "color"}), randomTime(), randomKeyword([]string{"ease", "ease-in-out", "linear"}), randomTime())
		},

		// U
		"user-select": func() string { return randomKeyword([]string{"auto", "none", "text", "all"}) },

		// V
		"vertical-align": func() string {
			return randomKeyword([]string{"baseline", "sub", "super", "top", "middle", "bottom", randomLength(-20, 20, "px")})
		},
		"visibility": func() string { return randomKeyword([]string{"visible", "hidden", "collapse"}) },

		// W
		"white-space": func() string { return randomKeyword([]string{"normal", "nowrap", "pre", "pre-wrap", "pre-line"}) },
		"will-change": func() string {
			return randomKeyword([]string{"auto", "scroll-position", "contents", "transform", "opacity"})
		},
		"word-break":   func() string { return randomKeyword([]string{"normal", "break-all", "keep-all"}) },
		"word-spacing": func() string { return randomLength(-2, 20, "px", "em") },
		"writing-mode": func() string { return randomKeyword([]string{"horizontal-tb", "vertical-rl", "vertical-lr"}) },

		// Z
		"z-index": func() string { return strconv.Itoa(rand.IntN(2000) - 1000) },
	}

	// Pre-cache the keys for performance.
	cssPropertyKeys = make([]string, 0, len(cssPropertyGenerators))
	for k := range cssPropertyGenerators {
		cssPropertyKeys = append(cssPropertyKeys, k)
	}
}

func randomHexString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = lowerHexChars[rand.IntN(len(lowerHexChars))]
	}
	return string(b)
}

// randomKeyword randomly chooses a string from a list of strings, and is used as a helper here and in the expensive funcs
func randomKeyword(keywords []string) string {
	return keywords[rand.IntN(len(keywords))]
}

// randomColor generates a random hexadecimal color code string (e.g., "#a1f6b3").
func randomColor() string {
	return fmt.Sprintf("#%06x", rand.IntN(0xFFFFFF+1))
}

// randomId generates a plausible-looking, random HTML ID string with a given prefix.
func randomId(prefix string, length int) string {
	return prefix + "-" + randomHexString(length)
}

// randomClasses generates a space-separated string of random, utility-style CSS class names.
func randomClasses(count int) template.CSS {
	var builder strings.Builder
	for i := 0; i < count; i++ {
		prefix := classPrefixes[rand.IntN(len(classPrefixes))]
		builder.WriteString(prefix)
		builder.WriteByte('-')
		builder.WriteString(randomHexString(4))
		if i < count-1 {
			builder.WriteByte(' ')
		}
	}
	return template.CSS(builder.String())
}

// randomCSSStyle generates a string of `count` random CSS property declarations.
func randomCSSStyle(count int) template.CSS {
	var builder strings.Builder
	for i := 0; i < count; i++ {
		key := cssPropertyKeys[rand.IntN(len(cssPropertyKeys))]
		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(cssPropertyGenerators[key]())
		builder.WriteString("; ")
	}
	return template.CSS(builder.String())
}

// randomInlineStyle generates a complete HTML style attribute string.
func randomInlineStyle(count int) template.HTMLAttr {
	if count <= 0 {
		return ""
	}
	return template.HTMLAttr(`style="` + randomCSSStyle(count) + `"`)
}
