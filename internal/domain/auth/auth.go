package auth

import (
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rayhan889/neatspace/internal/application/services"
	authRepo "github.com/rayhan889/neatspace/internal/domain/auth/repositories"
	"github.com/rayhan889/neatspace/internal/notification"
)

type Options struct {
	PgPool      *pgxpool.Pool                 // PostgreSQL connection pool (required)
	UserService services.UserServiceInterface // User service (optional)
	Logger      *slog.Logger                  // Slog logger instance (optional)
	Mailer      *notification.Mailer          // Mailer service (optional)
	BaseURL     string                        // Base URL for constructing links (required)
	Port        int                           // Server port (required)
	Host        string                        // Server host (required)
}

type AuthDomain struct {
	authService *services.AuthService
}

func NewAuthDomain(opts *Options) *AuthDomain {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	}

	authService := services.NewAuthService(services.AuthServiceOpts{
		AuthRepository: authRepo.NewAuthRepository(opts.PgPool, logger),
		UserService:    opts.UserService,
		Logger:         logger,
		Mailer:         opts.Mailer,
		BaseURL:        opts.BaseURL,
		Port:           opts.Port,
		Host:           opts.Host,
	})

	return &AuthDomain{
		authService: authService,
	}
}

func (d *AuthDomain) GetAuthService() *services.AuthService {
	return d.authService
}
