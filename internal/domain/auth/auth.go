package auth

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lestrrat-go/jwx/jwa"
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

	JWTSecretKey       []byte                 // Secret key for signing JWTs
	AccessTokenExpiry  time.Duration          // Access token expiration duration
	RefreshTokenExpiry time.Duration          // Refresh token expiration duration
	SigningAlg         jwa.SignatureAlgorithm // Signing algorithm (default: HS256)
}

type AuthDomain struct {
	authService  *services.AuthService
	jwtSecretKey []byte
	signingAlgo  jwa.SignatureAlgorithm
}

func NewAuthDomain(opts *Options) *AuthDomain {
	if err := opts.validateAndSetDefaults(); err != nil {
		panic("invalid auth module options: " + err.Error())
	}

	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	}

	authService := services.NewAuthService(services.AuthServiceOpts{
		AuthRepository:     authRepo.NewAuthRepository(opts.PgPool, logger),
		UserService:        opts.UserService,
		Logger:             logger,
		Mailer:             opts.Mailer,
		BaseURL:            opts.BaseURL,
		JWTSecretKey:       opts.JWTSecretKey,
		AccessTokenExpiry:  opts.AccessTokenExpiry,
		RefreshTokenExpiry: opts.RefreshTokenExpiry,
		SigningAlg:         opts.SigningAlg,
	})

	return &AuthDomain{
		authService:  authService,
		jwtSecretKey: opts.JWTSecretKey,
		signingAlgo:  opts.SigningAlg,
	}
}

func (d *AuthDomain) GetAuthService() *services.AuthService {
	return d.authService
}

func (d *AuthDomain) GetJWTSecretKey() []byte {
	return d.jwtSecretKey
}

func (d *AuthDomain) GetSigningAlgo() jwa.SignatureAlgorithm {
	return d.signingAlgo
}

func (opts *Options) validateAndSetDefaults() error {
	if opts.PgPool == nil {
		return errors.New("pgPool is required")
	}
	if opts.UserService == nil {
		return errors.New("userService is required")
	}
	if opts.Logger == nil {
		return errors.New("logger is required")
	}
	if opts.Mailer == nil {
		return errors.New("mailer is required")
	}
	if len(opts.JWTSecretKey) == 0 {
		return errors.New("jWTSecretKey is required")
	}
	if opts.SigningAlg == "" {
		opts.SigningAlg = jwa.HS256
	}
	if opts.AccessTokenExpiry == 0 {
		opts.AccessTokenExpiry = 24 * time.Hour
	}
	if opts.RefreshTokenExpiry == 0 {
		opts.RefreshTokenExpiry = 7 * 24 * time.Hour
	}

	// BaseURL is mandatory and must be provided via Options.BaseURL
	if opts.BaseURL == "" {
		return errors.New("baseURL is required (set Options.BaseURL)")
	}

	return nil
}
