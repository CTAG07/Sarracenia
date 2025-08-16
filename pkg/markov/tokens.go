package markov

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
)

// Token represents a single tokenized unit of text. It contains the text itself
// and a boolean flag indicating if it marks the end of a chain (e.g., a sentence).
type Token struct {
	Text string
	EOC  bool
}

// Tokenizer is an interface that defines the contract for splitting input text
// into tokens. This allows the core generator logic to be independent of the
// specific tokenization strategy.
type Tokenizer interface {
	// NewStream returns a stateful StreamTokenizer for processing an io.Reader.
	NewStream(io.Reader) StreamTokenizer
	// Separator returns the string that should be used to join tokens
	// when building a final generated string, using the previous and current
	// tokens.
	Separator(prev, current string) string
	// EOC returns the string representation for an End-Of-Chain token
	// in the final generated output, using the last token in the sequence.
	EOC(last string) string
}

// StreamTokenizer is an interface for a stateful tokenizer that processes a
// stream of data, returning one token at a time.
type StreamTokenizer interface {
	// Next returns the next token from the stream. It returns io.EOF as the
	// error when the stream is fully consumed.
	Next() (*Token, error)
}

// GetNextTokens retrieves all possible subsequent tokens for a given prefix key
// from a specific model. It returns a slice of ChainTokens, the sum of all their
// frequencies, and any error that occurred. If the prefix is not found, it
// returns a nil slice and a total frequency of 0.
func (g *Generator) GetNextTokens(ctx context.Context, model ModelInfo, prefix string) ([]ChainToken, int, error) {
	var prefixID int
	err := g.stmtGetPrefixID.QueryRowContext(ctx, prefix).Scan(&prefixID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// This prefix has never been seen before, so there are no possible next tokens.
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("could not get prefix ID for '%s': %w", prefix, err)
	}

	rows, err := g.stmtGetChain.QueryContext(ctx, model.Id, prefixID)
	if err != nil {
		return nil, 0, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var tokens []ChainToken
	var totalFreq int
	for rows.Next() {
		var token ChainToken
		if err = rows.Scan(&token.Id, &token.Freq); err != nil {
			return nil, 0, err
		}
		tokens = append(tokens, token)
		totalFreq += token.Freq
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return tokens, totalFreq, nil
}

// VocabStr looks up a token string in the vocabulary and returns its corresponding ID.
// It returns an error if the token is not found.
func (g *Generator) VocabStr(ctx context.Context, token string) (int, error) {
	var tokenId int
	err := g.stmtGetTokenID.QueryRowContext(ctx, token).Scan(&tokenId)
	if err != nil {
		return 0, err
	}
	return tokenId, nil
}

// VocabInt looks up a token ID in the vocabulary and returns its corresponding text.
// It returns an error if the ID is not found.
func (g *Generator) VocabInt(ctx context.Context, id int) (string, error) {
	var tokenText string
	err := g.stmtGetTokenText.QueryRowContext(ctx, id).Scan(&tokenText)
	if err != nil {
		return "", err
	}
	return tokenText, nil
}
