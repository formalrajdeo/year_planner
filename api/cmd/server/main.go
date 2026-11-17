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

	"yearplanner/internal/auth"
	"yearplanner/internal/config"
	httpapi "yearplanner/internal/http"
	"yearplanner/internal/store"
	"yearplanner/internal/users"
	"yearplanner/internal/workspaces"

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
	workspaceRepository := workspaces.NewMongoRepository(database.Collection("workspaces"))
	sessionRepository := store.NewMongoSessionStore(database.Collection("sessions"))
	loginLimiter := store.NewMongoLoginRateLimiter(database.Collection("auth_rate_limits"), cfg.SessionSecret)
	resetTokenStore := store.NewMongoPasswordResetStore(database.Collection("password_reset_tokens"))
	indexCtx, cancelIndexes := context.WithTimeout(context.Background(), cfg.MongoTimeout)
	defer cancelIndexes()
	if err := userRepository.EnsureIndexes(indexCtx); err != nil {
		return fmt.Errorf("create user indexes: %w", err)
	}
	if err := workspaceRepository.EnsureIndexes(indexCtx); err != nil {
		return fmt.Errorf("create workspace indexes: %w", err)
	}
	if err := sessionRepository.EnsureIndexes(indexCtx); err != nil {
		return fmt.Errorf("create session indexes: %w", err)
	}
	if err := loginLimiter.EnsureIndexes(indexCtx); err != nil {
		return fmt.Errorf("create auth rate limit indexes: %w", err)
	}
	if err := resetTokenStore.EnsureIndexes(indexCtx); err != nil {
		return fmt.Errorf("create password reset indexes: %w", err)
	}

	var emailSender auth.PasswordResetEmailSender
	var mongoEmailQueue *store.MongoPasswordResetEmailQueue
	var emailQueue auth.PasswordResetEmailQueue
	if cfg.SMTPHost != "" {
		emailSender = auth.NewSMTPEmailSender(auth.SMTPConfig{
			Host:       cfg.SMTPHost,
			Port:       cfg.SMTPPort,
			Username:   cfg.SMTPUsername,
			Password:   cfg.SMTPPassword,
			From:       cfg.SMTPFrom,
			RequireTLS: cfg.AppEnv == "production",
		})
		mongoEmailQueue = store.NewMongoPasswordResetEmailQueue(
			database.Collection("password_reset_email_queue"),
			cfg.SessionSecret,
		)
		if err := mongoEmailQueue.EnsureIndexes(indexCtx); err != nil {
			return fmt.Errorf("create password reset email queue indexes: %w", err)
		}
		emailQueue = mongoEmailQueue
	}
	server := httpapi.NewServerWithWorkspaceRepository(
		cfg,
		userRepository,
		sessionRepository,
		database,
		loginLimiter,
		resetTokenStore,
		emailQueue,
		workspaceRepository,
	)
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
	var emailWorkerDone chan struct{}
	workerCtx, cancelEmailWorker := context.WithCancel(context.Background())
	defer cancelEmailWorker()
	if mongoEmailQueue != nil {
		emailWorkerDone = make(chan struct{})
		go func() {
			defer close(emailWorkerDone)
			mongoEmailQueue.Run(workerCtx, emailSender)
		}()
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	slog.Info("API server started", "port", cfg.AppPort)

	select {
	case <-stop:
	case err := <-serverErrors:
		cancelEmailWorker()
		return fmt.Errorf("HTTP server failed: %w", err)
	}

	cancelEmailWorker()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful HTTP shutdown failed: %w", err)
	}
	if emailWorkerDone != nil {
		select {
		case <-emailWorkerDone:
		case <-ctx.Done():
			return fmt.Errorf("password reset email worker shutdown timed out")
		}
	}
	return nil
}
