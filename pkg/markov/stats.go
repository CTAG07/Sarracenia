package markov

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

// DBStats holds aggregated statistics for the entire database, including a
// list of all models and their individual stats.
type DBStats struct {
	Models     []ModelInfo        // A list of models in the database
	Stats      map[int]ModelStats // A mapping of model ids to their stats
	VocabSize  int                // The number of unique tokens in all models' vocabularies
	PrefixSize int                // The number of unique prefixes in all models' chains
}

// ModelStats holds aggregated statistics for a single Markov model.
type ModelStats struct {
	TotalChains    int // The number of unique prefix->next_token links.
	TotalFrequency int // The sum of frequencies of all links; the total number of trained transitions.
	StartingTokens int // The number of unique tokens that can start a chain.
}

// GetStats returns a snapshot of statistics for the entire database,
// including global counts and per-model stats.
func (g *Generator) GetStats(ctx context.Context) (*DBStats, error) {
	modelInfos, err := g.GetModelInfos(ctx)
	if err != nil {
		return nil, err
	}

	var vocabLen int
	err = g.stmtGetVocabLen.QueryRowContext(ctx).Scan(&vocabLen)
	if err != nil {
		return nil, err
	}

	var prefixLen int
	err = g.stmtGetPrefixLen.QueryRowContext(ctx).Scan(&prefixLen)
	if err != nil {
		return nil, err
	}

	models := make([]ModelInfo, 0)
	modelStats := make(map[int]ModelStats)
	for _, v := range modelInfos {
		models = append(models, v)
		var totalChains, totalFrequency, startingTokens int
		err = g.stmtModelChains.QueryRowContext(ctx, v.Id).Scan(&totalChains)
		if err != nil {
			return nil, err
		}
		err = g.stmtModelFreq.QueryRowContext(ctx, v.Id).Scan(&totalFrequency)
		if err != nil {
			return nil, err
		}
		chain := make([]string, v.Order)
		socStr := strconv.Itoa(0) // SOC
		for i := range chain {
			chain[i] = socStr
		}
		var socId int
		err = g.stmtGetPrefixID.QueryRowContext(ctx, strings.Join(chain, " ")).Scan(&socId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				startingTokens = 0
			} else {
				return nil, err
			}
		} else {
			err = g.stmtModelStarters.QueryRowContext(ctx, v.Id, socId).Scan(&startingTokens)
			if err != nil {
				return nil, err
			}
		}
		modelStats[v.Id] = ModelStats{
			TotalChains:    totalChains,
			TotalFrequency: totalFrequency,
			StartingTokens: startingTokens,
		}
	}

	return &DBStats{
		Models:     models,
		Stats:      modelStats,
		VocabSize:  vocabLen,
		PrefixSize: prefixLen,
	}, nil
}
