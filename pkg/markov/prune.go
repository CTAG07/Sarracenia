package markov

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

// PruneModel removes all chain links from a specific model that have a frequency
// less than or equal to `minFreq`. This is useful for reducing the size of a model
// by removing rare, and often noisy, transitions.
func (g *Generator) PruneModel(ctx context.Context, model ModelInfo, minFreq int) error {
	res, err := g.stmtPruneModel.ExecContext(ctx, model.Id, minFreq)
	if err != nil {
		return fmt.Errorf("could not prune model %d: %w", model.Id, err)
	}
	rowsAffected, _ := res.RowsAffected()

	g.logger.InfoContext(ctx, "Model pruned",
		slog.String("model_name", model.Name),
		slog.Int("model_id", model.Id),
		slog.Int("min_frequency", minFreq),
		slog.Int64("chains_removed", rowsAffected),
	)
	return nil
}

// VocabularyPrune performs a database-wide cleanup, removing tokens from the
// global vocabulary that are used less than `minFrequency` times across all models.
// This is a destructive operation that will also delete all chain links and prefixes
// that rely on the removed tokens. It should be used with caution to reduce
// the overall database size. Special tokens (<SOC>, <EOC>) are never pruned.
func (g *Generator) VocabularyPrune(ctx context.Context, minFrequency int) error {
	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction for pruning: %w", err)
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	// Find all 'rare' tokens
	rows, err := tx.QueryContext(ctx,
		`SELECT next_token_id FROM markov_chains GROUP BY next_token_id HAVING SUM(frequency) < ? AND next_token_id NOT IN (?, ?)`,
		minFrequency, SOCTokenID, EOCTokenID)
	if err != nil {
		return fmt.Errorf("failed to query for rare tokens: %w", err)
	}

	var rareTokenIDs []int
	var rareTokenIDSet = make(map[int]struct{})
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return fmt.Errorf("failed to scan rare token id: %w", err)
		}
		rareTokenIDs = append(rareTokenIDs, id)
		rareTokenIDSet[id] = struct{}{}
	}
	_ = rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error after iterating rare token rows: %w", err)
	}

	if len(rareTokenIDs) == 0 {
		g.logger.InfoContext(ctx, "No vocabulary to prune",
			slog.Int("min_frequency", minFrequency),
		)
		return tx.Commit() // Nothing to do
	}

	// 2. Find all prefixes that contain any of the rare tokens.
	// This is the most complex part. We fetch all prefixes and check them in Go,
	// which is more portable and clearer than complex, non-SARGable SQL LIKE queries.
	pRows, err := tx.QueryContext(ctx, `SELECT prefix_id, prefix_text FROM markov_prefixes`)
	if err != nil {
		return fmt.Errorf("failed to query all prefixes for checking: %w", err)
	}

	var affectedPrefixIDs []int
	for pRows.Next() {
		var prefixID int
		var prefixText string
		if err := pRows.Scan(&prefixID, &prefixText); err != nil {
			_ = pRows.Close()
			return fmt.Errorf("failed to scan prefix row: %w", err)
		}

		tokenIDs := strings.Split(prefixText, " ")
		for _, idStr := range tokenIDs {
			id, _ := strconv.Atoi(idStr)
			if _, isRare := rareTokenIDSet[id]; isRare {
				affectedPrefixIDs = append(affectedPrefixIDs, prefixID)
				break // Found a rare token, no need to check others in this prefix
			}
		}
	}
	_ = pRows.Close()
	if err := pRows.Err(); err != nil {
		return fmt.Errorf("error after iterating prefix rows: %w", err)
	}

	// 3. Perform deletions in the correct order (chains -> prefixes -> vocabulary).
	// We build dynamic query strings to handle `IN (...)` clauses.

	// Delete chains that point to a rare token OR start from an affected prefix.
	if err := g.batchDelete(ctx, tx, "markov_chains", "next_token_id", intSliceToInterface(rareTokenIDs)); err != nil {
		return fmt.Errorf("failed to prune chains by next_token_id: %w", err)
	}
	if err := g.batchDelete(ctx, tx, "markov_chains", "prefix_id", intSliceToInterface(affectedPrefixIDs)); err != nil {
		return fmt.Errorf("failed to prune chains by prefix_id: %w", err)
	}

	// Delete the affected prefixes themselves.
	if err := g.batchDelete(ctx, tx, "markov_prefixes", "prefix_id", intSliceToInterface(affectedPrefixIDs)); err != nil {
		return fmt.Errorf("failed to prune affected prefixes: %w", err)
	}

	// Finally, delete the rare tokens from the vocabulary.
	if err := g.batchDelete(ctx, tx, "markov_vocabulary", "token_id", intSliceToInterface(rareTokenIDs)); err != nil {
		return fmt.Errorf("failed to prune rare tokens from vocabulary: %w", err)
	}

	numPruned := len(rareTokenIDs)
	g.logger.InfoContext(ctx, "Vocabulary pruned successfully",
		slog.Int("min_frequency", minFrequency),
		slog.Int("tokens_removed", numPruned),
		slog.Int("prefixes_affected", len(affectedPrefixIDs)),
	)

	return tx.Commit()
}

// batchDelete is a private helper to robustly delete from a table. It handles empty lists and splits large lists into smaller batches to avoid SQL limits.
func (g *Generator) batchDelete(ctx context.Context, tx *sql.Tx, table, column string, ids []interface{}) error {
	if len(ids) == 0 {
		return nil
	}

	// SQLite's default variable limit is 999, so around half that is good
	const batchSize = 500

	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[i:end]

		query := fmt.Sprintf("DELETE FROM %s WHERE %s IN (?%s)", table, column, strings.Repeat(",?", len(batch)-1))

		if _, err := tx.ExecContext(ctx, query, batch...); err != nil {
			return err
		}
	}
	return nil
}

// intSliceToInterface is a helper to convert []int to []interface{} for SQL args.
func intSliceToInterface(s []int) []interface{} {
	if s == nil {
		return nil
	}
	i := make([]interface{}, len(s))
	for j, v := range s {
		i[j] = v
	}
	return i
}
