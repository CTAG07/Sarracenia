package markov

import (
	"context"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	ctx, g, modelInfo := setupTestDBWithTraining(t)

	// With temperature 0, the output is deterministic.
	// It will always pick the most frequent start token, which is a tie.
	output, err := g.Generate(ctx, modelInfo, WithMaxLength(10), WithTemperature(0))
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expected1 := "one fish two fish."
	expected2 := "red fish blue fish."
	if output != expected1 && output != expected2 {
		t.Errorf("Generate() got = %v, want one of [%q, %q]", output, expected1, expected2)
	}
}

func TestGenerateFrom(t *testing.T) {
	ctx, g, modelInfo := setupTestDBWithTraining(t)

	testCases := []struct {
		name          string
		seed          string
		maxLength     int
		canEndEarly   bool
		expected      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "Successful generation from seed",
			seed:        "one fish",
			maxLength:   4,
			canEndEarly: true,
			expected:    "one fish two fish",
		},
		{
			name:        "Early end with EOC",
			seed:        "red fish",
			maxLength:   10,
			canEndEarly: true,
			expected:    "red fish blue fish.",
		},
		{
			name:        "Generation stopped by maxLength",
			seed:        "one fish",
			maxLength:   3,
			canEndEarly: true,
			expected:    "one fish two",
		},
		{
			name:      "Seed is longer than maxLength",
			seed:      "one fish two fish",
			maxLength: 3,
			expected:  "one fish two",
		},
		{
			name:          "Seed contains unknown token",
			seed:          "green fish",
			maxLength:     10,
			canEndEarly:   true,
			expectError:   true,
			errorContains: "not found in model vocabulary",
		},
		{
			name:        "Empty seed behaves like Generate",
			seed:        "",
			maxLength:   10,
			canEndEarly: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := g.GenerateFromString(ctx, modelInfo, tc.seed, WithMaxLength(tc.maxLength), WithEarlyTermination(tc.canEndEarly), WithTemperature(0))

			if tc.expectError {
				if err == nil {
					t.Errorf("expected an error but got none")
				} else if !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("expected error to contain %q, but got %q", tc.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("got unexpected error: %v", err)
			}
			if tc.seed == "" {
				expected1 := "one fish two fish."
				expected2 := "red fish blue fish."
				if output != expected1 && output != expected2 {
					t.Errorf("with empty seed got = %v, want one of [%q, %q]", output, expected1, expected2)
				}
			} else if output != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, output)
			}
		})
	}
}

func BenchmarkGenerate(b *testing.B) {
	corpus := createBenchmarkCorpus()
	ctx := context.Background()
	_, g := setupTestDBBench(b)

	model := ModelInfo{Name: "bench_generate", Order: 2}
	if err := g.InsertModel(ctx, model); err != nil {
		b.Fatal(err)
	}
	model, _ = g.GetModelInfo(ctx, "bench_generate")
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
				s, err := g.Generate(ctx, model, opts...)
				b.SetBytes(int64(len(s)))
				if err != nil {
					b.Fatalf("Generate() failed: %v", err)
				}
			}
		})
	}
}
