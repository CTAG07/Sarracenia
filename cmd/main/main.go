package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/CTAG07/Sarracenia/pkg/markov"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

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

	config, err := LoadConfig("./config.json")
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
