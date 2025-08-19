package markov

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
)

// GenerateStream creates a new Markov chain and returns a read-only channel of Tokens.
// This allows for processing the generated text token-by-token, which is useful for
// real-time applications or when generating very long sequences. The channel will be
// closed once generation is complete or the context is cancelled.
func (g *Generator) GenerateStream(ctx context.Context, model ModelInfo, opts ...GenerateOption) (<-chan Token, error) {
	// Start with a chain of <SOC> tokens.
	initialChain := make([]int, model.Order)
	return g.generateStreamFromChain(ctx, model, initialChain, opts...)
}

// GenerateStreamFromString is a convenience wrapper around GenerateStreamFromStream that uses a
// string as the seed. If the string is empty, it behaves identically to GenerateStream.
func (g *Generator) GenerateStreamFromString(ctx context.Context, model ModelInfo, startText string, opts ...GenerateOption) (<-chan Token, error) {
	if startText == "" {
		return g.GenerateStream(ctx, model, opts...)
	} else {
		return g.GenerateStreamFromStream(ctx, model, strings.NewReader(startText), opts...)
	}
}

// GenerateStreamFromStream uses an io.Reader as a seed to begin a streaming generation.
// It tokenizes the content from r and uses it as the initial chain state.
func (g *Generator) GenerateStreamFromStream(ctx context.Context, model ModelInfo, r io.Reader, opts ...GenerateOption) (<-chan Token, error) {
	stream := g.tokenizer.NewStream(r)

	var seedTokens []int
	for {
		token, err := stream.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tokenizer error while reading seed: %w", err)
		}

		if !token.EOC {
			var tokenID int
			err := g.stmtGetTokenID.QueryRowContext(ctx, token.Text).Scan(&tokenID)
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("seed token '%s' not found in model vocabulary", token.Text)
			}
			if err != nil {
				return nil, fmt.Errorf("failed to look up seed token '%s': %w", token.Text, err)
			}
			seedTokens = append(seedTokens, tokenID)
		}
	}

	initialChain := make([]int, model.Order)
	initialChain = append(initialChain, seedTokens...)

	return g.generateStreamFromChain(ctx, model, initialChain, opts...)
}

// generateStreamFromChain contains the core logic for streaming generation.
func (g *Generator) generateStreamFromChain(ctx context.Context, model ModelInfo, initialChain []int, opts ...GenerateOption) (<-chan Token, error) {
	options := &generateOptions{
		maxLength:   100,
		canEndEarly: true,
		temperature: 1.0,
		topK:        0,
	}
	for _, opt := range opts {
		opt(options)
	}

	tokenChan := make(chan Token)

	go func() {
		defer close(tokenChan)

		tokenCache := make(map[int]string)
		tokenCache[SOCTokenID] = SOCTokenText
		tokenCache[EOCTokenID] = EOCTokenText

		prefix := make([]int, model.Order)
		generatedCount := 0

		var seedTokens []int
		if len(initialChain) > options.maxLength+model.Order {
			seedTokens = initialChain[model.Order : options.maxLength+model.Order]
		} else {
			seedTokens = initialChain[model.Order:]
		}

		var lastWord string
		var firstWord = true

		for _, tokenID := range seedTokens {
			text, err := g.getTokenTextWithCache(ctx, tokenID, tokenCache)
			lastWord = text
			if err != nil {
				g.logger.ErrorContext(ctx, "failed to get seed token text", slog.Int("token_id", tokenID), slog.Any("error", err))
				return
			}
			select {
			case <-ctx.Done():
				return
			case tokenChan <- Token{Text: text, EOC: false}:
			}
			prefix = append(prefix[1:], tokenID)
			generatedCount++
		}

		var keyBuf []byte
		for generatedCount < options.maxLength {
			select {
			case <-ctx.Done():
				g.logger.DebugContext(ctx, "Generation stream cancelled by context")
				return
			default:
				// continue
			}

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
				g.logger.ErrorContext(ctx, "failed to get next tokens for stream", slog.String("prefix", prefixKey), slog.Any("error", err))
				return
			}

			var nextToken int
			if len(choices) > 0 {
				nextToken = chooseNextToken(choices, totalFreq, options)
			} else {
				nextToken = EOCTokenID
			}

			if nextToken == EOCTokenID {
				var separator string
				eoc := g.tokenizer.EOC(lastWord)
				if !firstWord {
					separator = g.tokenizer.Separator(lastWord, eoc)
				} else {
					firstWord = false
					separator = ""
				}
				select {
				case <-ctx.Done():
					return
				case tokenChan <- Token{Text: separator + eoc, EOC: true}:
				}
				if options.canEndEarly {
					return
				}
				clear(prefix) //
			} else {
				var text string
				text, err = g.getTokenTextWithCache(ctx, nextToken, tokenCache)
				if err != nil {
					g.logger.ErrorContext(ctx, "failed to get generated token text", slog.Int("token_id", nextToken), slog.Any("error", err))
					return
				}
				var separator string
				if !firstWord {
					separator = g.tokenizer.Separator(lastWord, text)
				} else {
					firstWord = false
					separator = ""
				}
				lastWord = text
				select {
				case <-ctx.Done():
					return
				case tokenChan <- Token{Text: separator + text, EOC: false}:
				}
				// Update the state by shifting the prefix window and adding the new token.
				prefix = append(prefix[1:], nextToken)
			}
			generatedCount++
		}
	}()

	return tokenChan, nil
}

// getTokenTextWithCache is a helper for generation to minimize DB lookups.
func (g *Generator) getTokenTextWithCache(ctx context.Context, id int, cache map[int]string) (string, error) {
	if text, ok := cache[id]; ok {
		return text, nil
	}
	text, err := g.VocabInt(ctx, id)
	if err != nil {
		return "", err
	}
	cache[id] = text
	return text, nil
}
