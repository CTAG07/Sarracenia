package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const statsSchema = `
CREATE TABLE IF NOT EXISTS stats_ip (
    ip_address    TEXT PRIMARY KEY,
    total_hits    INTEGER NOT NULL DEFAULT 1,
    first_seen    DATETIME NOT NULL,
    last_seen     DATETIME NOT NULL
);
CREATE TABLE IF NOT EXISTS stats_user_agent (
    user_agent    TEXT PRIMARY KEY,
    total_hits    INTEGER NOT NULL DEFAULT 1,
    first_seen    DATETIME NOT NULL,
    last_seen     DATETIME NOT NULL
);
`

// RequestMetrics is the data structure returned for a single request.
type RequestMetrics struct {
	IPAddress        string        `json:"ip_address"`
	UserAgent        string        `json:"user_agent"`
	IPTotalHits      int           `json:"ip_total_hits"`
	UATotalHits      int           `json:"ua_total_hits"`
	TimeSinceIPFirst time.Duration `json:"time_since_ip_first_seen"`
	TimeSinceUAFirst time.Duration `json:"time_since_ua_first_seen"`
}

// GlobalStatsSummary provides a high-level overview of all collected stats.
type GlobalStatsSummary struct {
	TotalRequests    int64 `json:"total_requests"`
	UniqueIPs        int64 `json:"unique_ips"`
	UniqueUserAgents int64 `json:"unique_user_agents"`
}

// StatsAPI holds the dependencies for the statistics handlers.
type StatsAPI struct {
	db     *sql.DB
	logger *slog.Logger
}

func setupStatsSchema(db *sql.DB) error {
	_, err := db.Exec(statsSchema)
	return err
}

func NewStatsAPI(db *sql.DB, logger *slog.Logger) *StatsAPI {
	return &StatsAPI{
		db:     db,
		logger: logger,
	}
}

func (s *StatsAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/stats/summary", s.handleSummary)
	mux.HandleFunc("/api/stats/top_ips", s.handleTopIPs)
	mux.HandleFunc("/api/stats/top_user_agents", s.handleTopUserAgents)
}

// LogAndGetMetrics is the core function called by the tarpit handler.
// It logs the request and returns up-to-date metrics for it in a single transaction.
func (s *StatsAPI) LogAndGetMetrics(r *http.Request) (*RequestMetrics, error) {
	ip := r.RemoteAddr
	ua := r.UserAgent()
	now := time.Now()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	_, err = tx.ExecContext(r.Context(), `
        INSERT INTO stats_ip (ip_address, first_seen, last_seen) VALUES (?, ?, ?)
        ON CONFLICT(ip_address) DO UPDATE SET total_hits = total_hits + 1, last_seen = ?
    `, ip, now, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert stats_ip: %w", err)
	}

	_, err = tx.ExecContext(r.Context(), `
        INSERT INTO stats_user_agent (user_agent, first_seen, last_seen) VALUES (?, ?, ?)
        ON CONFLICT(user_agent) DO UPDATE SET total_hits = total_hits + 1, last_seen = ?
    `, ua, now, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert stats_user_agent: %w", err)
	}

	// 3. Retrieve the updated metrics for the response
	metrics := &RequestMetrics{IPAddress: ip, UserAgent: ua}
	var ipFirstSeen, uaFirstSeen time.Time

	err = tx.QueryRowContext(r.Context(), "SELECT total_hits, first_seen FROM stats_ip WHERE ip_address = ?", ip).Scan(&metrics.IPTotalHits, &ipFirstSeen)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated stats_ip: %w", err)
	}

	err = tx.QueryRowContext(r.Context(), "SELECT total_hits, first_seen FROM stats_user_agent WHERE user_agent = ?", ua).Scan(&metrics.UATotalHits, &uaFirstSeen)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated stats_user_agent: %w", err)
	}

	metrics.TimeSinceIPFirst = now.Sub(ipFirstSeen)
	metrics.TimeSinceUAFirst = now.Sub(uaFirstSeen)

	// 4. Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit stats transaction: %w", err)
	}

	return metrics, nil
}

func (s *StatsAPI) handleSummary(w http.ResponseWriter, r *http.Request) {
	if !hasScope(r, "stats:read") {
		respondWithError(w, http.StatusForbidden, "Forbidden")
		return
	}
	var summary GlobalStatsSummary
	// Calculate total requests by summing all hits from the IP stats table.
	_ = s.db.QueryRowContext(r.Context(), "SELECT COALESCE(SUM(total_hits), 0) FROM stats_ip").Scan(&summary.TotalRequests)
	_ = s.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM stats_ip").Scan(&summary.UniqueIPs)
	_ = s.db.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM stats_user_agent").Scan(&summary.UniqueUserAgents)
	respondWithJSON(w, http.StatusOK, summary)
}

func (s *StatsAPI) handleTopIPs(w http.ResponseWriter, r *http.Request) {
	if !hasScope(r, "stats:read") {
		respondWithError(w, http.StatusForbidden, "Forbidden")
		return
	}
	rows, err := s.db.QueryContext(r.Context(), "SELECT ip_address, total_hits, first_seen, last_seen FROM stats_ip ORDER BY total_hits DESC LIMIT 100")
	if err != nil {
		s.logger.Error("Failed to query top IPs", "error", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		return
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var results []map[string]any
	for rows.Next() {
		var ip string
		var hits int
		var first, last time.Time
		err = rows.Scan(&ip, &hits, &first, &last)
		if err != nil {
			s.logger.Error("Failed to scan top IPs", "error", err)
		}
		results = append(results, map[string]any{
			"ip_address": ip,
			"total_hits": hits,
			"first_seen": first,
			"last_seen":  last,
		})
	}
	respondWithJSON(w, http.StatusOK, results)
}

func (s *StatsAPI) handleTopUserAgents(w http.ResponseWriter, r *http.Request) {
	if !hasScope(r, "stats:read") {
		respondWithError(w, http.StatusForbidden, "Forbidden")
		return
	}
	rows, err := s.db.QueryContext(r.Context(), "SELECT user_agent, total_hits, first_seen, last_seen FROM stats_user_agent ORDER BY total_hits DESC LIMIT 100")
	if err != nil {
		s.logger.Error("Failed to query top UAs", "error", err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
		return
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var results []map[string]any
	for rows.Next() {
		var ua string
		var hits int
		var first, last time.Time
		err = rows.Scan(&ua, &hits, &first, &last)
		if err != nil {
			s.logger.Error("Failed to scan top UAs", "error", err)
		}
		results = append(results, map[string]any{
			"user_agent": ua,
			"total_hits": hits,
			"first_seen": first,
			"last_seen":  last,
		})
	}
	respondWithJSON(w, http.StatusOK, results)
}
