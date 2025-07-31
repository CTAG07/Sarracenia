package templating

import (
	"html/template"
	"math/rand/v2"
	"strconv"
	"strings"
)

// DefinitionData is a simple struct to hold a term and its definition.
// It is used by the randomDefinitionData template function to return a slice
// that templates can easily range over to build definition lists (<dl>).
type DefinitionData struct {
	Term string
	Def  string
}

// randomForm generates a <form> with a specified number of varied input fields.
func (tm *TemplateManager) randomForm(count, styleCount int) template.HTML {
	count = min(count, tm.config.MaxFormFields)

	var builder strings.Builder
	inputTypes := []string{"text", "password", "radio", "checkbox", "submit", "button", "date", "email"}

	builder.WriteString("<form method=\"post\" action=\"#\">\n")
	for i := 0; i < count; i++ {
		id := randomId("input", 8)
		inputType := inputTypes[rand.IntN(len(inputTypes))]

		builder.WriteString("  <div>\n")
		builder.WriteString("    <label for=\"" + id + "\">" + randomWord() + "</label>\n")
		builder.WriteString("    <input type=\"" + inputType + "\" id=\"" + id + "\" name=\"" + id + "\" ")
		builder.WriteString(string(randomInlineStyle(styleCount)))
		builder.WriteString(">\n")
		builder.WriteString("  </div>\n")
	}
	builder.WriteString("</form>")
	return template.HTML(builder.String())
}

// randomDefinitionData returns a slice of {Term, Def} structs.
func randomDefinitionData(count, sentenceLength int) []DefinitionData {
	data := make([]DefinitionData, count)
	for i := 0; i < count; i++ {
		data[i] = DefinitionData{
			Term: randomWord(),
			Def:  randomSentence(sentenceLength),
		}
	}
	return data
}

// nestDivs generates a specified number of deeply nested <div> elements.
func (tm *TemplateManager) nestDivs(depth int) template.HTML {
	depth = min(depth, tm.config.MaxNestDivs)
	if depth <= 0 {
		return ""
	}

	var builder strings.Builder
	for i := 0; i < depth; i++ {
		builder.WriteString("<div class=\"")
		builder.WriteString(string(randomClasses(3)))
		builder.WriteString("\">")
	}
	builder.WriteString(randomSentence(5))
	for i := 0; i < depth; i++ {
		builder.WriteString("</div>")
	}
	return template.HTML(builder.String())
}

// randomComplexTable generates an irregular HTML table with random colspans.
func (tm *TemplateManager) randomComplexTable(rows, cols int) template.HTML {
	rows = min(rows, tm.config.MaxTableRows)
	cols = min(cols, tm.config.MaxTableCols)
	if rows <= 0 || cols <= 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("<table border=\"1\">\n")
	for r := 0; r < rows; r++ {
		builder.WriteString("  <tr>\n")
		c := 0
		for c < cols {
			colspan := rand.IntN(3) + 1
			if c+colspan > cols {
				colspan = cols - c
			}
			tag := "td"
			if rand.IntN(5) == 0 {
				tag = "th"
			}
			builder.WriteString("    <" + tag + " colspan=\"" + strconv.Itoa(colspan) + "\">")
			builder.WriteString(randomSentence(3))
			builder.WriteString("</" + tag + ">\n")
			c += colspan
		}
		builder.WriteString("  </tr>\n")
	}
	builder.WriteString("</table>")
	return template.HTML(builder.String())
}
