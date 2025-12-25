package main

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CTAG07/Sarracenia/pkg/templating"
)

// TemplateAPI holds the dependencies for the template API handlers.
type TemplateAPI struct {
	tm     *templating.TemplateManager
	tc     *ThreatCalculator
	logger *slog.Logger
}

// NewTemplateAPI creates a new instance of the TemplateAPI.
func NewTemplateAPI(tm *templating.TemplateManager, tc *ThreatCalculator, logger *slog.Logger) *TemplateAPI {
	return &TemplateAPI{
		tm:     tm,
		tc:     tc,
		logger: logger,
	}
}

// RegisterRoutes sets up the routing for all /api/templates endpoints.
func (t *TemplateAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/templates/refresh", t.handleRefresh)
	mux.HandleFunc("/api/templates/test", t.handleTest)
	mux.HandleFunc("/api/templates/preview", t.handlePreview) // New preview endpoint
	mux.HandleFunc("/api/templates", t.handleList)
	mux.HandleFunc("/api/templates/", t.handleFile)
}

// handleRefresh triggers a manual refresh of templates from disk.
func (t *TemplateAPI) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if !hasScope(r, "templates:write") {
		respondWithError(w, http.StatusForbidden, "Forbidden: requires 'templates:write' scope")
		return
	}
	if err := t.tm.Refresh(); err != nil {
		t.logger.Error("API triggered refresh failed", "error", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to refresh templates: %v", err))
		return
	}
	t.logger.Info("Templates refreshed via API")
	w.WriteHeader(http.StatusNoContent)
}

// handleList returns a list of all available template names.
func (t *TemplateAPI) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if !hasScope(r, "templates:read") {
		respondWithError(w, http.StatusForbidden, "Forbidden: requires 'templates:read' scope")
		return
	}
	respondWithJSON(w, http.StatusOK, t.tm.GetTemplateNames())
}

// handleTest validates template syntax without saving the file by executing it as a string.
func (t *TemplateAPI) handleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if !hasScope(r, "templates:read") {
		respondWithError(w, http.StatusForbidden, "Forbidden: requires 'templates:read' scope")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to read request body: %v", err))
		return
	}

	threat, err := strconv.Atoi(r.URL.Query().Get("threat"))
	if err != nil {
		threat = 0
	}

	var buf bytes.Buffer
	err = t.tm.ExecuteTemplateString(&buf, string(body), TemplateInput{threat, t.tc.GetStage(threat)})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Template execution failed: %v", err))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

// handlePreview renders a template with temporarily overridden "threat" levels.
func (t *TemplateAPI) handlePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if !hasScope(r, "templates:read") {
		respondWithError(w, http.StatusForbidden, "Forbidden: requires 'templates:read' scope")
		return
	}

	name := r.URL.Query().Get("name")
	threat, err := strconv.Atoi(r.URL.Query().Get("threat"))
	if err != nil {
		threat = 0
	}
	if name == "" {
		respondWithError(w, http.StatusBadRequest, "Query parameter 'name' is required")
		return
	}

	var buf bytes.Buffer
	if err = t.tm.Execute(&buf, name, TemplateInput{threat, t.tc.GetStage(threat)}); err != nil {
		if strings.Contains(err.Error(), "is undefined") {
			respondWithError(w, http.StatusNotFound, fmt.Sprintf("Template '%s' not found", name))
			return
		}
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to render preview: %v", err))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
}

// handleFile manages CRUD operations for a single template file.
func (t *TemplateAPI) handleFile(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/templates/")
	if name == "" || strings.HasSuffix(name, "/") {
		respondWithError(w, http.StatusNotFound, "Not Found")
		return
	}

	if strings.Contains(name, "..") || (!strings.HasSuffix(name, ".tmpl.html") && !strings.HasSuffix(name, ".part.html")) {
		respondWithError(w, http.StatusBadRequest, "Invalid template name format")
		return
	}

	templateDir, err := filepath.Abs(t.tm.GetTemplateDir())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to resolve template directory")
		return
	}

	path := filepath.Join(templateDir, name)
	absPath, err := filepath.Abs(path)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	if !strings.HasPrefix(absPath, templateDir) {
		respondWithError(w, http.StatusForbidden, "Access denied: Path outside template directory")
		return
	}

	switch r.Method {
	case http.MethodGet:
		if !hasScope(r, "templates:read") {
			respondWithError(w, http.StatusForbidden, "Forbidden: requires 'templates:read' scope")
			return
		}
		content, err := os.ReadFile(path)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Template not found")
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write(content)

	case http.MethodPut:
		if !hasScope(r, "templates:write") {
			respondWithError(w, http.StatusForbidden, "Forbidden: requires 'templates:write' scope")
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to read request body: %v", err))
			return
		}
		if err = os.WriteFile(path, body, 0644); err != nil {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to write template file: %v", err))
			return
		}
		_ = t.tm.Refresh()
		w.WriteHeader(http.StatusNoContent)

	case http.MethodDelete:
		if !hasScope(r, "templates:write") {
			respondWithError(w, http.StatusForbidden, "Forbidden: requires 'templates:write' scope")
			return
		}
		if err := os.Remove(path); err != nil {
			if os.IsNotExist(err) {
				respondWithError(w, http.StatusNotFound, "Template not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete template file: %v", err))
			return
		}
		_ = t.tm.Refresh()
		w.WriteHeader(http.StatusNoContent)

	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
