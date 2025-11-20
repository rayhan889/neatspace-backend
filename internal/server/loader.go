package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rayhan889/intern-payroll/internal/application/handler"
	"github.com/rayhan889/intern-payroll/internal/application/middlewares"
	"github.com/rayhan889/intern-payroll/internal/config"
	authDomain "github.com/rayhan889/intern-payroll/internal/domain/auth"
	userDomain "github.com/rayhan889/intern-payroll/internal/domain/user"
	"github.com/rayhan889/intern-payroll/internal/notification"
)

// Initialize application modules : containing services, repositories, etc.
func (s *HTTPServer) initializeApplication(cfg *config.Config, pgPool *pgxpool.Pool, mailer *notification.Mailer, fiberApp *fiber.App) error {
	fiberApp.Use(middlewares.CORSMiddleware(cfg))
	fiberApp.Use(middlewares.RateLimitMiddleware(
		cfg.App.RateLimitRequests, cfg.App.RateLimitBurstSize,
	))

	// Create api v1 group
	apiV1Route := fiberApp.Group("/api/v1")

	// Load domain application
	userDomain := userDomain.NewUserDomain(&userDomain.Options{
		PgPool: pgPool,
		Logger: s.logger,
	})
	authDomain := authDomain.NewAuthDomain(&authDomain.Options{
		PgPool:      pgPool,
		UserService: userDomain.GetUserService(),
		Logger:      s.logger,
		Mailer:      mailer,
		BaseURL:     cfg.GetAppBaseURL(),
		Port:        cfg.GetPort(),
		Host:        cfg.GetHost(),
	})

	handler.NewAuthHandler(handler.AuthHandlerOpts{
		RouteGroup:  apiV1Route,
		AuthService: authDomain.GetAuthService(),
	})
	handler.NewUserHandler(handler.UserHandlerOpts{
		RouteGroup:  apiV1Route,
		UserService: userDomain.GetUserService(),
	})

	// Register main application routes
	serverHandler := handler.NewServerHandler(pgPool, s.logger)
	serverHandler.RegisterRoutes(fiberApp)

	return nil
}
