package markov

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestGenerateStream(t *testing.T) {
	ctx, g, modelInfo := setupTestDBWithTraining(t)

	t.Run("Successful stream", func(t *testing.T) {
		stream, err := g.GenerateStream(ctx, modelInfo, WithMaxLength(4), WithTemperature(0))
		if err != nil {
			t.Fatalf("GenerateStream failed: %v", err)
		}

		var tokens []string
		for token := range stream {
			if !token.EOC {
				tokens = append(tokens, token.Text)
			}
		}

		// The stream might start with "one" or "red".
		got := strings.Join(tokens, "")
		isExpected := got == "one fish two fish" || got == "red fish blue fish"

		if !isExpected {
			t.Errorf("expected stream to generate a valid sequence, but got %q", got)
		}
	})

	t.Run("Stream cancellation", func(t *testing.T) {
		ctxCancel, cancel := context.WithCancel(ctx)
		defer cancel()

		streamCancel, err := g.GenerateStream(ctxCancel, modelInfo, WithMaxLength(100))
		if err != nil {
			t.Fatalf("GenerateStream failed: %v", err)
		}

		// Read one token, then cancel
		<-streamCancel
		cancel()

		// The channel should now close quickly
		timeout := time.After(100 * time.Millisecond)
		select {
		case _, ok := <-streamCancel:
			if ok {
				t.Error("channel should have been closed after context cancellation but was not")
			}
			// Success, channel is closed.
		case <-timeout:
			t.Error("timed out waiting for stream channel to close after cancellation")
		}
	})
}

func BenchmarkGenerateStream(b *testing.B) {
	corpus := createBenchmarkCorpus()
	ctx := context.Background()
	_, g := setupTestDBBench(b)

	model := ModelInfo{Name: "bench_generate_stream", Order: 2}
	if err := g.InsertModel(ctx, model); err != nil {
		b.Fatal(err)
	}
	model, _ = g.GetModelInfo(ctx, model.Name)
	if err := g.Train(ctx, model, strings.NewReader(corpus)); err != nil {
		b.Fatalf("Train() setup for benchmark failed: %v", err)
	}

	genOpts := map[string][]GenerateOption{
		"Simple":          {WithMaxLength(50), WithEarlyTermination(false)},
		"WithTemp":        {WithMaxLength(50), WithTemperature(0.7), WithEarlyTermination(false)},
		"WithTopK":        {WithMaxLength(50), WithTopK(10), WithEarlyTermination(false)},
		"WithTempAndTopK": {WithMaxLength(50), WithTemperature(0.7), WithTopK(10), WithEarlyTermination(false)},
	}

	for name, opts := range genOpts {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				stream, err := g.GenerateStream(ctx, model, opts...)
				if err != nil {
					b.Fatalf("GenerateStream() failed: %v", err)
				}
				// We must drain the channel to measure the full lifecycle
				var bytes int64
				for t := range stream {
					bytes = bytes + int64(len(t.Text))
				}
				b.SetBytes(bytes)
			}
		})
	}
}
