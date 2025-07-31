package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CTAG07/Sarracenia/pkg/markov"
	"github.com/CTAG07/Sarracenia/pkg/templating"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type TemplateInput struct {
	ThreatLevel int
	ThreatStage int
}

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

type ServerConfig struct {
	ServerAddr    string `json:"server_addr"`
	ApiAddr       string `json:"api_addr"`
	LogLevel      string `json:"log_level"`
	DataDir       string `json:"data_dir"`
	DatabasePath  string `json:"database_path"`
	DashboardPath string `json:"dashboard_path"`
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		ServerAddr:    ":7277",
		ApiAddr:       ":7278",
		LogLevel:      "info",
		DataDir:       "./data",
		DatabasePath:  "./data/sarracenia.db",
		DashboardPath: "./data/dashboard.html",
	}
}

type Config struct {
	Server    *ServerConfig              `json:"server_config"`
	Templates *templating.TemplateConfig `json:"template_config"`
	Threat    *ThreatConfig              `json:"threat_config"`
}

type Server struct {
	config       *Config
	db           *sql.DB
	logger       *slog.Logger
	mg           *markov.Generator
	tm           *templating.TemplateManager
	tc           *ThreatCalculator
	wlc          *WhitelistCache
	authAPI      *AuthAPI
	templateAPI  *TemplateAPI
	markovAPI    *MarkovAPI
	statsAPI     *StatsAPI
	serverAPI    *ServerAPI
	whitelistAPI *WhitelistAPI
	tarpitMux    *http.ServeMux
	apiMux       *http.ServeMux
}

func NewServer(config *Config, logger *slog.Logger, db *sql.DB, actionChan chan string) (*Server, error) {

	// markov initialization
	mg, err := markov.NewGenerator(db, markov.NewDefaultTokenizer())
	if err != nil {
		return nil, fmt.Errorf("error creating markov generator: %v", err)
	}

	tc := NewThreatCalculator(config.Threat, logger)

	tm, err := templating.NewTemplateManager(logger, mg, config.Templates, "./data")
	if err != nil {
		return nil, fmt.Errorf("failed to create template manager: %w", err)
	}

	wlc := NewWhitelistCache()
	err = wlc.LoadFromDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to load whitelist from db: %w", err)
	}

	// api initialization
	authAPI := NewAuthAPI(db, logger)
	templateAPI := NewTemplateAPI(tm, tc, logger)
	markovAPI := NewMarkovAPI(mg, tm, logger)
	statsAPI := NewStatsAPI(db, logger)
	serverAPI := NewServerAPI(config, actionChan, tm, logger)
	whitelistAPI := NewWhitelistAPI(db, logger, wlc)

	// create object, register routes to the mux, and return it
	server := &Server{
		config:       config,
		db:           db,
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

	server.apiMux.Handle("/api/", authedAPI)
	server.apiMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, config.Server.DashboardPath)
	})
	server.tarpitMux.HandleFunc("/", server.handleTarpit)

	return server, nil
}

func (s *Server) handleTarpit(w http.ResponseWriter, r *http.Request) {
	if s.wlc.IsWhitelisted(r.RemoteAddr, r.UserAgent()) {
		s.logger.Debug("Request from whitelisted client, serving 404.", "remote_addr", r.RemoteAddr, "user_agent", r.UserAgent())
		http.NotFound(w, r)
		return
	}
	metrics, err := s.statsAPI.LogAndGetMetrics(r)
	threatLevel := s.tc.GetThreatLevel(metrics)
	threatState := s.tc.GetStage(threatLevel)
	if err != nil {
		s.logger.Error("Failed to get stats / Threat", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	templateName := s.tm.GetRandomTemplate()
	s.logger.Info(
		"Serving tarpit page",
		"template", templateName,
		"remote_addr", r.RemoteAddr,
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
	_, _ = buf.WriteTo(w)
}

func (s *Server) setTarpitHeaders(w http.ResponseWriter) {

	w.Header().Set("Cache-Control", "no-store, no-cache")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func loadConfig(path string) (*Config, error) {
	config := &Config{
		Server:    DefaultServerConfig(),
		Templates: templating.DefaultConfig(),
		Threat:    DefaultThreatConfig(),
	}

	file, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, create it with defaults
			data, _ := json.MarshalIndent(config, "", "  ")
			_ = os.WriteFile(path, data, 0644)
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err = json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	return config, nil
}

func main() {
	baseLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	actionChan := make(chan string, 1)

	go func() {
		osSignalChan := make(chan os.Signal, 1)
		signal.Notify(osSignalChan, syscall.SIGINT, syscall.SIGTERM)
		<-osSignalChan // Wait for a signal
		baseLogger.Info("OS signal received, initiating shutdown.")
		actionChan <- actionShutdown
	}()

	for {
		action, err := run(actionChan)
		if err != nil {
			baseLogger.Error("An error occurred during server run, shutting down.", "error", err)
			break
		}

		if action == actionRestart {
			baseLogger.Info("--- Server Restarting ---")
			continue
		} else {
			break
		}
	}

	baseLogger.Info("Sarracenia has shut down.")
}

// run is the main loop that hosts both servers, and returns whenever the server is shutdown or restarted
func run(actionChan chan string) (string, error) {

	config, err := loadConfig("./config.json")
	if err != nil {
		return "", fmt.Errorf("failed to load configuration: %w", err)
	}

	var logLevel slog.Level
	switch strings.ToLower(config.Server.LogLevel) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	logger.Info("Starting server cycle...")

	db, err := initDB(config.Server.DatabasePath)
	if err != nil {
		return "", fmt.Errorf("failed to initialize database: %w", err)
	}

	if err = markov.SetupSchema(db); err != nil {
		logger.Error("Failed to setup markov schema", "error", err)
	}
	if err = setupAuthSchema(db); err != nil {
		logger.Error("Failed to setup auth schema", "error", err)
	}
	if err = setupStatsSchema(db); err != nil {
		logger.Error("Failed to setup stats schema", "error", err)
	}
	if err = setupWhitelistSchema(db); err != nil {
		logger.Error("Failed to setup whitelist schema", "error", err)
	}

	tarpitHttpServer := &http.Server{Addr: config.Server.ServerAddr}
	apiHttpServer := &http.Server{Addr: config.Server.ApiAddr}

	server, err := NewServer(config, logger, db, actionChan)
	if err != nil {
		_ = db.Close()
		return "", fmt.Errorf("failed to create server object: %w", err)
	}

	tarpitHttpServer.Handler = server.tarpitMux
	apiHttpServer.Handler = server.apiMux

	go func() {
		logger.Info("Starting api/dashboard server", "address", apiHttpServer.Addr)
		if err := apiHttpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Api server failed", "error", err)
		}
	}()

	go func() {
		logger.Info("Starting Sarracenia tarpit server", "address", tarpitHttpServer.Addr)
		if err := tarpitHttpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Tarpit server failed", "error", err)
		}
	}()

	action := <-actionChan // Block here until API or OS signal sends an action.

	logger.Info("Stopping servers for " + action + "...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = apiHttpServer.Shutdown(ctx); err != nil {
		logger.Error("Api server shutdown failed", "error", err)
	}
	if err = tarpitHttpServer.Shutdown(ctx); err != nil {
		logger.Error("Tarpit server shutdown failed", "error", err)
	}
	logger.Info("HTTP servers stopped.")

	logger.Info("Closing database connection.")
	if err = db.Close(); err != nil {
		logger.Error("Failed to close database", "error", err)
	}

	return action, nil
}
