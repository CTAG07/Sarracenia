package markov

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
)

// chainLink Is a struct used for batching chain inserts.
type chainLink struct {
	prefixID    int
	nextTokenID int
}

// InsertToken provides a low-level way to insert or increment a single
// chain link (`prefix -> token`) for a given model.
// For most use cases, the high-level Train function is recommended as it is
// significantly more efficient for bulk data.
func (g *Generator) InsertToken(ctx context.Context, model ModelInfo, prefix string, token int) error {
	var prefixID int
	err := g.stmtGetPrefixID.QueryRowContext(ctx, prefix).Scan(&prefixID)
	if err != nil {
		return fmt.Errorf("could not get prefix ID for '%s': %w", prefix, err)
	}
	_, err = g.stmtInsertLink.ExecContext(ctx, model.Id, prefixID, token)
	if err != nil {
		return fmt.Errorf("could not insert token for '%s': %w", prefix, err)
	}
	return nil
}

// Train processes a stream of text from an io.Reader, tokenizes it, and uses
// it to train the specified Markov model. The training process is highly
// optimized, using in-memory caching and database batching to handle large
// datasets efficiently. The entire operation is performed within a single
// database transaction to ensure data integrity.
func (g *Generator) Train(ctx context.Context, model ModelInfo, data io.Reader) error {
	// maxSentenceLength prevents massive sentences from taking up a large amount of memory
	const maxSentenceLength = 4096
	// chainBatchSize determines how many chain links are buffered in memory before being written to the database in a single batch.
	const chainBatchSize = 1000

	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// All transaction-specific statements will also be closed with this or the .Commit()
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	prefixCache := make(map[string]int)

	chainBatch := make([]chainLink, 0, chainBatchSize)

	var sentenceCount int64

	stmtInsertVocab := tx.StmtContext(ctx, g.stmtInsertVocab)
	stmtGetOrInsertPrefix := tx.StmtContext(ctx, g.stmtGetOrInsertPrefix)
	stmtInsertChainBatch, err := tx.PrepareContext(ctx, `INSERT INTO markov_chains (model_id, prefix_id, next_token_id, frequency) VALUES (?, ?, ?, 1) ON CONFLICT(model_id, prefix_id, next_token_id) DO UPDATE SET frequency = frequency + 1;`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch chain insert statement: %w", err)
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmtInsertChainBatch)

	commitChainBatch := func(batch *[]chainLink) error {
		if len(*batch) == 0 {
			return nil
		}
		for _, link := range *batch {
			if _, err := stmtInsertChainBatch.ExecContext(ctx, model.Id, link.prefixID, link.nextTokenID); err != nil {
				return fmt.Errorf("failed during batch insert of chain link (%d -> %d): %w", link.prefixID, link.nextTokenID, err)
			}
		}
		*batch = (*batch)[:0]
		return nil
	}

	stream := g.tokenizer.NewStream(data)
	var currentSentence []int
	var token *Token

	for {
		token, err = stream.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("tokenizer error: %w", err)
		}

		if !token.EOC && len(currentSentence) < maxSentenceLength {
			var tokenID int
			if err = stmtInsertVocab.QueryRowContext(ctx, token.Text).Scan(&tokenID); err != nil {
				return fmt.Errorf("sql insert vocabulary error for token '%s': %w", token.Text, err)
			}
			currentSentence = append(currentSentence, tokenID)
		} else {
			if len(currentSentence) > 0 {
				if err := processSentence(ctx, model, currentSentence, prefixCache, &chainBatch, stmtGetOrInsertPrefix); err != nil {
					return fmt.Errorf("sentence processing error: %w", err)
				}
				sentenceCount++
				currentSentence = currentSentence[:0]
			}

			if len(chainBatch) >= chainBatchSize {
				if err := commitChainBatch(&chainBatch); err != nil {
					return err
				}
			}
		}
	}

	if len(currentSentence) > 0 {
		if err := processSentence(ctx, model, currentSentence, prefixCache, &chainBatch, stmtGetOrInsertPrefix); err != nil {
			return fmt.Errorf("final sentence processing error: %w", err)
		}
		sentenceCount++
	}

	if err := commitChainBatch(&chainBatch); err != nil {
		return err
	}

	g.logger.InfoContext(ctx, "Training completed",
		slog.String("model_name", model.Name),
		slog.Int("model_id", model.Id),
		slog.Int64("sentences_processed", sentenceCount),
	)

	return tx.Commit()
}

func processSentence(ctx context.Context, model ModelInfo, sentence []int, prefixCache map[string]int, chainBatch *[]chainLink, stmtGetOrInsertPrefix *sql.Stmt) error {
	if len(sentence) == 0 {
		return nil
	}

	fullSlice := make([]int, len(sentence)+model.Order+1)
	copy(fullSlice[model.Order:len(fullSlice)-1], sentence)
	fullSlice[len(fullSlice)-1] = EOCTokenID

	var keyBuf []byte
	for i := 0; i < len(sentence)+1; i++ { // Iterate len+1 to include the final EOC token.
		prefixSlice := fullSlice[i : i+model.Order]
		nextToken := fullSlice[i+model.Order]

		keyBuf = keyBuf[:0]
		for j, tokenID := range prefixSlice {
			if j > 0 {
				keyBuf = append(keyBuf, ' ')
			}
			keyBuf = strconv.AppendInt(keyBuf, int64(tokenID), 10)
		}
		prefixKey := string(keyBuf)

		var prefixID int
		var ok bool
		if prefixID, ok = prefixCache[prefixKey]; !ok {
			if err := stmtGetOrInsertPrefix.QueryRowContext(ctx, prefixKey).Scan(&prefixID); err != nil {
				return fmt.Errorf("failed to get or insert prefix '%s': %w", prefixKey, err)
			}
			prefixCache[prefixKey] = prefixID
		}

		*chainBatch = append(*chainBatch, chainLink{prefixID: prefixID, nextTokenID: nextToken})
	}
	return nil
}
