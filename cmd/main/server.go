package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/CTAG07/Sarracenia/pkg/markov"
	"github.com/CTAG07/Sarracenia/pkg/templating"
)

type TemplateInput struct {
	ThreatLevel int
	ThreatStage int
}

type Server struct {
	config            *Config
	configMux         *sync.RWMutex
	markovDB          *sql.DB
	authDB            *sql.DB
	statsDB           *sql.DB
	logger            *slog.Logger
	mg                *markov.Generator
	tm                *templating.TemplateManager
	tc                *ThreatCalculator
	wlc               *WhitelistCache
	authAPI           *AuthAPI
	templateAPI       *TemplateAPI
	markovAPI         *MarkovAPI
	statsAPI          *StatsAPI
	serverAPI         *ServerAPI
	whitelistAPI      *WhitelistAPI
	tarpitMux         *http.ServeMux
	apiMux            *http.ServeMux
	dashboardTemplate *template.Template
}

func NewServer(config *Config, logger *slog.Logger, markovDB *sql.DB, authDB *sql.DB, statsDB *sql.DB, actionChan chan string) (*Server, error) {

	// markov initialization
	mg, err := markov.NewGenerator(markovDB, markov.NewDefaultTokenizer())
	if err != nil {
		return nil, fmt.Errorf("error creating markov generator: %v", err)
	}

	tc := NewThreatCalculator(config.Threat, logger)

	tm, err := templating.NewTemplateManager(logger, mg, config.Templates, "./data")
	if err != nil {
		return nil, fmt.Errorf("failed to create template manager: %w", err)
	}

	wlc := NewWhitelistCache()
	err = wlc.LoadFromDB(authDB)
	if err != nil {
		return nil, fmt.Errorf("failed to load whitelist from db: %w", err)
	}

	configMux := &sync.RWMutex{}

	// api initialization
	authAPI := NewAuthAPI(authDB, logger)
	templateAPI := NewTemplateAPI(tm, tc, logger)
	markovAPI := NewMarkovAPI(mg, tm, logger)
	statsAPI := NewStatsAPI(statsDB, logger)
	serverAPI := NewServerAPI(config, configMux, actionChan, tm, logger)
	whitelistAPI := NewWhitelistAPI(authDB, logger, wlc)

	// initialize the stats cache with configuration
	if err = statsAPI.InitializeCache(config.Server.StatsConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize stats cache: %w", err)
	}

	// create object, register routes to the mux, and return it
	server := &Server{
		config:       config,
		configMux:    configMux,
		markovDB:     markovDB,
		authDB:       authDB,
		statsDB:      statsDB,
		logger:       logger,
		tm:           tm,
		tc:           tc,
		mg:           mg,
		wlc:          wlc,
		authAPI:      authAPI,
		templateAPI:  templateAPI,
		markovAPI:    markovAPI,
		statsAPI:     statsAPI,
		serverAPI:    serverAPI,
		whitelistAPI: whitelistAPI,
		tarpitMux:    http.NewServeMux(),
		apiMux:       http.NewServeMux(),
	}

	apiMux := http.NewServeMux()

	server.authAPI.RegisterRoutes(apiMux)
	server.templateAPI.RegisterRoutes(apiMux)
	server.markovAPI.RegisterRoutes(apiMux)
	server.statsAPI.RegisterRoutes(apiMux)
	server.serverAPI.RegisterRoutes(apiMux)
	server.whitelistAPI.RegisterRoutes(apiMux)

	// Make sure api functions must pass through authentication first
	authedAPI := server.authAPI.Authenticate(apiMux)
	// ... except for the health check, which is unauthed so something like docker can use it
	server.apiMux.HandleFunc("/api/health", server.serverAPI.handleHealthCheck)

	server.apiMux.Handle("/api/", authedAPI)

	staticFs := http.FileServer(http.Dir(config.Server.DashboardStaticPath))
	server.apiMux.Handle("/static/", http.StripPrefix("/static/", staticFs))

	server.dashboardTemplate, err = template.ParseGlob(filepath.Join(config.Server.DashboardTmplPath, "*.gohtml"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse dashboard template: %w", err)
	}

	server.apiMux.HandleFunc("/", server.handleDashboard)
	server.tarpitMux.HandleFunc("/favicon.ico", handleFavicon)
	server.tarpitMux.HandleFunc("/", server.handleTarpit)

	return server, nil
}

// handleDashboard is the dedicated handler for rendering the main dashboard page.
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Simple check to avoid serving the template for non-root paths like /favicon.ico
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := s.dashboardTemplate.ExecuteTemplate(w, "index.gohtml", nil); err != nil {
		s.logger.Error("Failed to render dashboard template", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) handleTarpit(w http.ResponseWriter, r *http.Request) {
	ipAddr := getClientIP(r)
	if s.wlc.IsWhitelisted(ipAddr, r.UserAgent()) {
		s.logger.Debug("Request from whitelisted client, serving 404.", "remote_addr", r.RemoteAddr, "user_agent", r.UserAgent())
		http.NotFound(w, r)
		return
	}
	metrics, err := s.statsAPI.LogAndGetMetrics(r)
	if err != nil {
		s.logger.Warn("Failed to log and get metrics, proceeding with default threat assessment", "error", err)
		metrics = &RequestMetrics{
			IPAddress: ipAddr,
			UserAgent: r.UserAgent(),
			// Everything else default (0)
		}
	}
	threatLevel := s.tc.GetThreatLevel(metrics)
	threatState := s.tc.GetStage(threatLevel)

	s.configMux.RLock()
	enabledTemplates := make([]string, len(s.config.Server.EnabledTemplates))
	copy(enabledTemplates, s.config.Server.EnabledTemplates)
	tarpitConfig := *s.config.Server.TarpitConfig
	s.configMux.RUnlock()

	var templateName string
	if len(enabledTemplates) > 0 {
		templateName = enabledTemplates[rand.Intn(len(enabledTemplates))]
	} else {
		templateName = s.tm.GetRandomTemplate()
	}
	s.logger.Info(
		"Serving tarpit page",
		"template", templateName,
		"remote_addr", ipAddr,
		"Threat_level", threatLevel,
		"Threat_state", threatState)

	var buf bytes.Buffer
	err = s.tm.Execute(&buf, templateName, TemplateInput{ThreatLevel: threatLevel, ThreatStage: threatState})
	if err != nil {
		s.logger.Error("Failed to execute template", "template", templateName, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	s.setTarpitHeaders(w)

	// If drip feeding is disabled in the config, or any of the config is invalid, send the response normally.
	if !tarpitConfig.EnableDripFeed || tarpitConfig.DripFeedChunksMax <= 0 || tarpitConfig.DripFeedDelayMax < 0 || tarpitConfig.DripFeedChunksMax < 0 {
		_, _ = buf.WriteTo(w)
		return
	}

	// Enforce an initial delay before any data is sent.
	if tarpitConfig.InitialDelayMax > 0 {
		time.Sleep(time.Duration(randRangeMinZero(tarpitConfig.InitialDelayMin, tarpitConfig.InitialDelayMax)) * time.Millisecond)
	}

	// Assert that the ResponseWriter supports flushing.
	flusher, ok := w.(http.Flusher)
	if !ok {
		s.logger.Warn("ResponseWriter does not support flushing, sending response at once.")
		_, _ = buf.WriteTo(w)
		return
	}

	responseBytes := buf.Bytes()
	totalSize := len(responseBytes)
	chunks := randRangeMinZero(tarpitConfig.DripFeedChunksMin, tarpitConfig.DripFeedChunksMax)
	chunkSize := totalSize / chunks

	// Ensure chunk size is at least 1 to avoid an infinite loop on small responses.
	if chunkSize <= 0 {
		chunkSize = 1
	}

	// Loop through the response and write it chunk by chunk.
	for i := 0; i < totalSize; i += chunkSize {
		end := i + chunkSize
		if end > totalSize {
			end = totalSize
		}

		// Write the chunk to the client.
		if _, err = w.Write(responseBytes[i:end]); err != nil {
			s.logger.Error("Failed to write tarpit chunk to client", "error", err, "remote_addr", r.RemoteAddr)
			return // Stop if the client closes the connection.
		}

		// Flush the writer to ensure the chunk is sent over the network immediately.
		flusher.Flush()

		// Wait before sending the next chunk, but not after the last one.
		if end < totalSize {
			time.Sleep(time.Duration(randRangeMinZero(tarpitConfig.DripFeedDelayMin, tarpitConfig.DripFeedDelayMax)) * time.Millisecond)
		}
	}
}

// I don't want to write all this out twice, I'm sorry.
func randRangeMinZero(min, max int) int {
	if min < 0 {
		min = 0
	}
	return rand.Intn(max-min) + min
}

func (s *Server) setTarpitHeaders(w http.ResponseWriter) {

	w.Header().Set("Cache-Control", "no-store, no-cache")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func getClientIP(r *http.Request) string {
	// First check X-Real-IP header (usually set by proxies)
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Then check X-Forwarded-For header
	// The leftmost IP is usually the original client when using trusted proxies
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				return ip
			}
		}
	}

	// Cloudflare also adds this header
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}

	// As a last resort, fall back to the remote address
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If splitting fails (e.g., no port), return the address as is.
		return r.RemoteAddr
	}
	return ip
}

// handleFavicon is a small function to make sure that favicon requests aren't tarpitted, and instead return no
// content. This prevents double-counting of requests if a favicon is requested.
func handleFavicon(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
