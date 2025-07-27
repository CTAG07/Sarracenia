package markov

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestVocabLookup(t *testing.T) {
	ctx, g, _ := setupTestDBWithTraining(t)

	id, err := g.VocabStr(ctx, "fish")
	if err != nil {
		t.Fatalf("VocabStr('fish') failed: %v", err)
	}
	if id == 0 {
		t.Error("expected a non-zero ID for 'fish'")
	}

	text, err := g.VocabInt(ctx, id)
	if err != nil {
		t.Fatalf("VocabInt(%d) failed: %v", id, err)
	}
	if text != "fish" {
		t.Errorf("expected 'fish', got '%s'", text)
	}
}

func TestGetNextTokens(t *testing.T) {
	ctx, g, modelInfo := setupTestDBWithTraining(t)

	// Prefix "one fish" is followed by "two" in the training data.
	oneId, _ := g.VocabStr(ctx, "one")
	fishId, _ := g.VocabStr(ctx, "fish")
	twoId, _ := g.VocabStr(ctx, "two")

	prefixKey := strings.Join([]string{strconv.Itoa(oneId), strconv.Itoa(fishId)}, " ")
	tokens, totalFreq, err := g.GetNextTokens(ctx, modelInfo, prefixKey)
	if err != nil {
		t.Fatalf("GetNextTokens failed: %v", err)
	}
	if totalFreq != 1 {
		t.Errorf("expected total frequency of 1, got %d", totalFreq)
	}
	expectedTokens := []ChainToken{{Id: twoId, Freq: 1}}
	if !reflect.DeepEqual(tokens, expectedTokens) {
		t.Errorf("expected tokens %+v, got %+v", expectedTokens, tokens)
	}

	// Test unseen prefix
	tokens, totalFreq, err = g.GetNextTokens(ctx, modelInfo, "999 998")
	if err != nil {
		t.Fatalf("GetNextTokens for unseen prefix failed: %v", err)
	}
	if len(tokens) != 0 || totalFreq != 0 {
		t.Error("expected no tokens for an unseen prefix")
	}
}

func TestInsertToken(t *testing.T) {
	ctx, g, modelInfo := setupTestDBWithTraining(t)

	// We can use the main model since this is an isolated test of the function.
	blueId, _ := g.VocabStr(ctx, "blue")
	fishId, _ := g.VocabStr(ctx, "fish")
	redId, _ := g.VocabStr(ctx, "red")
	prefixKey := strings.Join([]string{strconv.Itoa(blueId), strconv.Itoa(fishId)}, " ")

	if err := g.InsertToken(ctx, modelInfo, prefixKey, redId); err != nil {
		t.Fatalf("InsertToken failed: %v", err)
	}

	// "blue fish" is followed by EOC (freq 1) and now also "red" (freq 1)
	tokens, totalFreq, _ := g.GetNextTokens(ctx, modelInfo, prefixKey)
	if totalFreq != 2 {
		t.Errorf("expected total frequency of 2 after InsertToken, got %d", totalFreq)
	}

	var found bool
	for _, token := range tokens {
		if token.Id == redId {
			found = true
			if token.Freq != 1 {
				t.Errorf("expected freq of 1 for inserted token, got %d", token.Freq)
			}
		}
	}
	if !found {
		t.Error("did not find artificially inserted token")
	}
}
