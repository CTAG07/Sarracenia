package markov

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
)

// ChainToken represents a potential next token in a Markov chain, including its
// unique ID and its frequency of occurrence after a given prefix.
type ChainToken struct {
	Id   int
	Freq int
}

// generateOptions Is used by the generate functions to configure default options.
type generateOptions struct {
	maxLength   int
	canEndEarly bool
	// New optional parameters
	temperature float64
	topK        int
}

// GenerateOption is a function that configures generation parameters. It's used
// as a variadic argument in generation functions like Generate and GenerateStream.
type GenerateOption func(*generateOptions)

// WithMaxLength sets the maximum number of tokens to generate. The generation
// may stop earlier if an EOC token is chosen and WithEarlyTermination is enabled.
func WithMaxLength(n int) GenerateOption {
	return func(o *generateOptions) { o.maxLength = n }
}

// WithEarlyTermination specifies whether the generation process can stop before
// reaching maxLength if an End-Of-Chain (EOC) token is generated.
func WithEarlyTermination(canEnd bool) GenerateOption {
	return func(o *generateOptions) { o.canEndEarly = canEnd }
}

// WithTemperature adjusts the randomness of the token selection.
// A value of 1.0 is standard weighted random selection.
// Values > 1.0 increase randomness (making less frequent tokens more likely).
// Values < 1.0 decrease randomness (making more frequent tokens even more likely).
// A value of 0 or less results in deterministic selection (always choosing the most frequent token).
func WithTemperature(t float64) GenerateOption {
	return func(o *generateOptions) { o.temperature = t }
}

// WithTopK restricts the token selection pool to the top `k` most frequent tokens
// at each step. A value of 0 disables Top-K sampling.
func WithTopK(k int) GenerateOption {
	return func(o *generateOptions) { o.topK = k }
}

// Generate creates a new Markov chain, builds it into a single string, and returns it.
// It starts from a default initial state of Start-Of-Chain (SOC) tokens.
// Generation can be customized with GenerateOption functions.
func (g *Generator) Generate(ctx context.Context, model ModelInfo, opts ...GenerateOption) (string, error) {
	// Start with a chain of <SOC> tokens.
	initialChain := make([]int, model.Order)

	options := &generateOptions{
		maxLength:   100,
		canEndEarly: true,
		temperature: 1.0,
		topK:        0,
	}
	for _, opt := range opts {
		opt(options)
	}

	return g.generateChain(ctx, model, initialChain, options)
}

// GenerateFromStream uses the content of an io.Reader as a seed for generation.
// The provided text is tokenized and used as the initial state of the chain,
// from which generation continues. An error is returned if a seed token is not
// found in the model's vocabulary.
func (g *Generator) GenerateFromStream(ctx context.Context, model ModelInfo, r io.Reader, opts ...GenerateOption) (string, error) {
	stream := g.tokenizer.NewStream(r)

	var seedTokens []int
	for {
		token, err := stream.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("tokenizer error while reading seed: %w", err)
		}

		// We do not process EOC tokens from the seed, as we want to continue the chain from it.
		if !token.EOC {
			var tokenID int
			err := g.stmtGetTokenID.QueryRowContext(ctx, token.Text).Scan(&tokenID)
			if errors.Is(err, sql.ErrNoRows) {
				return "", fmt.Errorf("seed token '%s' not found in model vocabulary", token.Text)
			}
			if err != nil {
				return "", fmt.Errorf("failed to look up seed token '%s': %w", token.Text, err)
			}
			seedTokens = append(seedTokens, tokenID)
		}
	}

	// Construct the initial chain, starting with <SOC> tokens and appending the seed.
	initialChain := make([]int, model.Order)
	initialChain = append(initialChain[:model.Order], seedTokens...)

	options := &generateOptions{
		maxLength:   100,
		canEndEarly: true,
		temperature: 1.0,
		topK:        0,
	}
	for _, opt := range opts {
		opt(options)
	}

	// Ensure that the initial chain is not larger than the max length
	if len(initialChain) >= options.maxLength {
		initialChain = initialChain[:options.maxLength]
	}

	return g.generateChain(ctx, model, initialChain, options)
}

// GenerateFromString is a convenience wrapper around GenerateFromStream that uses a
// string as the seed. If the string is empty, it behaves identically to Generate.
func (g *Generator) GenerateFromString(ctx context.Context, model ModelInfo, startText string, opts ...GenerateOption) (string, error) {
	if startText == "" { // Why would you ever call this with an empty string
		return g.Generate(ctx, model, opts...)
	} else {
		return g.GenerateFromStream(ctx, model, strings.NewReader(startText), opts...)
	}
}

// generateChain contains the main loop for generating a markov chain.
func (g *Generator) generateChain(ctx context.Context, model ModelInfo, initialChain []int, options *generateOptions) (string, error) {
	var builder strings.Builder

	tokenCache := make(map[int]string)
	tokenCache[SOCTokenID] = SOCTokenText
	tokenCache[EOCTokenID] = EOCTokenText

	prefix := make([]int, model.Order)
	generatedCount := 0
	firstWord := true
	var lastWord = SOCTokenText

	if len(initialChain) > model.Order { // If we have seed tokens to deal with
		seedTokens := initialChain[model.Order:]
		if len(initialChain) > options.maxLength+model.Order {
			seedTokens = initialChain[model.Order : options.maxLength+model.Order]
		}
		for _, tokenID := range seedTokens {
			text, err := g.getTokenTextWithCache(ctx, tokenID, tokenCache)
			if err != nil {
				return "", fmt.Errorf("failed to get text for seed token %d: %w", tokenID, err)
			}
			if !firstWord {
				builder.WriteString(g.tokenizer.Separator(lastWord, text))
			}
			lastWord = text
			builder.WriteString(text)
			firstWord = false

			prefix = append(prefix[1:], tokenID)
			generatedCount++
		}
	}

	var keyBuf []byte
	terminatedEarly := false

	for generatedCount < options.maxLength {
		keyBuf = keyBuf[:0]
		for j, tokenID := range prefix {
			if j > 0 {
				keyBuf = append(keyBuf, ' ')
			}
			keyBuf = strconv.AppendInt(keyBuf, int64(tokenID), 10)
		}
		prefixKey := string(keyBuf)

		choices, totalFreq, err := g.GetNextTokens(ctx, model, prefixKey)
		if err != nil {
			return "", fmt.Errorf("failed to get next tokens for prefix '%s': %w", prefixKey, err)
		}

		if len(choices) == 0 { // Dead end in chain
			terminatedEarly = true
			g.logger.DebugContext(ctx, "Generation terminated due to dead-end",
				slog.String("model_name", model.Name),
				slog.Int("model_id", model.Id),
				slog.String("last_prefix", prefixKey),
				slog.Int("generated_length", generatedCount),
			)
			// Append EOC token to the output string.
			builder.WriteString(g.tokenizer.EOC(lastWord))
			break
		}

		nextToken := chooseNextToken(choices, totalFreq, options)

		if nextToken == EOCTokenID {
			// Default assumes no separator before EOC,
			// and that a separator between EOC and a token after is needed
			builder.WriteString(g.tokenizer.EOC(lastWord))

			if options.canEndEarly {
				terminatedEarly = true
				g.logger.DebugContext(ctx, "Generation terminated by EOC token",
					slog.String("model_name", model.Name),
					slog.Int("model_id", model.Id),
					slog.Int("generated_length", generatedCount),
				)
				break
			}

			lastWord = EOCTokenText
			clear(prefix)
		} else {
			var text string
			text, err = g.getTokenTextWithCache(ctx, nextToken, tokenCache)
			if err != nil {
				return "", fmt.Errorf("failed to get text for generated token %d: %w", nextToken, err)
			}
			if !firstWord {
				builder.WriteString(g.tokenizer.Separator(lastWord, text))
			} else {
				firstWord = false
			}
			lastWord = text
			builder.WriteString(text)

			prefix = append(prefix[1:], nextToken)
		}
		generatedCount++
	}

	if !terminatedEarly {
		// Ensure that all returned sentences end with an EOC for standardization purposes.
		builder.WriteString(g.tokenizer.EOC(lastWord))
		g.logger.DebugContext(ctx, "Generation terminated by reaching maxLength",
			slog.String("model_name", model.Name),
			slog.Int("model_id", model.Id),
			slog.Int("max_length", options.maxLength),
			slog.Int("generated_length", generatedCount),
		)
	}

	return builder.String(), nil
}

// chooseNextToken abstracts the token selection logic from the generation loop.
// It is a private helper to avoid duplicating the selection logic.
func chooseNextToken(choices []ChainToken, totalFreq int, options *generateOptions) int {
	var nextToken int

	// topK filtering
	if options.topK > 0 && options.topK < len(choices) {
		sort.Slice(choices, func(i, j int) bool {
			return choices[i].Freq > choices[j].Freq
		})
		choices = choices[:options.topK]
		totalFreq = 0
		for _, choice := range choices {
			totalFreq += choice.Freq
		}
	}

	// temperature selection
	if options.temperature <= 0 { // Deterministic
		maxFreq := -1
		for _, choice := range choices {
			if choice.Freq > maxFreq {
				maxFreq = choice.Freq
				nextToken = choice.Id
			}
		}
	} else if options.temperature == 1.0 { // Standard weighted random
		randChoice := rand.IntN(totalFreq)
		for _, choice := range choices {
			randChoice -= choice.Freq
			if randChoice < 0 {
				nextToken = choice.Id
				break
			}
		}
	} else { // Temperature-based sampling
		logProbabilities := make([]float64, len(choices))
		epsilon := -1e9
		for i, choice := range choices {
			lp := math.Log(float64(choice.Freq)) / options.temperature
			logProbabilities[i] = lp
			if lp > epsilon {
				epsilon = lp
			}
		}
		var totalWeight float64
		weights := make([]float64, len(choices))
		for i, lp := range logProbabilities {
			w := math.Exp(lp - epsilon)
			weights[i] = w
			totalWeight += w
		}
		randChoice := rand.Float64() * totalWeight
		for i, choice := range choices {
			randChoice -= weights[i]
			if randChoice < 0 {
				nextToken = choice.Id
				break
			}
		}
	}
	return nextToken
}
