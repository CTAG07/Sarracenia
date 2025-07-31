package main

import (
	"encoding/json"
	"fmt"
	"github.com/CTAG07/Sarracenia/pkg/templating"
	"log/slog"
	"net/http"
	"os"
)

const (
	actionShutdown = "shutdown"
	actionRestart  = "restart"
)

// ServerAPI holds the dependencies for the main application API handlers.
type ServerAPI struct {
	config     *Config
	actionChan chan string
	tm         *templating.TemplateManager
	logger     *slog.Logger
}

// VersionInfo defines the structure for build/version information.
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
}

// NewServerAPI creates a new instance of the ServerAPI.
func NewServerAPI(config *Config, actionChan chan string, tm *templating.TemplateManager, logger *slog.Logger) *ServerAPI {
	return &ServerAPI{
		config:     config,
		actionChan: actionChan,
		tm:         tm,
		logger:     logger,
	}
}

// RegisterRoutes sets up the routing for all /api/server endpoints.
func (a *ServerAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/server/config", a.handleConfig)
	mux.HandleFunc("/api/server/version", a.handleVersion)
	mux.HandleFunc("/api/server/shutdown", a.handleShutdown)
	mux.HandleFunc("/api/server/restart", a.handleRestart)
}

// handleConfig gets or updates the main server configuration.
func (a *ServerAPI) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if !hasScope(r, "server:config") {
			respondWithError(w, http.StatusForbidden, "Forbidden: requires 'server:config' scope")
			return
		}
		respondWithJSON(w, http.StatusOK, a.config)
	case http.MethodPut:
		if !hasScope(r, "server:config") {
			respondWithError(w, http.StatusForbidden, "Forbidden: requires 'server:config' scope")
			return
		}
		var newConfig Config
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid JSON request body")
			return
		}

		// Update the live config object
		*a.config = newConfig
		a.tm.SetConfig(newConfig.Templates)
		_ = a.tm.Refresh()

		// Persist the changes to disk
		data, err := json.MarshalIndent(a.config, "", "  ")
		if err != nil {
			a.logger.Error("Failed to marshal new config for saving", "error", err)
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to prepare config for saving: %v", err))
			return
		}
		if err = os.WriteFile("config.json", data, 0644); err != nil {
			a.logger.Error("Failed to save config.json", "error", err)
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to save configuration to disk: %v", err))
			return
		}

		a.logger.Info("Application configuration updated and saved via API. Some changes may require a restart.")
		respondWithJSON(w, http.StatusOK, a.config)
	default:
		w.Header().Set("Allow", "GET, PUT")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleVersion returns the application's build information.
func (a *ServerAPI) handleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	// This endpoint is often left open or given a broad scope for simple diagnostics.
	// We'll still protect it for consistency.
	if !hasScope(r, "stats:read") { // Re-using stats:read scope is reasonable here.
		respondWithError(w, http.StatusForbidden, "Forbidden: requires 'stats:read' scope")
		return
	}

	info := VersionInfo{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
	}
	respondWithJSON(w, http.StatusOK, info)
}

// handleShutdown initiates a graceful shutdown of the server.
func (a *ServerAPI) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if !hasScope(r, "server:control") {
		respondWithError(w, http.StatusForbidden, "Forbidden: requires 'server:control' scope")
		return
	}

	a.logger.Warn("Shutdown initiated via API")
	respondWithJSON(w, http.StatusAccepted, map[string]string{"message": "Server is shutting down..."})

	go func() {
		a.actionChan <- actionShutdown
	}()
}

// handleShutdown initiates a graceful restart of the server.
func (a *ServerAPI) handleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if !hasScope(r, "server:control") {
		respondWithError(w, http.StatusForbidden, "Forbidden: requires 'server:control' scope")
		return
	}

	a.logger.Warn("Restart initiated via API")
	respondWithJSON(w, http.StatusAccepted, map[string]string{"message": "Server is restarting..."})

	go func() {
		a.actionChan <- actionRestart
	}()
}
