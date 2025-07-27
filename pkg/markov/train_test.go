package markov

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestTrain(t *testing.T) {
	db, g := setupTestDB(t)
	ctx := context.Background()
	modelInfo := ModelInfo{Name: "train_test", Order: 2}

	if err := g.InsertModel(ctx, modelInfo); err != nil {
		t.Fatalf("InsertModel failed: %v", err)
	}
	modelInfo, _ = g.GetModelInfo(ctx, modelInfo.Name)

	trainingData := "a b c. a b d."
	if err := g.Train(ctx, modelInfo, strings.NewReader(trainingData)); err != nil {
		t.Fatalf("Train() failed: %v", err)
	}

	// Verify that chains were created
	var chainCount int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM markov_chains WHERE model_id = ?", modelInfo.Id).Scan(&chainCount)
	if err != nil {
		t.Fatal(err)
	}
	if chainCount < 4 { // "a b" -> c, "a b" -> d, etc.
		t.Errorf("expected at least 4 chains to be created, but got %d", chainCount)
	}

	// Verify that a specific chain has the correct frequency
	aID, _ := g.VocabStr(ctx, "a")
	bID, _ := g.VocabStr(ctx, "b")
	prefixKey := strings.Join([]string{fmt.Sprint(aID), fmt.Sprint(bID)}, " ")
	tokens, totalFreq, err := g.GetNextTokens(ctx, modelInfo, prefixKey)
	if err != nil {
		t.Fatalf("GetNextTokens failed: %v", err)
	}
	if totalFreq != 2 {
		t.Errorf("expected prefix 'a b' to have total frequency of 2, got %d", totalFreq)
	}
	if len(tokens) != 2 {
		t.Errorf("expected prefix 'a b' to lead to 2 unique next tokens, got %d", len(tokens))
	}
}

func BenchmarkTrain(b *testing.B) {
	corpus := createBenchmarkCorpus()
	ctx := context.Background()

	for _, order := range []int{1, 2, 3, 4, 5} {
		b.Run(fmt.Sprintf("Order%d", order), func(b *testing.B) {
			_, g := setupTestDBBench(b)
			model := ModelInfo{Name: "bench_train", Order: order}
			if err := g.InsertModel(ctx, model); err != nil {
				b.Fatalf("InsertModel failed: %v", err)
			}
			model, _ = g.GetModelInfo(ctx, model.Name)

			b.SetBytes(int64(len(corpus)))
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				err := g.Train(ctx, model, strings.NewReader(corpus))
				if err != nil {
					b.Fatalf("Train() failed: %v", err)
				}
			}
		})
	}
}
