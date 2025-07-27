package markov

import (
	"database/sql"
	"fmt"
	"io"
	"log/slog"
)

const (
	// SOCTokenID is the reserved ID for the Start-Of-Chain token.
	SOCTokenID = 0
	// EOCTokenID is the reserved ID for the End-Of-Chain token.
	EOCTokenID = 1
	// SOCTokenText is the reserved text for the Start-Of-Chain token.
	SOCTokenText = "<SOC>"
	// EOCTokenText is the reserved text for the End-Of-Chain token.
	EOCTokenText = "<EOC>"
)

// SetupSchema initializes the necessary tables and special vocabulary entries
// in the provided database. This function should be called once on a new
// database before any other operations are performed. It is idempotent and
// safe to call on an already-initialized database.
func SetupSchema(db *sql.DB) error {

	const (
		schemaVocab = `
CREATE TABLE IF NOT EXISTS markov_vocabulary (
    token_id INTEGER PRIMARY KEY,
    token_text TEXT NOT NULL UNIQUE
);
`
		schemaPrefixes = `
CREATE TABLE IF NOT EXISTS markov_prefixes (
	prefix_id INTEGER PRIMARY KEY,
	prefix_text TEXT NOT NULL UNIQUE
);
`
		schemaModels = `
CREATE TABLE IF NOT EXISTS markov_models (
    model_id INTEGER PRIMARY KEY,
    model_name TEXT NOT NULL UNIQUE,
    model_order INTEGER NOT NULL
);
`
		schemaChains = `
CREATE TABLE IF NOT EXISTS markov_chains (
    model_id INTEGER NOT NULL,
    prefix_id INTEGER NOT NULL,
    next_token_id INTEGER NOT NULL,
    frequency  INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (model_id, prefix_id, next_token_id)
);
`
	)

	startToken := fmt.Sprintf("INSERT OR IGNORE INTO markov_vocabulary (token_id, token_text) VALUES (%d, '%s');", SOCTokenID, SOCTokenText)
	endToken := fmt.Sprintf("INSERT OR IGNORE INTO markov_vocabulary (token_id, token_text) VALUES (%d, '%s');", EOCTokenID, EOCTokenText)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}

	// If the transaction succeeds, tx.Commit() will be called first, and the rollback will do nothing. If it fails, this will clean up.
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	if _, err = tx.Exec(schemaVocab); err != nil {
		return fmt.Errorf("could not create schema: %w", err)
	}

	if _, err = tx.Exec(schemaPrefixes); err != nil {
		return fmt.Errorf("could not create prefixes schema: %w", err)
	}

	if _, err = tx.Exec(schemaModels); err != nil {
		return fmt.Errorf("could not create schema: %w", err)
	}

	if _, err = tx.Exec(schemaChains); err != nil {
		return fmt.Errorf("could not create schema: %w", err)
	}

	if _, err = tx.Exec(startToken); err != nil {
		return fmt.Errorf("could not insert special tokens: %w", err)
	}

	if _, err = tx.Exec(endToken); err != nil {
		return fmt.Errorf("could not insert special tokens: %w", err)
	}

	// If all commands were successful, commit the transaction.
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}

// Generator is the main entry point for interacting with the Markov chain library.
// It holds the database connection, a tokenizer, and prepared SQL statements
// for efficient database interaction.
type Generator struct {
	db                    *sql.DB
	tokenizer             Tokenizer
	stmtGetModelInfo      *sql.Stmt
	stmtGetModels         *sql.Stmt
	stmtAddModel          *sql.Stmt
	stmtPruneModel        *sql.Stmt
	stmtModelChains       *sql.Stmt
	stmtModelStarters     *sql.Stmt
	stmtModelFreq         *sql.Stmt
	stmtInsertLink        *sql.Stmt
	stmtGetTokenID        *sql.Stmt
	stmtGetPrefixID       *sql.Stmt
	stmtGetTokenText      *sql.Stmt
	stmtGetChain          *sql.Stmt
	stmtGetVocabLen       *sql.Stmt
	stmtGetPrefixLen      *sql.Stmt
	stmtInsertVocab       *sql.Stmt
	stmtGetOrInsertPrefix *sql.Stmt
	logger                *slog.Logger
}

// NewGenerator creates and returns a new Generator. It takes a database connection
// and a Tokenizer implementation. It pre-compiles all necessary SQL statements,
// returning an error if any preparation fails.
func NewGenerator(db *sql.DB, tokenizer Tokenizer) (*Generator, error) {
	stmtGetModelInfo, err := db.Prepare(`SELECT model_id, model_order FROM markov_models WHERE model_name = ?;`)
	if err != nil {
		return nil, err
	}

	stmtGetModels, err := db.Prepare(`SELECT model_id, model_name, model_order FROM markov_models;`)
	if err != nil {
		return nil, err
	}

	stmtAddModel, err := db.Prepare(`INSERT INTO markov_models (model_name, model_order) VALUES (?, ?);`)
	if err != nil {
		return nil, err
	}

	stmtPruneModel, err := db.Prepare(`DELETE FROM markov_chains WHERE model_id = ? AND frequency <= ?;`)
	if err != nil {
		return nil, err
	}

	stmtModelChains, err := db.Prepare(`SELECT COUNT(*) FROM markov_chains WHERE model_id = ?;`)
	if err != nil {
		return nil, err
	}

	stmtModelStarters, err := db.Prepare(`SELECT COUNT(*) FROM markov_chains WHERE model_id = ? AND prefix_id = ?;`)
	if err != nil {
		return nil, err
	}

	stmtModelFreq, err := db.Prepare(`SELECT coalesce(SUM(frequency), 0) FROM markov_chains WHERE model_id = ?;`)
	if err != nil {
		return nil, err
	}

	stmtInsertLink, err := db.Prepare(`INSERT INTO markov_chains (model_id, prefix_id, next_token_id) VALUES (?, ?, ?) ON CONFLICT DO UPDATE SET frequency = frequency + 1;`)
	if err != nil {
		return nil, err
	}

	stmtGetTokenID, err := db.Prepare(`SELECT token_id FROM markov_vocabulary WHERE token_text = ?;`)
	if err != nil {
		return nil, err
	}

	stmtGetPrefixID, err := db.Prepare(`SELECT prefix_id FROM markov_prefixes WHERE prefix_text = ?;`)
	if err != nil {
		return nil, err
	}

	stmtGetTokenText, err := db.Prepare(`SELECT token_text FROM markov_vocabulary WHERE token_id = ?;`)
	if err != nil {
		return nil, err
	}

	stmtGetChain, err := db.Prepare(`SELECT next_token_id, frequency FROM markov_chains WHERE model_id = ? AND prefix_id = ?;`)
	if err != nil {
		return nil, err
	}

	stmtGetVocabLen, err := db.Prepare(`SELECT COUNT(*) FROM markov_vocabulary;`)
	if err != nil {
		return nil, err
	}

	stmtGetPrefixLen, err := db.Prepare(`SELECT COUNT(*) FROM markov_prefixes;`)
	if err != nil {
		return nil, err
	}

	stmtInsertVocab, err := db.Prepare(`INSERT INTO markov_vocabulary (token_text) VALUES (?) ON CONFLICT(token_text) DO UPDATE SET token_text=excluded.token_text RETURNING token_id;`)
	if err != nil {
		return nil, err
	}

	stmtGetOrInsertPrefix, err := db.Prepare(`INSERT INTO markov_prefixes (prefix_text) VALUES (?) ON CONFLICT(prefix_text) DO UPDATE SET prefix_text=excluded.prefix_text RETURNING prefix_id;`)
	if err != nil {
		return nil, err
	}

	return &Generator{
		db:                    db,
		tokenizer:             tokenizer,
		stmtGetModelInfo:      stmtGetModelInfo,
		stmtGetModels:         stmtGetModels,
		stmtAddModel:          stmtAddModel,
		stmtPruneModel:        stmtPruneModel,
		stmtModelChains:       stmtModelChains,
		stmtModelStarters:     stmtModelStarters,
		stmtModelFreq:         stmtModelFreq,
		stmtInsertLink:        stmtInsertLink,
		stmtGetTokenID:        stmtGetTokenID,
		stmtGetPrefixID:       stmtGetPrefixID,
		stmtGetTokenText:      stmtGetTokenText,
		stmtGetChain:          stmtGetChain,
		stmtGetVocabLen:       stmtGetVocabLen,
		stmtGetPrefixLen:      stmtGetPrefixLen,
		stmtInsertVocab:       stmtInsertVocab,
		stmtGetOrInsertPrefix: stmtGetOrInsertPrefix,
		logger:                slog.New(slog.NewTextHandler(io.Discard, nil)),
	}, nil
}

// Close releases all prepared SQL statements held by the Generator. It should be
// called when the Generator is no longer needed to free up database resources.
func (g *Generator) Close() {
	_ = g.stmtGetModelInfo.Close()
	_ = g.stmtGetModels.Close()
	_ = g.stmtAddModel.Close()
	_ = g.stmtPruneModel.Close()
	_ = g.stmtModelChains.Close()
	_ = g.stmtModelStarters.Close()
	_ = g.stmtModelFreq.Close()
	_ = g.stmtInsertLink.Close()
	_ = g.stmtGetTokenID.Close()
	_ = g.stmtGetPrefixID.Close()
	_ = g.stmtGetTokenText.Close()
	_ = g.stmtGetChain.Close()
	_ = g.stmtGetVocabLen.Close()
}

// SetLogger sets the logger for the Generator. By default, all logs are discarded.
// Providing a `log/slog.Logger` will enable logging for training, generation,
// and other operations.
func (g *Generator) SetLogger(logger *slog.Logger) {
	if logger != nil {
		g.logger = logger
	}
}
