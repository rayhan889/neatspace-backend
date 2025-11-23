package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"reflect"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rayhan889/neatspace/internal/application/handler/dto"
	authEntity "github.com/rayhan889/neatspace/internal/domain/auth/entities"
	authRepo "github.com/rayhan889/neatspace/internal/domain/auth/repositories"
	"github.com/rayhan889/neatspace/internal/notification"
	"github.com/rayhan889/neatspace/pkg/apputils"
)

type AuthServiceInterface interface {
	InitiateEmailVerification(ctx context.Context, req *dto.InitiateEmailVerification) error
	ValidateEmailVerification(ctx context.Context, req *dto.ValidateEmailVerification) error
}

var _ AuthServiceInterface = (*AuthService)(nil)

type AuthService struct {
	authRepo    authRepo.AuthRepositoryInterface
	userService UserServiceInterface
	logger      *slog.Logger
	mailer      *notification.Mailer
	baseURL     string
	port        int
	host        string
}

type AuthServiceOpts struct {
	AuthRepository authRepo.AuthRepositoryInterface
	UserService    UserServiceInterface
	Logger         *slog.Logger
	Mailer         *notification.Mailer
	BaseURL        string
	Port           int
	Host           string
}

func NewAuthService(opts AuthServiceOpts) *AuthService {
	if opts.AuthRepository == nil {
		panic("auth repo is required")
	}

	if opts.UserService == nil {
		panic("user service is required")
	}

	if opts.Logger == nil {
		panic("logger is required")
	}

	if opts.Mailer == nil {
		panic("mailer is required")
	}

	if opts.BaseURL == "" {
		panic("base URL is required")
	}

	if opts.Port == 0 {
		panic("port is required")
	}

	if opts.Host == "" {
		panic("host is required")
	}

	return &AuthService{
		authRepo:    opts.AuthRepository,
		userService: opts.UserService,
		logger:      opts.Logger,
		mailer:      opts.Mailer,
		baseURL:     opts.BaseURL,
		port:        opts.Port,
		host:        opts.Host,
	}
}

func (s *AuthService) InitiateEmailVerification(ctx context.Context, req *dto.InitiateEmailVerification) error {
	user, err := s.userService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	if user == nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	if s.isEmailVerified(user) {
		return fiber.NewError(fiber.StatusBadRequest, "email already verified")
	}

	userID := user.ID
	email := user.Email

	tokens, err := s.authRepo.FindAllOneTimeTokens(ctx)
	if err != nil {
		return err
	}

	var existingToken *authEntity.OneTimeToken
	now := time.Now()
	for _, t := range tokens {
		if t.UserID != nil && *t.UserID == userID && t.Subject == authEntity.OneTimeTokenSubjectEmailVerification && now.Before(t.ExpiresAt) {
			existingToken = &t
			break
		}
	}

	if existingToken != nil {
		existingToken.LastSentAt = &now
		if err := s.authRepo.UpdateOneTImeTokenLastSentAt(ctx, existingToken.ID, now); err != nil {
			return err
		}
	}

	rawToken, err := apputils.GenerateURLSafeToken(48)
	if err != nil {
		return err
	}
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])
	expiresAt := now.Add(15 * time.Minute)

	for _, t := range tokens {
		if t.UserID != nil && *t.UserID == userID && t.Subject == authEntity.OneTimeTokenSubjectEmailVerification {
			if err := s.authRepo.DeleteOneTimeToken(ctx, t.ID); err != nil {
				return err
			}
		}
	}

	var metadata map[string]any
	if req.RedirectTo != "" {
		metadata = map[string]any{
			"redirect_to": req.RedirectTo,
		}
	}

	token := &authEntity.OneTimeToken{
		ID:         uuid.New(),
		UserID:     &userID,
		Subject:    authEntity.OneTimeTokenSubjectEmailVerification,
		TokenHash:  tokenHash,
		RelatesTo:  email,
		Metadata:   metadata,
		CreatedAt:  now,
		ExpiresAt:  expiresAt,
		LastSentAt: &now,
	}
	if err := s.authRepo.CreateOneTimeToken(ctx, token); err != nil {
		return err
	}

	if err := s.sendVerificationEmail(ctx, email, rawToken, req.RedirectTo); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) ValidateEmailVerification(ctx context.Context, req *dto.ValidateEmailVerification) error {
	hash := sha256.Sum256([]byte(req.Token))
	tokenHash := hex.EncodeToString(hash[:])

	oneTimeToken, err := s.authRepo.GetOneTimeTokenByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}

	if oneTimeToken == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "one time token is not found")
	}

	if oneTimeToken.ExpiresAt.Before(time.Now()) {
		return fiber.NewError(fiber.StatusUnauthorized, "one time token is expired")
	}

	if oneTimeToken.UserID == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "token is not related to any user")
	}
	userID := *oneTimeToken.UserID

	_ = s.authRepo.DeleteOneTimeToken(ctx, oneTimeToken.ID)

	if err = s.userService.MarkEmailVerified(ctx, userID); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) isEmailVerified(u any) bool {
	if u == nil {
		return false
	}
	v := reflect.ValueOf(u)
	if !v.IsValid() {
		return false
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}

	// bool style fields
	for _, name := range []string{"EmailVerified", "IsEmailVerified", "Verified"} {
		f := v.FieldByName(name)
		if f.IsValid() && f.Kind() == reflect.Bool && f.Bool() {
			return true
		}
	}

	// time/timestamp style fields
	for _, name := range []string{"EmailVerifiedAt", "VerifiedAt"} {
		f := v.FieldByName(name)
		if !f.IsValid() {
			continue
		}
		// pointer to time
		if f.Kind() == reflect.Pointer {
			if !f.IsNil() {
				return true
			}
		}
		// struct (likely time.Time)
		if f.Kind() == reflect.Struct {
			if t, ok := f.Interface().(time.Time); ok && !t.IsZero() {
				return true
			}
		}
	}

	return false
}

func (s *AuthService) sendVerificationEmail(ctx context.Context, toEmail, rawToken, redirectTo string) error {
	// Determine base URL:
	// 1) prefer configured s.baseURL
	// 2) fallback to environment SERVER_HOST/SERVER_PORT
	// 3) final fallback to localhost:8000
	base := s.baseURL
	if base == "" {
		host := os.Getenv("SERVER_HOST")
		port := os.Getenv("SERVER_PORT")
		if host == "" {
			host = "localhost"
		}
		if port == "0" {
			port = "8000"
		}
		base = fmt.Sprintf("http://%s:%s", host, port)
	}

	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" {
		// fallback to SERVER_HOST/SERVER_PORT env vars explicitly
		host := os.Getenv("SERVER_HOST")
		port := os.Getenv("SERVER_PORT")
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "8000"
		}
		u = &url.URL{Scheme: "http", Host: fmt.Sprintf("%s:%s", host, port)}
	}

	// Use only the token in the verification link (do NOT include the email)
	u.Path = "/api/v1/auth/verify-email"
	q := u.Query()
	q.Set("token", rawToken)
	if redirectTo != "" {
		q.Set("redirect_to", redirectTo)
	}
	u.RawQuery = q.Encode()
	verifyURL := u.String()

	// Try to fetch user to pass display name to template
	var displayName string
	if s.userService != nil {
		if user, err := s.userService.GetUserByEmail(ctx, toEmail); err == nil && user != nil {
			displayName = user.DisplayName
		}
	}

	// Template data passed to the email template; template can access .VerifyURL, .Email and .DisplayName
	data := map[string]any{
		"Email":       toEmail,
		"DisplayName": displayName,
		"VerifyURL":   verifyURL,
		"AppName":     "Neatspace",
	}

	subject := "Verify your email address"
	templateFile := "email_verification.html"

	if s.mailer != nil {
		if err := s.mailer.SendMail(ctx, []string{toEmail}, subject, templateFile, data); err != nil {
			s.logger.Error("failed to send verification email", slog.String("op", "sendVerificationEmail"), slog.String("error", err.Error()))
			return err
		}
		return nil
	}

	s.logger.Warn("mailer not configured, cannot send verification email", slog.String("op", "sendVerificationEmail"))
	return nil
}
