package templating

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestTemplateFunctions validates the behavior of each category of template functions.
func TestTemplateFunctions(t *testing.T) {
	tm := setupTestManager(t)

	t.Run("ContentFuncs", func(t *testing.T) {
		word := randomWord()
		if !containsString(wordList, word) {
			t.Errorf("randomWord() returned '%s', which is not in the global word list", word)
		}

		// This test now works because setupTestManager correctly inserts and trains the model.
		sent, err := tm.markovSentence("test_model", 5)
		if err != nil {
			t.Fatalf("markovSentence failed: %v", err)
		}
		// Based on the training data "one two three four." and "one two three five."
		// a likely output starting from a common prefix is "one two three four." or "one two three five."
		if !strings.HasPrefix(sent, "one two three") {
			t.Errorf("markovSentence returned unexpected sentence: '%s'", sent)
		}

		jsonData, err := tm.randomJSON(tm.config.MaxJSONDepth+1, 2, 2)
		if err != nil {
			t.Fatalf("randomJSON failed: %v", err)
		}
		var data interface{}
		if err = json.Unmarshal([]byte(jsonData), &data); err != nil {
			t.Errorf("randomJSON did not produce valid JSON: %v", err)
		}
	})

	t.Run("StructureFuncs", func(t *testing.T) {
		// Test safety limits
		formHTML := tm.randomForm(999, 1)
		if strings.Count(string(formHTML), "<input") != tm.config.MaxFormFields {
			t.Errorf("randomForm did not respect MaxFormFields limit")
		}
		divHTML := tm.nestDivs(999)
		if divs := strings.Count(string(divHTML), "<div"); divs != tm.config.MaxNestDivs {
			t.Errorf("nestDivs did not respect MaxNestDivs limit, generated %d divs", divs)
		}
	})

	t.Run("StylingFuncs", func(t *testing.T) {
		if !strings.HasPrefix(randomColor(), "#") || len(randomColor()) != 7 {
			t.Error("randomColor has incorrect format")
		}
		if !strings.HasPrefix(randomId("prefix", 8), "prefix-") {
			t.Error("randomId has incorrect format")
		}
		classes := string(randomClasses(3))
		if strings.Count(classes, " ") != 2 {
			t.Error("randomClasses generated wrong number of classes")
		}
		style := string(randomInlineStyle(1))
		if !strings.HasPrefix(style, `style="`) || !strings.HasSuffix(style, `"`) {
			t.Error("randomInlineStyle has incorrect format")
		}
	})

	t.Run("LinkFuncs", func(t *testing.T) {
		link := tm.randomLink()
		if !strings.HasPrefix(link, "/") || strings.HasPrefix(link, "/api") {
			t.Error("randomLink failed whitelist check or format")
		}
		jsonHTML, _ := tm.randomJSON(99, 2, 10)
		var data any
		// Check if it's valid JSON
		if err := json.Unmarshal([]byte(jsonHTML), &data); err != nil {
			t.Errorf("randomJSON did not produce valid JSON: %v", err)
		}
	})

	t.Run("LogicFuncs", func(t *testing.T) {
		if len(repeat(5)) != 5 {
			t.Error("repeat failed")
		}
		choice := randomChoice([]string{"a", "b", "c"}).(string)
		if choice != "a" && choice != "b" && choice != "c" {
			t.Error("randomChoice failed")
		}
		if randomInt(10, 11) != 10 {
			t.Error("randomInt failed")
		}
	})

	t.Run("SimpleFuncs", func(t *testing.T) {
		if add(2, 3) != 5 {
			t.Error("add failed")
		}
		if min(2, 3) != 2 {
			t.Error("min failed")
		}
		if mod(10, 3) != 1 {
			t.Error("mod failed")
		}
	})

	t.Run("ExpensiveFuncs", func(t *testing.T) {
		// Test jsInteractiveContent obfuscation and structure
		secret := "my-secret-data"
		jsHTML := tm.jsInteractiveContent("div", secret, 1000)
		if strings.Contains(string(jsHTML), secret) {
			t.Error("jsInteractiveContent failed to obfuscate content")
		}
		if !strings.Contains(string(jsHTML), "<script>") || !strings.Contains(string(jsHTML), `id="p-`) {
			t.Error("jsInteractiveContent did not generate correct HTML structure")
		}
		// Test SVG generation (just ensure it's not empty)
		svgHTML := tm.randomSVG("fractal", 5)
		if !strings.HasPrefix(string(svgHTML), "<svg") {
			t.Error("randomSVG failed to generate an SVG")
		}
	})

}

// containsString is a test helper.
func containsString(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
