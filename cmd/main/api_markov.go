package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/CTAG07/Sarracenia/pkg/templating"

	"github.com/CTAG07/Sarracenia/pkg/markov"
)

// MarkovAPI holds the dependencies for the Markov model API handlers.
type MarkovAPI struct {
	gen                    *markov.Generator
	tm                     *templating.TemplateManager
	logger                 *slog.Logger
	trainingMux            sync.Mutex
	infoMux                sync.RWMutex
	isTraining             bool
	currentlyTrainingModel string
}

// NewMarkovAPI creates a new instance of the MarkovAPI.
func NewMarkovAPI(gen *markov.Generator, tm *templating.TemplateManager, logger *slog.Logger) *MarkovAPI {
	return &MarkovAPI{
		gen:                    gen,
		tm:                     tm,
		logger:                 logger,
		trainingMux:            sync.Mutex{},
		infoMux:                sync.RWMutex{},
		isTraining:             false,
		currentlyTrainingModel: "",
	}
}

// RegisterRoutes sets up the routing for all /api/markov endpoints.
func (m *MarkovAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/markov/models", m.handleListAndCreateModels)
	mux.HandleFunc("/api/markov/models/", m.handleModelByName)
	mux.HandleFunc("/api/markov/import", m.handleImport)
	mux.HandleFunc("/api/markov/vocabulary/prune", m.handleVocabPrune)
	mux.HandleFunc("/api/markov/training/status", m.handleTrainingStatus)
}

type CreateModelRequest struct {
	Name  string `json:"name"`
	Order int    `json:"order"`
}

type PruneRequest struct {
	MinFreq int `json:"minFreq"`
}

type GenerateRequest struct {
	MaxLength   int     `json:"maxLength"`
	Temperature float64 `json:"temperature"`
	TopK        int     `json:"topK"`
	StartText   string  `json:"startText"`
}

// handleListAndCreateModels handles GET for listing and POST for creating models.
func (m *MarkovAPI) handleListAndCreateModels(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if !hasScope(r, "markov:read") {
			respondWithError(w, http.StatusForbidden, "Forbidden: requires 'markov:read' scope")
			return
		}
		models, err := m.gen.GetModelInfos(r.Context())
		if err != nil {
			m.logger.Error("Failed to get model infos", "error", err)
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve models: %v", err))
			return
		}
		// Convert map to slice for consistent JSON output
		modelList := make([]markov.ModelInfo, 0, len(models))
		for _, model := range models {
			modelList = append(modelList, model)
		}
		respondWithJSON(w, http.StatusOK, modelList)

	case http.MethodPost:
		if !hasScope(r, "markov:write") {
			respondWithError(w, http.StatusForbidden, "Forbidden: requires 'markov:write' scope")
			return
		}
		var req CreateModelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid JSON request body")
			return
		}
		if req.Name == "" || req.Order <= 0 {
			respondWithError(w, http.StatusBadRequest, "Model name and a positive order are required")
			return
		}

		model := markov.ModelInfo{Name: req.Name, Order: req.Order}
		if err := m.gen.InsertModel(r.Context(), model); err != nil {
			m.logger.Error("Failed to insert new model", "name", req.Name, "error", err)
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create model: %v", err))
			return
		}
		newModel, err := m.gen.GetModelInfo(r.Context(), req.Name)
		if err != nil {
			m.logger.Error("Failed to retrieve newly created model", "name", req.Name, "error", err)
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to verify model creation: %v", err))
			return
		}
		_ = m.tm.Refresh()
		respondWithJSON(w, http.StatusCreated, newModel)
	default:
		w.Header().Set("Allow", "GET, POST")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleModelByName routes actions for a specific model, e.g., train, prune, export, delete.
func (m *MarkovAPI) handleModelByName(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/api/markov/models/")
	parts := strings.Split(path, "/")
	modelName := parts[0]

	if modelName == "" {
		respondWithError(w, http.StatusBadRequest, "Model name not specified")
		return
	}

	model, err := m.gen.GetModelInfo(r.Context(), modelName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Model not found")
			return
		}
		m.logger.Error("Failed to get model info by name", "name", modelName, "error", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		return
	}

	if len(parts) == 1 { // Path is just /api/markov/models/{name}
		if r.Method == http.MethodDelete {
			if !hasScope(r, "markov:write") {
				respondWithError(w, http.StatusForbidden, "Forbidden: requires 'markov:write' scope")
				return
			}
			if err = m.gen.RemoveModel(r.Context(), model); err != nil {
				m.logger.Error("Failed to remove model", "name", modelName, "error", err)
				respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to remove model: %v", err))
				return
			}
			_ = m.tm.Refresh()
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.Header().Set("Allow", "DELETE")
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	action := parts[1]
	switch action {
	case "train":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		if !hasScope(r, "markov:write") {
			respondWithError(w, http.StatusForbidden, "Forbidden: requires 'markov:write' scope")
			return
		}

		tempDir := filepath.Join("./data", "tmp")
		if err = os.MkdirAll(tempDir, 0755); err != nil {
			m.logger.Error("Could not create temp directory for training", "error", err)
			respondWithError(w, http.StatusInternalServerError, "Internal server error: cannot create temp dir")
			return
		}

		var tempFile *os.File
		tempFile, err = os.CreateTemp(tempDir, "sarracenia-corpus-*.txt")
		if err != nil {
			m.logger.Error("Could not create temp file for training", "error", err)
			respondWithError(w, http.StatusInternalServerError, "Internal server error: cannot create temp file")
			return
		}

		_, err = io.Copy(tempFile, r.Body)
		if err != nil {
			_ = tempFile.Close()
			_ = os.Remove(tempFile.Name())
			m.logger.Error("Could not write corpus to temp file", "error", err)
			respondWithError(w, http.StatusInternalServerError, "Internal server error: failed to write corpus")
			return
		}
		_ = tempFile.Close()

		go m.runTrainingJob(modelName, tempFile.Name())
		w.WriteHeader(http.StatusAccepted)

	case "prune":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		if !hasScope(r, "markov:write") {
			respondWithError(w, http.StatusForbidden, "Forbidden: requires 'markov:write' scope")
			return
		}
		var req PruneRequest
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid JSON request body")
			return
		}
		if err = m.gen.PruneModel(r.Context(), model, req.MinFreq); err != nil {
			m.logger.Error("Failed to prune model", "name", modelName, "error", err)
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Pruning failed: %v", err))
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case "export":
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		if !hasScope(r, "markov:read") {
			respondWithError(w, http.StatusForbidden, "Forbidden: requires 'markov:read' scope")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.json\"", modelName))
		if err = m.gen.ExportModel(r.Context(), model, w); err != nil {
			m.logger.Error("Failed to export model", "name", modelName, "error", err)
		}

	case "generate":
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		if !hasScope(r, "markov:read") {
			respondWithError(w, http.StatusForbidden, "Forbidden: requires 'markov:read' scope")
			return
		}

		var req GenerateRequest
		if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid JSON request body")
			return
		}

		genOpts := []markov.GenerateOption{
			markov.WithMaxLength(req.MaxLength),
			markov.WithTemperature(req.Temperature),
			markov.WithTopK(req.TopK),
			markov.WithEarlyTermination(true),
		}

		var generatedText string
		generatedText, err = m.gen.GenerateFromString(r.Context(), model, req.StartText, genOpts...)
		if err != nil {
			m.logger.Error("Failed to generate text from model", "name", modelName, "error", err)
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Generation failed: %v", err))
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"text": generatedText})

	default:
		respondWithError(w, http.StatusNotFound, "Action not found")
	}
}

// handleImport imports a model from an uploaded JSON file.
func (m *MarkovAPI) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if !hasScope(r, "markov:write") {
		respondWithError(w, http.StatusForbidden, "Forbidden: requires 'markov:write' scope")
		return
	}

	if err := m.gen.ImportModel(r.Context(), r.Body); err != nil {
		m.logger.Error("Failed to import model", "error", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Import failed: %v", err))
		return
	}

	_ = m.tm.Refresh()
	w.WriteHeader(http.StatusAccepted)
}

// handleVocabPrune performs a global vocabulary prune.
func (m *MarkovAPI) handleVocabPrune(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if !hasScope(r, "markov:write") {
		respondWithError(w, http.StatusForbidden, "Forbidden: requires 'markov:write' scope")
		return
	}
	var req PruneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON request body for minFrequency")
		return
	}
	if err := m.gen.VocabularyPrune(r.Context(), req.MinFreq); err != nil {
		m.logger.Error("Failed to prune vocabulary", "error", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Vocabulary prune failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleTrainingStatus checks if a model training job is currently in progress.
func (m *MarkovAPI) handleTrainingStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if !hasScope(r, "markov:read") {
		respondWithError(w, http.StatusForbidden, "Forbidden: requires 'markov:read' scope")
		return
	}

	m.infoMux.RLock()
	defer m.infoMux.RUnlock()

	response := map[string]interface{}{
		"is_training": m.isTraining,
		"model_name":  m.currentlyTrainingModel,
	}
	respondWithJSON(w, http.StatusOK, response)
}

func (m *MarkovAPI) runTrainingJob(modelName, tempFileName string) {

	// Get training lock (only train one model at once)
	m.trainingMux.Lock()
	defer m.trainingMux.Unlock()

	m.infoMux.Lock()
	m.currentlyTrainingModel = modelName
	m.isTraining = true
	m.infoMux.Unlock()

	defer func() {
		m.infoMux.Lock()
		m.currentlyTrainingModel = ""
		m.isTraining = false
		m.infoMux.Unlock()
	}()

	tempFile, err := os.Open(tempFileName)
	if err != nil {
		m.logger.Error("Training job failed: could not re-open corpus file for reading", "file", tempFileName, "error", err)
		_ = os.Remove(tempFileName)
		return
	}

	// Defer closing the file then deleting it
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFileName)
	}()

	ctx := context.Background()

	modelInfo, err := m.gen.GetModelInfo(ctx, modelName)
	if err != nil {
		m.logger.Error("Training job failed: could not get model info", "model", modelName, "error", err)
		return
	}

	err = m.gen.Train(ctx, modelInfo, tempFile)
	if err != nil {
		m.logger.Error("Training job failed during training", "model", modelName, "error", err)
	} else {
		m.logger.Info("Training job completed successfully", "model", modelName)
	}
}
