package user

import (
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rayhan889/neatspace/internal/application/services"
	"github.com/rayhan889/neatspace/internal/domain/user/repositories"
)

type Options struct {
	PgPool *pgxpool.Pool // PostgreSQL connection pool (required)
	Logger *slog.Logger  // Slog logger instance (optional)
}

type UserDomain struct {
	logger      *slog.Logger
	userService *services.UserService
}

func NewUserDomain(opts *Options) *UserDomain {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	}

	userService := services.NewUserService(services.UserServiceOpts{
		UserRepo: repositories.NewUserRepository(opts.PgPool, logger),
	})

	return &UserDomain{
		logger:      logger,
		userService: userService,
	}
}

func (d *UserDomain) GetUserService() *services.UserService {
	return d.userService
}
