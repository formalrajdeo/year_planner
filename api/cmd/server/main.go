package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"yearplanner/internal/config"
	httpapi "yearplanner/internal/http"
	"yearplanner/internal/store"
	"yearplanner/internal/users"

	"github.com/joho/godotenv"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load environment file: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), cfg.MongoTimeout)
	database, err := store.ConnectMongo(startupCtx, cfg)
	cancelStartup()
	if err != nil {
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := database.Disconnect(ctx); err != nil {
			slog.Error("MongoDB disconnect failed", "error", err)
		}
	}()

	userRepository := users.NewMongoRepository(database.Collection("users"))
	sessionRepository := store.NewMongoSessionStore(database.Collection("sessions"))
	indexCtx, cancelIndexes := context.WithTimeout(context.Background(), cfg.MongoTimeout)
	defer cancelIndexes()
	if err := userRepository.EnsureIndexes(indexCtx); err != nil {
		return fmt.Errorf("create user indexes: %w", err)
	}
	if err := sessionRepository.EnsureIndexes(indexCtx); err != nil {
		return fmt.Errorf("create session indexes: %w", err)
	}

	server := httpapi.NewServerWithDependencies(cfg, userRepository, sessionRepository, database)
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.AppPort),
		Handler:           server.Router(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	serverErrors := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	slog.Info("API server started", "port", cfg.AppPort)

	select {
	case <-stop:
	case err := <-serverErrors:
		return fmt.Errorf("HTTP server failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful HTTP shutdown failed: %w", err)
	}
	return nil
}
