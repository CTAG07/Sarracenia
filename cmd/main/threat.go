package main

import (
	"log/slog"
	"math"
)

// StageConfig defines the parameters for a single Threat stage.
type StageConfig struct {
	Enabled   bool `json:"enabled"`
	Threshold int  `json:"threshold"`
}

// ThreatStages holds the configuration for the 5 discrete Threat stages.
type ThreatStages struct {
	Stage0 StageConfig `json:"stage_1"`
	Stage1 StageConfig `json:"stage_2"`
	Stage2 StageConfig `json:"stage_3"`
	Stage3 StageConfig `json:"stage_4"`
	Stage4 StageConfig `json:"stage_5"`
}

// ThreatConfig holds all parameters for calculating the Threat score.
// This allows an administrator to fine-tune how aggressively the tarpit should
// respond to different patterns of client behavior.
type ThreatConfig struct {
	// BaseThreat is the starting score for any request.
	BaseThreat int `json:"base_threat"`

	// IPHitFactor is the value added to the score for each total hit from a given IP.
	IPHitFactor float64 `json:"ip_hit_factor"`

	// UAHitFactor is the value added to the score for each total hit from a given User Agent.
	UAHitFactor float64 `json:"ua_hit_factor"`

	// IPHitRateFactor determines how much weight is given to the frequency of requests
	// from an IP (hits per minute). A high value strongly penalizes rapid requests.
	IPHitRateFactor float64 `json:"ip_hit_rate_factor"`

	// UAHitRateFactor determines how much weight is given to the frequency of requests
	// from a User Agent. This can help identify distributed botnets using the same UA.
	UAHitRateFactor float64 `json:"ua_hit_rate_factor"`

	// MaxThreat is the absolute ceiling for the Threat score to prevent runaway values.
	MaxThreat int `json:"max_threat"`

	// FallbackLevel is the default Threat level (0-4) to use if an incoming
	// request's score does not meet any enabled stage thresholds.
	FallbackLevel int `json:"fallback_level"`

	// Stages defines the score thresholds for each of the 5 Threat levels.
	Stages ThreatStages `json:"stages"`
}

// DefaultThreatConfig returns a new ThreatConfig with the Threat system disabled by default.
func DefaultThreatConfig() *ThreatConfig {
	return &ThreatConfig{
		BaseThreat:      0,
		IPHitFactor:     1.0,
		UAHitFactor:     0.5,
		IPHitRateFactor: 10.0,
		UAHitRateFactor: 5.0,
		MaxThreat:       1000,
		FallbackLevel:   0, // Default to the least aggressive level.
		Stages: ThreatStages{
			// Stage 0 is always enabled with a threshold of 0.
			Stage0: StageConfig{Enabled: true, Threshold: 0},
			// Subsequent stages are disabled by default, effectively disabling the Threat system
			Stage1: StageConfig{Enabled: false, Threshold: 25},
			Stage2: StageConfig{Enabled: false, Threshold: 50},
			Stage3: StageConfig{Enabled: false, Threshold: 75},
			Stage4: StageConfig{Enabled: false, Threshold: 100},
		},
	}
}

// ThreatCalculator is responsible for turning request metrics into a quantifiable
// Threat score and level.
type ThreatCalculator struct {
	config *ThreatConfig
	logger *slog.Logger
}

// NewThreatCalculator creates a new calculator with the given configuration.
func NewThreatCalculator(config *ThreatConfig, logger *slog.Logger) *ThreatCalculator {
	return &ThreatCalculator{
		config: config,
		logger: logger,
	}
}

// GetThreatLevel calculates a raw threat score based on the provided request metrics
// and the configured weights and factors.
func (c *ThreatCalculator) GetThreatLevel(metrics *RequestMetrics) int {
	score := float64(c.config.BaseThreat)

	score += float64(metrics.IPTotalHits) * c.config.IPHitFactor
	score += float64(metrics.UATotalHits) * c.config.UAHitFactor

	if metrics.IPTotalHits > 1 {
		ipMinutes := math.Max(metrics.TimeSinceIPFirst.Minutes(), 1.0/60.0)
		ipRate := float64(metrics.IPTotalHits) / ipMinutes
		score += ipRate * c.config.IPHitRateFactor
	}

	if metrics.UATotalHits > 1 {
		uaMinutes := math.Max(metrics.TimeSinceUAFirst.Minutes(), 1.0/60.0)
		uaRate := float64(metrics.UATotalHits) / uaMinutes
		score += uaRate * c.config.UAHitRateFactor
	}

	finalScore := int(score)
	if finalScore > c.config.MaxThreat {
		finalScore = c.config.MaxThreat
	} else if finalScore < 0 {
		finalScore = 0
	}

	c.logger.Debug("Calculated threat",
		"ip", metrics.IPAddress,
		"user_agent", metrics.UserAgent,
		"raw_score", score,
		"final_score", finalScore,
	)

	return finalScore
}

// GetStage maps a raw threat level to a discrete stage from 0-4.
// It iterates from the highest stage (4) to the lowest, respecting the Enabled
// flag for each stage. If no enabled stage threshold is met, it returns the
// configured FallbackLevel.
func (c *ThreatCalculator) GetStage(threatLevel int) int {
	// Create a temporary slice for easy iteration from highest to lowest level.
	stages := []StageConfig{
		c.config.Stages.Stage4, // Level 4
		c.config.Stages.Stage3, // Level 3
		c.config.Stages.Stage2, // Level 2
		c.config.Stages.Stage1, // Level 1
		c.config.Stages.Stage0, // Level 0
	}

	// Check from the most aggressive stage downwards.
	for i, stage := range stages {
		level := 4 - i // Calculate the corresponding level (4, 3, 2, 1, 0)
		if stage.Enabled && threatLevel >= stage.Threshold {
			return level // Return the highest applicable level.
		}
	}

	// If no enabled stages were matched, return the fallback.
	// We clamp the value to ensure it's always within the valid 0-4 range.
	return max(0, min(c.config.FallbackLevel, 4))
}
