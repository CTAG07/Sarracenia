package markov

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestPruneModel(t *testing.T) {
	db, g := setupTestDB(t)
	ctx := context.Background()

	pruneModel := ModelInfo{Name: "prune_test", Order: 1}
	_ = g.InsertModel(ctx, pruneModel)
	pruneModel, _ = g.GetModelInfo(ctx, pruneModel.Name)
	_ = g.Train(ctx, pruneModel, strings.NewReader("a b c. a b d."))
	// Chain "a" -> "b" has freq 2. Chains "b" -> "c" and "b" -> "d" have freq 1.

	if err := g.PruneModel(ctx, pruneModel, 1); err != nil {
		t.Fatalf("PruneModel failed: %v", err)
	}

	// Verify that freq=1 chains are gone.
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM markov_chains WHERE model_id = ? AND frequency <= 1", pruneModel.Id).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected 0 chains with frequency 1 after pruning, got %d", count)
	}
}

func TestVocabularyPrune(t *testing.T) {
	db, g := setupTestDB(t)
	ctx := context.Background()

	pruneModel := ModelInfo{Name: "prune_vocab_test", Order: 1}
	_ = g.InsertModel(ctx, pruneModel)
	pruneModel, _ = g.GetModelInfo(ctx, pruneModel.Name)

	_ = g.Train(ctx, pruneModel, strings.NewReader("a b c. a b d e."))

	// Get IDs for later verification
	cID, _ := g.VocabStr(ctx, "c")
	dID, _ := g.VocabStr(ctx, "d")
	eID, _ := g.VocabStr(ctx, "e")
	var dPrefixID int
	_ = g.stmtGetPrefixID.QueryRowContext(ctx, strconv.Itoa(dID)).Scan(&dPrefixID)

	// Prune tokens with total chain frequency < 2
	err := g.VocabularyPrune(ctx, 2)
	if err != nil {
		t.Fatalf("VocabularyPrune failed: %v", err)
	}

	// Verify 'c', 'd', 'e' are gone from vocabulary
	for _, word := range []string{"c", "d", "e"} {
		_, err := g.VocabStr(ctx, word)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("token '%s' should have been pruned but was found", word)
		}
	}
	// Verify 'a', 'b' still exist
	for _, word := range []string{"a", "b"} {
		_, err := g.VocabStr(ctx, word)
		if err != nil {
			t.Errorf("token '%s' should not have been pruned but was: %v", word, err)
		}
	}

	// Verify chains involving c, d, e are gone
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM markov_chains WHERE next_token_id IN (?, ?, ?)", cID, dID, eID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count > 0 {
		t.Errorf("found %d chains pointing to pruned tokens", count)
	}
}

func BenchmarkVocabularyPrune(b *testing.B) {
	ctx := context.Background()

	var dirtyCorpus strings.Builder
	dirtyCorpus.WriteString("common word common word common word. ")
	for i := 0; i < 500; i++ {
		dirtyCorpus.WriteString(fmt.Sprintf("unique_%d ", i))
	}
	dirtyCorpus.WriteString(".")
	corpus := dirtyCorpus.String()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		_, g := setupTestDBBench(b)
		model := ModelInfo{Name: "bench_prune", Order: 1}
		if err := g.InsertModel(ctx, model); err != nil {
			b.Fatalf("InsertModel failed: %v", err)
		}
		model, _ = g.GetModelInfo(ctx, model.Name)
		if err := g.Train(ctx, model, strings.NewReader(corpus)); err != nil {
			b.Fatalf("Train() setup failed: %v", err)
		}

		b.StartTimer()

		err := g.VocabularyPrune(ctx, 2)
		if err != nil {
			b.Fatalf("VocabularyPrune() failed: %v", err)
		}
	}
}
