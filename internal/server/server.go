package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/rayhan889/neatspace/internal/application"
	"github.com/rayhan889/neatspace/internal/application/handler"
	"github.com/rayhan889/neatspace/internal/application/middlewares"
	"github.com/rayhan889/neatspace/internal/config"
	"github.com/rayhan889/neatspace/internal/infrasturcture/database"
	"github.com/rayhan889/neatspace/internal/notification"
	templateFS "github.com/rayhan889/neatspace/templates"
)

type HTTPServer struct {
	httpAddr string
	logger   *slog.Logger
}

func NewHTTPServer(httpAddr string, logger *slog.Logger) *HTTPServer {
	return &HTTPServer{
		httpAddr: httpAddr,
		logger:   logger,
	}
}

func (s *HTTPServer) Run() error {
	cfg := config.Get()

	pg, err := s.initializeDatabase(cfg)
	if err != nil {
		s.logger.Error("Failed to connect on Postgres database", "err", err)
		os.Exit(1)
	}

	defer pg.Close()

	var mailer *notification.Mailer
	s.logger.Info("Initializing SMTP mailer service")
	m, err := notification.NewMailer(notification.MailerOpts{
		SMTPHost:     cfg.Mailer.SMTPHost,
		SMTPPort:     cfg.Mailer.SMTPPort,
		SMTPUsername: cfg.Mailer.SMTPUsername,
		SMTPPassword: cfg.Mailer.SMTPPassword,
		FromName:     cfg.Mailer.SenderName,
		FromAddr:     cfg.Mailer.SenderEmail,
		TemplateFS:   templateFS.TemplateDir,
		Logger:       s.logger,
	})
	if err != nil {
		s.logger.Info("Mailer service not configured or failed to initialize, continuing without mailer", "err", err)
	} else {
		mailer = m
		s.logger.Info("Mailer service initialized", "host", cfg.Mailer.SMTPHost, "port", cfg.Mailer.SMTPPort)
	}

	fiberApp := fiber.New(fiber.Config{
		CaseSensitive:     true,
		StrictRouting:     true,
		AppName:           fmt.Sprintf("Interns neatspace Backend App %s", application.Version),
		EnablePrintRoutes: true,
		ErrorHandler:      handler.Error,
	})

	fiberApp.Use(logger.New())

	fiberApp.Use(middlewares.SecurityHeadersMiddleware())
	fiberApp.Use(middlewares.LoggerMiddleware(s.logger))

	// Initializing application modules
	if err := s.initializeApplication(cfg, pg.Pool, mailer, fiberApp); err != nil {
		s.logger.Error("Failed to initialize application", "err", err)
		return err
	}

	s.logger.Info("Starting HTTP server", "addr", s.httpAddr)

	// Gracefully shutting down handler
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start server in background
	serverErrCh := make(chan error, 1)
	go func() {
		if err := fiberApp.Listen(s.httpAddr); err != nil && err != http.ErrServerClosed {
			serverErrCh <- fmt.Errorf("server start failed: %w", err)
		}
		// If server ended without error, notify
		close(serverErrCh)
	}()

	// Wait for signal or server start error
	select {
	case <-ctx.Done():
		// received shutdown signal
		s.logger.Info("Shutdown signal received, shutting down HTTP server")
	case err := <-serverErrCh:
		if err != nil {
			// server failed to start or crashed
			s.logger.Error("HTTP server error", "err", err)
			// proceed to shutdown resources anyway
		}
	}

	// Shutdown timeout set to 10s
	shutdownTimeout := 10 * time.Second
	context, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown HTTP server gracefully
	if err := fiberApp.ShutdownWithContext(context); err != nil {
		s.logger.Error("failed to shutdown HTTP server gracefully", "err", err)
	}

	// Close DB pool
	s.logger.Info("Closing database connections")
	pg.Close()

	s.logger.Info("Shutdown complete")
	return nil
}

func (s *HTTPServer) initializeDatabase(cfg *config.Config) (*database.PostgresDB, error) {
	const baseDelay = 2 * time.Second
	const maxDelay = 30 * time.Second
	const defaultMaxRetries = 5

	maxRetries := defaultMaxRetries
	if cfg.Database.MaxRetries == -1 {
		maxRetries = -1
	} else if cfg.Database.MaxRetries > 0 {
		maxRetries = cfg.Database.MaxRetries
	}

	s.logger.Info("Initializing database connection", "max_retries", maxRetries)

	var lastErr error
	attempt := 1

	for {
		pgCfg := database.PostgresConfig{URL: cfg.GetDatabaseURL()}
		pg, err := database.NewPostgres(pgCfg)
		if err == nil {
			// verify connection with Ping and timeout
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			pingErr := pg.Ping(ctx)
			cancel()

			if pingErr == nil {
				s.logger.Info("Database connection established", "attempt", attempt)
				return pg, nil
			}

			// ping failed, close pool and treat as error
			pg.Close()
			err = fmt.Errorf("ping failed: %w", pingErr)
		}

		lastErr = err
		s.logger.Warn("Database connection attempt failed", "attempt", attempt, "err", lastErr)

		// If not infinite and we've reached max, stop retrying
		if maxRetries != -1 && attempt >= maxRetries {
			break
		}

		// Exponential-ish backoff bounded by maxDelay
		delay := min(baseDelay*time.Duration(attempt), maxDelay)

		s.logger.Info("Retrying database connection", "next_try_in", delay, "attempt", attempt+1)
		time.Sleep(delay)
		attempt++
	}

	return nil, fmt.Errorf("failed to establish database connection after %d attempts: %w", attempt, lastErr)
}
