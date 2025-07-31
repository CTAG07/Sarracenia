package templating

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
)

var (
	jsUnaryOps = []string{
		"Math.sin(%s)", "Math.cos(%s)", "Math.tan(%s)", "Math.sqrt(Math.abs(%s))", "Math.log(Math.abs(%s)+1)",
		"Math.exp(%s)", "Math.sinh(%s)", "Math.cosh(%s)", "Math.tanh(%s)", "Math.floor(%s)", "Math.ceil(%s)",
		"Math.round(%s)", "Math.sign(%s)",
	}
	jsBinaryOps = []string{
		"(%s + %s)", "(%s - %s)", "(%s * %s)", "(%s / (%s||1))", "(%s %% (%s||1))",
		"Math.pow(Math.abs(%s), Math.abs(%s))", "Math.max(%s, %s)", "Math.min(%s, %s)",
		"Math.hypot(%s, %s)",
	}
	jsVars = []string{"i", "j", "w", "i+j", "i*j", "i/(j+1)", "j+1", "i%%10", "j%%10", "(w%%1000)"}
)

// StyleBlock is a struct returned by the randomStyleBlock template function.
// It allows the template to access both the generated <style> block HTML and the
// unique parent class name required to apply those styles to a specific element.
type StyleBlock struct {
	// Style contains the full <style>...</style> block as safe HTML.
	Style template.HTML
	// Class contains the unique CSS class name (e.g., "c-a1b2c3d4") to be used
	// in an element's class attribute.
	Class template.CSS
}

// randomStyleBlock generates a <style> block with a specified number of complex CSS rules
func (tm *TemplateManager) randomStyleBlock(styleType string, requestedCount int) StyleBlock {
	count := min(requestedCount, tm.config.MaxStyleRules)
	parentClass := "c-" + randomHexString(12)
	var builder strings.Builder

	builder.WriteString("<style>\n")
	for i := 0; i < count; i++ {
		selector := buildComplexSelector("."+parentClass, styleType)
		// Use the pure randomCSSStyle function for the rule body.
		ruleBody := randomCSSStyle(rand.IntN(4) + 2)
		_, err := fmt.Fprintf(&builder, "%s { %s }\n", selector, ruleBody)
		if err != nil {
			return StyleBlock{}
		}
	}
	builder.WriteString("</style>")

	return StyleBlock{
		Style: template.HTML(builder.String()),
		Class: template.CSS(parentClass),
	}
}

// buildComplexSelector is an unexported helper for creating convoluted CSS selectors.
func buildComplexSelector(base, styleType string) string {
	switch styleType {
	case "nested":
		depth := rand.IntN(8) + 4
		selector := base
		for i := 0; i < depth; i++ {
			selector += " > " + randomKeyword([]string{"div", "span", "p", "a"})
			if rand.IntN(2) == 0 {
				selector += fmt.Sprintf(":nth-child(%d)", rand.IntN(10)+1)
			}
		}
		return selector
	case "complex":
		attr := fmt.Sprintf(`[data-%s^="%s"]`, randomWord(), randomHexString(3))
		pseudo := ":" + randomKeyword([]string{"hover", "active", "focus", "not(:last-child)", "first-of-type"})
		combinator := " " + randomKeyword([]string{"+", "~", ">"}) + " " + randomKeyword([]string{"span", "b", "i"})
		return base + attr + pseudo + combinator
	default: // "utility"
		return base + " ." + randomHexString(8)
	}
}

// randomCSSVars generates a <style> block defining a chain of interdependent CSS custom properties
func (tm *TemplateManager) randomCSSVars(requestedCount int) template.HTML {
	count := min(requestedCount, tm.config.MaxCssVars)
	if count < 2 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("<style>:root {\n")
	_, err := fmt.Fprintf(&builder, "  --v1: %dpx;\n", rand.IntN(100)+1)
	if err != nil {
		return ""
	}
	for i := 2; i <= count; i++ {
		op := randomKeyword([]string{"+", "-", "*"})
		refVar := fmt.Sprintf("var(--v%d)", rand.IntN(i-1)+1)
		val := randomInt(1, 50)
		expr := fmt.Sprintf("calc(%s %s %dpx)", refVar, op, val)
		_, err = fmt.Fprintf(&builder, "  --v%d: %s;\n", i, expr)
		if err != nil {
			return ""
		}
	}
	builder.WriteString("}\n</style>")
	return template.HTML(builder.String())
}

// randomSVG generates complex, computationally expensive inline SVG graphics.
// Types: "fractal", "filters". These stress the client's rendering engine.
func (tm *TemplateManager) randomSVG(svgType string, depth int) template.HTML {
	switch svgType {
	case "fractal":
		return tm.generateFractalSVG(depth)
	case "filters":
		return tm.generateFiltersSVG(depth)
	default:
		return ""
	}
}

// generateFractalSVG creates a recursive fractal tree, a classic CPU-intensive render task.
func (tm *TemplateManager) generateFractalSVG(depth int) template.HTML {
	var pathBuilder strings.Builder
	var drawBranch func(x, y, angle, length float64, depth int)

	drawBranch = func(x, y, angle, length float64, depth int) {
		if depth <= 0 {
			return
		}
		rad := angle * (math.Pi / 180.0)
		x2, y2 := x+length*math.Cos(rad), y+length*math.Sin(rad)
		_, err := fmt.Fprintf(&pathBuilder, "M%.2f,%.2f L%.2f,%.2f ", x, y, x2, y2)
		if err != nil {
			return
		}

		for i := 0; i < rand.IntN(2)+2; i++ {
			drawBranch(x2, y2, angle+(rand.Float64()*80-40), length*(rand.Float64()*0.2+0.7), depth-1)
		}
	}

	// Total elements are capped by config for safety.
	maxElements := tm.config.MaxSvgElements
	if int(math.Pow(3, float64(depth))) > maxElements {
		depth = int(math.Log(float64(maxElements)) / math.Log(3))
	}

	drawBranch(250, 500, -90, float64(randomInt(70, 90)), depth)

	svg := fmt.Sprintf(`<svg width="500" height="500" viewBox="0 0 500 500"><path d="%s" stroke="%s" stroke-width="1" fill="none"/></svg>`, pathBuilder.String(), randomColor())
	return template.HTML(svg)
}

// generateFiltersSVG's complexity now scales with aggression.
func (tm *TemplateManager) generateFiltersSVG(depth int) template.HTML {
	filterID := "f-" + randomHexString(8)

	filterGenerators := []func() string{
		func() string { return fmt.Sprintf(`<feGaussianBlur stdDeviation="%d"/>`, rand.IntN(5)+1) },
		func() string {
			return fmt.Sprintf(`<feMorphology operator="%s" radius="%d"/>`, randomKeyword([]string{"erode", "dilate"}), rand.IntN(4)+1)
		},
		func() string {
			return fmt.Sprintf(`<feTurbulence type="fractalNoise" baseFrequency="0.0%d" numOctaves="%d" result="t"/>`, rand.IntN(8)+1, rand.IntN(3)+2)
		},
		func() string { return `<feDisplacementMap in="SourceGraphic" in2="t" scale="50"/>` },
		func() string { return fmt.Sprintf(`<feColorMatrix type="hueRotate" values="%d"/>`, rand.IntN(361)) },
		func() string {
			return fmt.Sprintf(`<feConvolveMatrix order="3" kernelMatrix="1 -1 1 -1 %d -1 1 -1 1"/>`, rand.IntN(10)-5)
		},
	}

	var filterChain strings.Builder
	depth = min(depth, len(filterGenerators))

	rand.Shuffle(len(filterGenerators), func(i, j int) { filterGenerators[i], filterGenerators[j] = filterGenerators[j], filterGenerators[i] })

	for i := 0; i < depth; i++ {
		filterChain.WriteString(filterGenerators[i]())
	}

	svg := fmt.Sprintf(`<svg width="200" height="200"><defs><filter id="%s" x="-50%%" y="-50%%" width="200%%" height="200%%">%s</filter></defs><rect width="100%%" height="100%%" fill="%s" filter="url(#%s)"/></svg>`,
		filterID, filterChain.String(), randomColor(), filterID)
	return template.HTML(svg)
}

// jsInteractiveContent is the master function for JS-based deception.
// The `wasteCycles` parameter directly controls the number of iterations in the waste loop.
func (tm *TemplateManager) jsInteractiveContent(tag, content string, wasteCycles int) template.HTML {
	if len(content) > tm.config.MaxJsContentSize {
		content = content[:tm.config.MaxJsContentSize]
	}

	placeholderID := "p-" + randomHexString(12)

	var encodedContent, decoderJS string
	xorKey := byte(rand.IntN(254) + 1)

	switch rand.IntN(6) {
	case 0: // Reversed Base64
		encodedContent = reverseString(base64.StdEncoding.EncodeToString([]byte(content)))
		decoderJS = "d=atob(c.split('').reverse().join(''));"
	case 1: // Hex
		encodedContent = fmt.Sprintf("%x", content)
		decoderJS = `d='';for(let i=0;i<c.length;i+=2){d+=String.fromCharCode(parseInt(c.substr(i,2),16));}`
	case 2: // XOR + Base64
		encodedContent = base64.StdEncoding.EncodeToString(xorBytes([]byte(content), xorKey))
		decoderJS = fmt.Sprintf("b=atob(c);d='';for(let i=0;i<b.length;i++){d+=String.fromCharCode(b.charCodeAt(i)^%d);}", xorKey)
	case 3: // Character Code Array
		var codes []string
		for _, char := range []byte(content) {
			codes = append(codes, strconv.Itoa(int(char)))
		}
		encodedContent = strings.Join(codes, ",")
		decoderJS = "d=String.fromCharCode.apply(null,c.split(','));"
	case 4: // Reversed Hex
		encodedContent = reverseString(fmt.Sprintf("%x", content))
		decoderJS = `h=c.split('').reverse().join('');d='';for(let i=0;i<h.length;i+=2){d+=String.fromCharCode(parseInt(h.substr(i,2),16));}`
	default: // Plain Base64
		encodedContent = base64.StdEncoding.EncodeToString([]byte(content))
		decoderJS = "d=atob(c);"
	}

	iterations := min(wasteCycles, tm.config.MaxJsWasteCycles) // Safety cap

	var mathExprs []string
	for i := 0; i < rand.IntN(5)+2; i++ { // 2-6 lines of math
		mathExprs = append(mathExprs, "w+="+randomJSExpr(2)+";")
	}
	mathWaste := strings.Join(mathExprs, "")
	wasteJS := fmt.Sprintf(`let w=0;for(let i=0;i<%d;i++){%s}`, iterations, mathWaste)

	sideEffectJS := fmt.Sprintf(`p.style.borderLeft='%dpx solid transparent';p.style.opacity=(w%%100)/100+0.01;`, rand.IntN(5)+1)

	script := fmt.Sprintf(`(function(){let c='%s',d; %s; let p=document.getElementById('%s');try{%s;p.innerHTML=d;}catch(e){} %s})();`,
		encodedContent, wasteJS, placeholderID, decoderJS, sideEffectJS)

	var builder strings.Builder
	_, err := fmt.Fprintf(&builder, `<%s id="%s"></%s>`, tag, placeholderID, tag)
	if err != nil {
		return ""
	}
	_, err = fmt.Fprintf(&builder, "<script>%s</script>", script)
	if err != nil {
		return ""
	}

	return template.HTML(builder.String())
}

// xorBytes is a helper for XOR encoding.
func xorBytes(input []byte, key byte) []byte {
	output := make([]byte, len(input))
	for i, b := range input {
		output[i] = b ^ key
	}
	return output
}

// reverseString is a helper for reversing a string.
func reverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// randomJSExpr recursively builds a random, complex math expression string.
func randomJSExpr(depth int) string {
	if depth <= 0 || rand.Float64() < 0.3 {
		return jsVars[rand.IntN(len(jsVars))]
	}
	if rand.Float64() < 0.5 {
		op := jsUnaryOps[rand.IntN(len(jsUnaryOps))]
		return fmt.Sprintf(op, randomJSExpr(depth-1))
	}
	op := jsBinaryOps[rand.IntN(len(jsBinaryOps))]
	return fmt.Sprintf(op, randomJSExpr(depth-1), randomJSExpr(depth-1))
}
