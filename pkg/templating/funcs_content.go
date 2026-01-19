package templating

import (
	"context"
	"encoding/json"
	"html/template"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/amenyxia/Sarracenia/pkg/markov"
)

const (
	upperHexChars      = "0123456789ABCDEF"
	lowerHexChars      = "0123456789abcdef"
	upperAlphabetChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerAlphabetChars = "abcdefghijklmnopqrstuvwxyz"
	alphabetChars      = upperAlphabetChars + lowerAlphabetChars
	numericChars       = "0123456789"
	alphanumericChars  = alphabetChars + numericChars
)

var emailDomains = []string{
	"gmail.com",
	"yahoo.com",
	"hotmail.com",
	"outlook.com",
	"protonmail.com",
	"icloud.com",
}

// markovSentence generates a sentence from a named model.
func (tm *TemplateManager) markovSentence(modelName string, maxLength int) (string, error) {
	if !tm.config.MarkovEnabled { // fallback to random if markov is not enabled
		return randomSentence(maxLength), nil
	}

	ctx := context.Background()

	model, ok := tm.markovModels[modelName]
	if !ok {
		tm.logger.Error("markovSentence: model not found", "model", modelName)
		return "", nil
	}

	sentence, err := tm.markovGen.Generate(ctx, model, markov.WithMaxLength(maxLength))
	if err != nil {
		tm.logger.Error("markovSentence: generation failed", "model", modelName, "error", err)
		return "", nil
	}
	return sentence, nil
}

// markovParagraphs generates N paragraphs of thematic text.
func (tm *TemplateManager) markovParagraphs(modelName string, count, minSentences, maxSentences, minLength, maxLength int) (string, error) {
	if !tm.config.MarkovEnabled { // fallback to random if markov is not enabled
		return randomParagraphs(count, minSentences, maxSentences, minLength, maxLength), nil
	}

	var builder strings.Builder
	for i := 0; i < count; i++ {
		numSentences := rand.IntN(maxSentences-minSentences) + minSentences
		for j := 0; j < numSentences; j++ {
			sentence, err := tm.markovSentence(modelName, rand.IntN(maxLength-minLength)+minLength)
			if err != nil {
				return "", err
			}
			builder.WriteString(sentence)
			builder.WriteByte(' ')
		}
		if i < count-1 {
			builder.WriteString("\n\n")
		}
	}
	return builder.String(), nil
}

// randomWord returns a single random word from the manager's loaded word list.
func randomWord() string {
	if wordCount == 0 {
		return ""
	}
	return wordList[rand.IntN(wordCount)]
}

// randomSentence generates a nonsensical sentence from random words.
func randomSentence(length int) string {
	if length <= 0 {
		return ""
	}
	var builder strings.Builder
	words := make([]string, length)
	for i := 0; i < length; i++ {
		words[i] = randomWord()
	}
	// Capitalize the first letter of the first word
	if len(words[0]) > 0 {
		words[0] = strings.ToUpper(string(words[0][0])) + words[0][1:]
	}
	builder.WriteString(strings.Join(words, " "))
	builder.WriteByte('.')
	return builder.String()
}

// randomParagraphs generates N paragraphs of filler text.
func randomParagraphs(count, minSentences, maxSentences, minLength, maxLength int) string {
	if count <= 0 {
		return ""
	}
	var builder strings.Builder
	for i := 0; i < count; i++ {
		numSentences := rand.IntN(maxSentences-minSentences) + minSentences
		for j := 0; j < numSentences; j++ {
			builder.WriteString(randomSentence(rand.IntN(maxLength-minLength) + minLength))
			builder.WriteByte(' ')
		}
		if i < count-1 {
			builder.WriteString("\n\n")
		}
	}
	return builder.String()
}

func randomString(t string, length int) string {

	var builder strings.Builder

	switch t {
	case "username":
		word1 := wordList[rand.IntN(wordCount)]
		word2 := wordList[rand.IntN(wordCount)]
		num1 := rand.IntN(10)
		num2 := rand.IntN(10)

		// Pre-allocate: length of two words + 2 digits
		builder.Grow(len(word1) + len(word2) + 2)
		builder.WriteString(word1)
		builder.WriteString(word2)
		builder.WriteString(strconv.Itoa(num1))
		builder.WriteString(strconv.Itoa(num2))
		return builder.String()

	case "email":
		word1 := wordList[rand.IntN(wordCount)]
		word2 := wordList[rand.IntN(wordCount)]
		domain := emailDomains[rand.IntN(len(emailDomains))] // Pick a random domain

		// Pre-allocate: word lengths + at-sign + domain length
		builder.Grow(len(word1) + len(word2) + 1 + len(domain))
		builder.WriteString(word1)
		builder.WriteString(word2)
		builder.WriteByte('@')
		builder.WriteString(domain)
		return builder.String()

	case "uuid":
		// A UUID has a fixed length of 36 characters (32 hex + 4 dashes).
		builder.Grow(36)
		for i, n := range []int{8, 4, 4, 4, 12} {
			if i > 0 {
				builder.WriteByte('-')
			}
			for j := 0; j < n; j++ {
				builder.WriteByte(lowerHexChars[rand.IntN(len(lowerHexChars))])
			}
		}
		return builder.String()

	case "hex":
		if length <= 0 {
			return ""
		}
		builder.Grow(length)
		for i := 0; i < length; i++ {
			builder.WriteByte(upperHexChars[rand.IntN(len(upperHexChars))])
		}
		return builder.String()

	case "alphanum":
		if length <= 0 {
			return ""
		}
		builder.Grow(length)
		for i := 0; i < length; i++ {
			builder.WriteByte(alphanumericChars[rand.IntN(len(alphanumericChars))])
		}
		return builder.String()

	default:
		return ""
	}
}

func randomDate(layout, start, end string) (string, error) {
	startTime, err := time.Parse(layout, start)
	if err != nil {
		return "", err
	}
	endTime, err := time.Parse(layout, end)
	if err != nil {
		return "", err
	}

	if startTime.After(endTime) {
		startTime, endTime = endTime, startTime
	}

	delta := endTime.Unix() - startTime.Unix()
	if delta <= 0 {
		return startTime.Format(layout), nil
	}

	sec := rand.Int64N(delta) + startTime.Unix()

	return time.Unix(sec, 0).Format(layout), nil
}

// randomJSON generates a random, nested JSON object as a string.
func (tm *TemplateManager) randomJSON(requestedDepth, maxElements, maxStringLength int) (template.HTML, error) {

	depth := min(requestedDepth, tm.config.MaxJSONDepth)

	data := generateRandomJSONValue(depth, maxElements, maxStringLength)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		tm.logger.Error("failed to marshal random JSON", "error", err)
		return "", err
	}

	return template.HTML(jsonData), nil
}

func generateRandomJSONValue(depth, maxElements, maxStringLength int) any {
	// When depth is zero, return a primitive value.
	if depth <= 0 {
		switch rand.IntN(4) {
		case 0: // string
			return randomId("val", maxStringLength)
		case 1: // integer
			return rand.IntN(10000)
		case 2: // boolean
			return rand.IntN(2) == 0
		default: // float
			return rand.Float64() * 1000
		}
	}

	// Recursive step: Create an object or an array.
	if rand.IntN(2) == 0 {
		// Create a JSON object (map).
		obj := make(map[string]any)
		numElements := rand.IntN(maxElements) + 1
		for i := 0; i < numElements; i++ {
			key := randomId("key", 8)
			obj[key] = generateRandomJSONValue(depth-1, maxElements, maxStringLength)
		}
		return obj
	}

	// Create a JSON array (slice).
	arr := make([]any, 0)
	numElements := rand.IntN(maxElements) + 1
	for i := 0; i < numElements; i++ {
		arr = append(arr, generateRandomJSONValue(depth-1, maxElements, maxStringLength))
	}
	return arr
}
