package markov

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"
)

// ModelInfo holds the essential metadata for a Markov model, including its
// unique ID, name, and the order of the chain (the number of preceding tokens
// used to predict the next one).
type ModelInfo struct {
	Id    int
	Name  string
	Order int
}

// ExportedModel is the serializable representation of a trained model,
// used for JSON-based import and export.
type ExportedModel struct {
	Name       string          `json:"name"`
	Order      int             `json:"order"`
	Vocabulary map[string]int  `json:"vocabulary"` // token_text -> token_id
	Prefixes   map[string]int  `json:"prefixes"`   // prefix_text -> prefix_id
	Chains     []ExportedChain `json:"chains"`
}

// ExportedChain is the serializable representation of a single link
// in a Markov chain, used within an ExportedModel.
type ExportedChain struct {
	PrefixID    int `json:"prefix_id"`
	NextTokenID int `json:"next_token_id"`
	Frequency   int `json:"frequency"`
}

// GetModelInfos retrieves metadata for all models currently in the database,
// returning them in a map keyed by model name.
func (g *Generator) GetModelInfos(ctx context.Context) (map[string]ModelInfo, error) {
	rows, err := g.stmtGetModels.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	models := make(map[string]ModelInfo)
	for rows.Next() {
		var model ModelInfo
		if err = rows.Scan(&model.Id, &model.Name, &model.Order); err != nil {
			return nil, err
		}
		models[model.Name] = model
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return models, nil
}

// GetModelInfo retrieves the metadata for a single model specified by name.
// If multiple models are needed, GetModelInfos is more efficient.
func (g *Generator) GetModelInfo(ctx context.Context, modelName string) (ModelInfo, error) {
	var modelId, modelOrder int
	err := g.stmtGetModelInfo.QueryRowContext(ctx, modelName).Scan(&modelId, &modelOrder)
	if err != nil {
		return ModelInfo{}, err
	}
	return ModelInfo{
		Id:    modelId,
		Name:  modelName,
		Order: modelOrder,
	}, nil
}

// InsertModel creates a new model entry in the database.
func (g *Generator) InsertModel(ctx context.Context, model ModelInfo) error {
	_, err := g.stmtAddModel.ExecContext(ctx, model.Name, model.Order)
	return err
}

// RemoveModel deletes a model and all of its associated chain data from the
// database. The operation is performed within a transaction.
func (g *Generator) RemoveModel(ctx context.Context, model ModelInfo) error {

	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	if _, err = tx.ExecContext(ctx, "DELETE FROM markov_chains WHERE model_id = ?", model.Id); err != nil {
		return fmt.Errorf("failed to remove chains for model %d: %w", model.Id, err)
	}

	if _, err = tx.ExecContext(ctx, "DELETE FROM markov_models WHERE model_id = ?", model.Id); err != nil {
		return fmt.Errorf("failed to remove model %d: %w", model.Id, err)
	}

	g.logger.InfoContext(ctx, "Model removed successfully",
		slog.String("model_name", model.Name),
		slog.Int("model_id", model.Id),
	)

	return tx.Commit()
}

// ExportModel serializes a given model into a JSON format and writes it to the
// provided io.Writer. This is useful for backups or for transferring models.
func (g *Generator) ExportModel(ctx context.Context, modelInfo ModelInfo, w io.Writer) error {

	// Load all chains from the model
	rows, err := g.db.QueryContext(ctx, "SELECT prefix_id, next_token_id, frequency FROM markov_chains WHERE model_id = ?", modelInfo.Id)
	if err != nil {
		return fmt.Errorf("could not query chains for export: %w", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	// Make maps for all prefixes and tokens the model uses
	var exportedChains []ExportedChain
	prefixIDs := make(map[int]struct{})
	tokenIDs := make(map[int]struct{})

	for rows.Next() {
		var chain ExportedChain
		if err := rows.Scan(&chain.PrefixID, &chain.NextTokenID, &chain.Frequency); err != nil {
			return err
		}
		exportedChains = append(exportedChains, chain)
		prefixIDs[chain.PrefixID] = struct{}{}
		tokenIDs[chain.NextTokenID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	prefixIDToText := make(map[string]int)
	if len(prefixIDs) > 0 {
		args := make([]interface{}, 0, len(prefixIDs))
		placeholders := make([]string, 0, len(prefixIDs))
		for id := range prefixIDs {
			args = append(args, id)
			placeholders = append(placeholders, "?")
		}
		// Grab every prefix we need with one query
		query := fmt.Sprintf(`SELECT prefix_id, prefix_text FROM markov_prefixes WHERE prefix_id IN (%s)`, strings.Join(placeholders, ","))
		pRows, err := g.db.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		for pRows.Next() {
			var id int
			var text string
			_ = pRows.Scan(&id, &text)
			prefixIDToText[text] = id
			for _, idStr := range strings.Split(text, " ") {
				tokenID, _ := strconv.Atoi(idStr)
				tokenIDs[tokenID] = struct{}{}
			}
		}
		_ = pRows.Close()
	}

	tokenIDToText := make(map[string]int)
	if len(tokenIDs) > 0 {
		args := make([]interface{}, 0, len(tokenIDs))
		placeholders := make([]string, 0, len(tokenIDs))
		for id := range tokenIDs {
			args = append(args, id)
			placeholders = append(placeholders, "?")
		}
		// Grab every token we need with one query
		query := fmt.Sprintf(`SELECT token_id, token_text FROM markov_vocabulary WHERE token_id IN (%s)`, strings.Join(placeholders, ","))
		vRows, err := g.db.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		for vRows.Next() {
			var id int
			var text string
			_ = vRows.Scan(&id, &text)
			tokenIDToText[text] = id
		}
		_ = vRows.Close()
	}

	exported := ExportedModel{
		Name:       modelInfo.Name,
		Order:      modelInfo.Order,
		Vocabulary: tokenIDToText,
		Prefixes:   prefixIDToText,
		Chains:     exportedChains,
	}

	g.logger.InfoContext(ctx, "Model exported",
		slog.String("model_name", modelInfo.Name),
		slog.Int("model_id", modelInfo.Id),
		slog.Int("vocab_items_exported", len(tokenIDToText)),
		slog.Int("prefixes_exported", len(prefixIDToText)),
		slog.Int("chains_exported", len(exportedChains)),
	)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(exported)
}

// ImportModel reads a JSON representation of a model from an io.Reader and
// merges its data into the database. If the model name already exists, the
// new chain data is merged with the existing data (frequencies are added).
// If the model does not exist, it is created. The entire operation is
// transactional and handles re-mapping of vocabulary and prefix IDs.
func (g *Generator) ImportModel(ctx context.Context, r io.Reader) error {
	var imported ExportedModel
	if err := json.NewDecoder(r).Decode(&imported); err != nil {
		return fmt.Errorf("failed to decode json model: %w", err)
	}

	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction for import: %w", err)
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	var modelID int
	err = tx.QueryRowContext(ctx, "SELECT model_id FROM markov_models WHERE model_name = ?", imported.Name).Scan(&modelID)
	if errors.Is(err, sql.ErrNoRows) {
		res, err := tx.ExecContext(ctx, "INSERT INTO markov_models (model_name, model_order) VALUES (?, ?)", imported.Name, imported.Order)
		if err != nil {
			return fmt.Errorf("failed to insert new model '%s': %w", imported.Name, err)
		}
		newID, _ := res.LastInsertId()
		modelID = int(newID)
	} else if err != nil {
		return fmt.Errorf("failed to query for model '%s': %w", imported.Name, err)
	}

	stmtInsertVocab := tx.StmtContext(ctx, g.stmtInsertVocab)
	stmtGetOrInsertPrefix := tx.StmtContext(ctx, g.stmtGetOrInsertPrefix)

	vocabIDMap := make(map[int]int) // old_id -> new_id
	vocabIDMap[0] = 0               // <SOC> is always 0
	vocabIDMap[1] = 1               // <EOC> is always 1

	for text, oldID := range imported.Vocabulary {
		if text == "<SOC>" || text == "<EOC>" {
			continue
		}
		var newID int
		if err := stmtInsertVocab.QueryRowContext(ctx, text).Scan(&newID); err != nil {
			return fmt.Errorf("failed to get/insert vocab '%s': %w", text, err)
		}
		vocabIDMap[oldID] = newID
	}

	// Prefixes need to be re-made with the new VocabID's
	prefixIDMap := make(map[int]int) // old_id -> new_id
	newPrefixParts := make([]string, 0, imported.Order)

	for oldPrefixText, oldPrefixID := range imported.Prefixes {
		oldTokenIDs := strings.Split(oldPrefixText, " ")
		newPrefixParts = newPrefixParts[:0]

		for _, oldTokenIDStr := range oldTokenIDs {
			oldTokenID, _ := strconv.Atoi(oldTokenIDStr)
			newTokenID, ok := vocabIDMap[oldTokenID]
			if !ok {
				return fmt.Errorf("consistency error: old token id %d in prefix not found in vocab map", oldTokenID)
			}
			newPrefixParts = append(newPrefixParts, strconv.Itoa(newTokenID))
		}

		newPrefixText := strings.Join(newPrefixParts, " ")

		var newPrefixID int
		if err := stmtGetOrInsertPrefix.QueryRowContext(ctx, newPrefixText).Scan(&newPrefixID); err != nil {
			return fmt.Errorf("failed to get/insert rebuilt prefix '%s': %w", newPrefixText, err)
		}
		prefixIDMap[oldPrefixID] = newPrefixID
	}

	// Prepare a special query so that if we're updating instead of inserting, we don't overwrite the frequency value
	stmtInsertChain, err := tx.PrepareContext(ctx, `
		INSERT INTO markov_chains (model_id, prefix_id, next_token_id, frequency) VALUES (?, ?, ?, ?)
		ON CONFLICT(model_id, prefix_id, next_token_id) DO UPDATE SET frequency = frequency + excluded.frequency;
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare chain insert statement: %w", err)
	}
	defer func(stmtInsertChain *sql.Stmt) {
		_ = stmtInsertChain.Close()
	}(stmtInsertChain)

	for _, chain := range imported.Chains {
		newPrefixID, ok := prefixIDMap[chain.PrefixID]
		if !ok {
			return fmt.Errorf("import consistency error: old prefix id %d not found in prefix map", chain.PrefixID)
		}
		newNextTokenID, ok := vocabIDMap[chain.NextTokenID]
		if !ok {
			return fmt.Errorf("import consistency error: old token id %d not found in vocab map", chain.NextTokenID)
		}

		_, err = stmtInsertChain.ExecContext(ctx, modelID, newPrefixID, newNextTokenID, chain.Frequency)
		if err != nil {
			return fmt.Errorf("failed to insert chain link (%d -> %d): %w", newPrefixID, newNextTokenID, err)
		}
	}

	g.logger.InfoContext(ctx, "Model imported successfully",
		slog.String("model_name", imported.Name),
		slog.Int("target_model_id", modelID),
		slog.Int("vocab_items_merged", len(imported.Vocabulary)),
		slog.Int("prefixes_merged", len(imported.Prefixes)),
		slog.Int("chains_merged", len(imported.Chains)),
	)

	return tx.Commit()
}
