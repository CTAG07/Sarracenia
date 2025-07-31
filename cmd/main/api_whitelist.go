package main

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

// WhitelistAPI manages the whitelists for IPs and User Agents.
type WhitelistAPI struct {
	db     *sql.DB
	logger *slog.Logger
	cache  *WhitelistCache // A pointer to the in-memory cache
}

// setupWhitelistSchema creates the table for storing whitelisted IPs and User Agents.
func setupWhitelistSchema(db *sql.DB) error {
	const schema = `
	CREATE TABLE IF NOT EXISTS whitelist (
		id INTEGER PRIMARY KEY,
		type TEXT NOT NULL CHECK(type IN ('ip', 'user_agent')),
		value TEXT NOT NULL UNIQUE
	);
	`
	_, err := db.Exec(schema)
	return err
}

type WhitelistCache struct {
	mu                 sync.RWMutex
	ipWhitelist        map[string]struct{}
	userAgentWhitelist map[string]struct{}
}

func NewWhitelistCache() *WhitelistCache {
	return &WhitelistCache{
		ipWhitelist:        make(map[string]struct{}),
		userAgentWhitelist: make(map[string]struct{}),
	}
}

// LoadFromDB reads all whitelist entries from the database into the cache.
func (c *WhitelistCache) LoadFromDB(db *sql.DB) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ipWhitelist = make(map[string]struct{})
	c.userAgentWhitelist = make(map[string]struct{})

	rows, err := db.Query("SELECT type, value FROM whitelist")
	if err != nil {
		return err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var listType, value string
		if err = rows.Scan(&listType, &value); err != nil {
			return err
		}
		if listType == "ip" {
			c.ipWhitelist[value] = struct{}{}
		} else if listType == "user_agent" {
			c.userAgentWhitelist[value] = struct{}{}
		}
	}
	return rows.Err()
}

// Add safely adds a single entry to the cache.
func (c *WhitelistCache) Add(listType, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if listType == "ip" {
		c.ipWhitelist[value] = struct{}{}
	} else if listType == "user_agent" {
		c.userAgentWhitelist[value] = struct{}{}
	}
}

// Remove safely removes a single entry from the cache.
func (c *WhitelistCache) Remove(listType, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if listType == "ip" {
		delete(c.ipWhitelist, value)
	} else if listType == "user_agent" {
		delete(c.userAgentWhitelist, value)
	}
}

// IsWhitelisted safely checks if an IP or User Agent is in the cache.
func (c *WhitelistCache) IsWhitelisted(ip, userAgent string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if _, found := c.ipWhitelist[ip]; found {
		return true
	}
	if _, found := c.userAgentWhitelist[userAgent]; found {
		return true
	}
	return false
}

// NewWhitelistAPI creates a new instance of the WhitelistAPI.
func NewWhitelistAPI(db *sql.DB, logger *slog.Logger, cache *WhitelistCache) *WhitelistAPI {
	return &WhitelistAPI{
		db:     db,
		logger: logger,
		cache:  cache,
	}
}

// RegisterRoutes sets up the routing for all /api/whitelist endpoints.
func (a *WhitelistAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/whitelist/ip", a.handleWhitelist("ip"))
	mux.HandleFunc("/api/whitelist/useragent", a.handleWhitelist("user_agent"))
}

// handleWhitelist is a generic handler for both IP and User Agent whitelists.
func (a *WhitelistAPI) handleWhitelist(listType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			a.getList(w, r, listType)
		case http.MethodPost:
			a.addToList(w, r, listType)
		case http.MethodDelete:
			a.removeFromList(w, r, listType)
		default:
			w.Header().Set("Allow", "GET, POST, DELETE")
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	}
}

// getList retrieves all entries for a given whitelist type.
func (a *WhitelistAPI) getList(w http.ResponseWriter, r *http.Request, listType string) {
	if !hasScope(r, "whitelist:read") {
		respondWithError(w, http.StatusForbidden, "Forbidden: requires 'whitelist:read' scope")
		return
	}

	rows, err := a.db.Query("SELECT value FROM whitelist WHERE type = ?", listType)
	if err != nil {
		a.logger.Error("Failed to query whitelist", "type", listType, "error", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve whitelist")
		return
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var values []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			a.logger.Error("Failed to scan whitelist value", "error", err)
			continue
		}
		values = append(values, value)
	}

	respondWithJSON(w, http.StatusOK, values)
}

// addToList adds a new value to the specified whitelist.
func (a *WhitelistAPI) addToList(w http.ResponseWriter, r *http.Request, listType string) {
	if !hasScope(r, "whitelist:write") {
		respondWithError(w, http.StatusForbidden, "Forbidden: requires 'whitelist:write' scope")
		return
	}

	var payload struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}
	value := strings.TrimSpace(payload.Value)
	if value == "" {
		respondWithError(w, http.StatusBadRequest, "Whitelist value cannot be empty")
		return
	}

	_, err := a.db.Exec("INSERT INTO whitelist (type, value) VALUES (?, ?)", listType, value)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			respondWithError(w, http.StatusConflict, "Value already exists in the whitelist")
		} else {
			a.logger.Error("Failed to insert into whitelist", "type", listType, "value", value, "error", err)
			respondWithError(w, http.StatusInternalServerError, "Failed to add value to whitelist")
		}
		return
	}

	a.cache.Add(listType, value)
	a.logger.Info("Added value to whitelist", "type", listType, "value", value)
	respondWithJSON(w, http.StatusCreated, map[string]string{"message": "Value added to whitelist"})
}

// removeFromList removes a value from the specified whitelist.
func (a *WhitelistAPI) removeFromList(w http.ResponseWriter, r *http.Request, listType string) {
	if !hasScope(r, "whitelist:write") {
		respondWithError(w, http.StatusForbidden, "Forbidden: requires 'whitelist:write' scope")
		return
	}

	var payload struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}
	value := strings.TrimSpace(payload.Value)
	if value == "" {
		respondWithError(w, http.StatusBadRequest, "Whitelist value cannot be empty")
		return
	}

	res, err := a.db.Exec("DELETE FROM whitelist WHERE type = ? AND value = ?", listType, value)
	if err != nil {
		a.logger.Error("Failed to delete from whitelist", "type", listType, "value", value, "error", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to remove value from whitelist")
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		respondWithError(w, http.StatusNotFound, "Value not found in the whitelist")
		return
	}

	a.cache.Remove(listType, value)
	a.logger.Info("Removed value from whitelist", "type", listType, "value", value)
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Value removed from whitelist"})
}
