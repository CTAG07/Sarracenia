package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/CTAG07/Sarracenia/pkg/templating"
	"github.com/natefinch/atomic"
)

// ServerConfig holds the configuration for the HTTP servers.
type ServerConfig struct {
	ServerAddr          string        `json:"server_addr"`
	ApiAddr             string        `json:"api_addr"`
	LogLevel            string        `json:"log_level"`
	DataDir             string        `json:"data_dir"`
	DatabasePath        string        `json:"database_path"`
	DashboardTmplPath   string        `json:"dashboard_tmpl_path"`
	DashboardStaticPath string        `json:"dashboard_static_path"`
	EnabledTemplates    []string      `json:"enabled_templates"`
	TarpitConfig        *TarpitConfig `json:"tarpit_config"`
}

// TarpitConfig holds settings for response delaying and drip-feeding.
type TarpitConfig struct {
	EnableDripFeed    bool `json:"enable_drip_feed"`
	InitialDelayMin   int  `json:"min_initial_delay_ms"`
	InitialDelayMax   int  `json:"max_initial_delay_ms"`
	DripFeedDelayMin  int  `json:"min_drip_feed_delay_ms"`
	DripFeedDelayMax  int  `json:"max_drip_feed_delay_ms"`
	DripFeedChunksMin int  `json:"min_drip_feed_chunks"`
	DripFeedChunksMax int  `json:"max_drip_feed_chunks"`
}

// Config is the top-level configuration struct that aggregates all other configs.
type Config struct {
	Server    *ServerConfig              `json:"server_config"`
	Templates *templating.TemplateConfig `json:"template_config"`
	Threat    *ThreatConfig              `json:"threat_config"`
}

// DefaultServerConfig creates a server configuration with default values.
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		ServerAddr:          ":7277",
		ApiAddr:             ":7278",
		LogLevel:            "info",
		DataDir:             "./data",
		DatabasePath:        "./data/sarracenia.db?_journal_mode=WAL",
		DashboardTmplPath:   "./data/dashboard/templates/",
		DashboardStaticPath: "./data/dashboard/static/",
		EnabledTemplates:    []string{"page.tmpl.html"},
		TarpitConfig: &TarpitConfig{
			EnableDripFeed:    false,
			InitialDelayMin:   0,
			InitialDelayMax:   15000,
			DripFeedDelayMin:  500,
			DripFeedDelayMax:  1000,
			DripFeedChunksMin: 1,
			DripFeedChunksMax: 20,
		},
	}
}

// LoadConfig reads the configuration from a JSON file at the given path.
// If the file doesn't exist, it creates one with default values.
func LoadConfig(path string) (*Config, error) {
	// Initialize with default configurations
	config := &Config{
		Server:    DefaultServerConfig(),
		Templates: templating.DefaultConfig(),
		Threat:    DefaultThreatConfig(),
	}

	file, err := os.ReadFile(path)
	if err != nil {
		// If the file doesn't exist, create it with the default config.
		if os.IsNotExist(err) {
			var data []byte
			data, err = json.MarshalIndent(config, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal default config: %w", err)
			}
			if err = atomic.WriteFile(path, bytes.NewReader(data)); err != nil {
				// Log a warning instead of failing, as the server can still run with defaults.
				fmt.Printf("warning: failed to write default config file: %v\n", err)
			}
			return config, nil
		}
		// For other errors (e.g., permission denied), return the error.
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal the JSON from the file into the config struct.
	if err = json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}
