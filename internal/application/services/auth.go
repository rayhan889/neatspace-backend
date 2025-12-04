package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"reflect"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/rayhan889/neatspace/internal/application/handler/dto"
	authEntity "github.com/rayhan889/neatspace/internal/domain/auth/entities"
	authRepo "github.com/rayhan889/neatspace/internal/domain/auth/repositories"
	"github.com/rayhan889/neatspace/internal/notification"
	"github.com/rayhan889/neatspace/pkg/apputils"
)

var ErrInvalidCredentials = fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")

type AuthServiceInterface interface {
	InitiateEmailVerification(ctx context.Context, req *dto.InitiateEmailVerificationRequest) error
	ValidateEmailVerification(ctx context.Context, req *dto.ValidateEmailVerificationRequest) error
	SignInWithEmail(ctx context.Context, req *dto.SignInWithEmailRequest) (*authEntity.AuthenticatedUser, error)
	CreateSession(ctx context.Context, session *authEntity.SessionEntity) error
	CreateRefreshToken(ctx context.Context, refreshToken *authEntity.RefreshToken) error
	SetUserPassword(ctx context.Context, userPassword *authEntity.UserPasswordEntity) error
}

var _ AuthServiceInterface = (*AuthService)(nil)

type AuthService struct {
	authRepo    authRepo.AuthRepositoryInterface
	userService UserServiceInterface
	logger      *slog.Logger
	mailer      *notification.Mailer
	baseURL     string

	secretKey          []byte                 // Secret key for signing JWTs
	accessTokenExpiry  time.Duration          // Access token expiration duration
	refreshTokenExpiry time.Duration          // Refresh token expiration duration
	signingAlg         jwa.SignatureAlgorithm // Signing algorithm (default: HS256)
}

type AuthServiceOpts struct {
	AuthRepository authRepo.AuthRepositoryInterface
	UserService    UserServiceInterface
	Logger         *slog.Logger
	Mailer         *notification.Mailer
	BaseURL        string

	JWTSecretKey       []byte                 // Secret key for signing JWTs
	AccessTokenExpiry  time.Duration          // Access token expiration duration
	RefreshTokenExpiry time.Duration          // Refresh token expiration duration
	SigningAlg         jwa.SignatureAlgorithm // Signing algorithm (default: HS256)
}

func NewAuthService(opts AuthServiceOpts) *AuthService {
	return &AuthService{
		authRepo:           opts.AuthRepository,
		userService:        opts.UserService,
		logger:             opts.Logger,
		mailer:             opts.Mailer,
		baseURL:            opts.BaseURL,
		secretKey:          opts.JWTSecretKey,
		accessTokenExpiry:  opts.AccessTokenExpiry,
		refreshTokenExpiry: opts.RefreshTokenExpiry,
		signingAlg:         opts.SigningAlg,
	}
}

func (s *AuthService) InitiateEmailVerification(ctx context.Context, req *dto.InitiateEmailVerificationRequest) error {
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

func (s *AuthService) ValidateEmailVerification(ctx context.Context, req *dto.ValidateEmailVerificationRequest) error {
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

func (s *AuthService) CreateSession(ctx context.Context, session *authEntity.SessionEntity) error {
	if session == nil {
		return errors.New("session is required")
	}
	if session.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if session.TokenHash == "" {
		return errors.New("token_hash is required")
	}
	if session.ExpiresAt.IsZero() || session.ExpiresAt.Before(time.Now()) {
		return errors.New("expires_at must be set and in the future")
	}

	return s.authRepo.CreateSession(ctx, session)
}

func (s *AuthService) CreateRefreshToken(ctx context.Context, refreshToken *authEntity.RefreshToken) error {
	if refreshToken == nil {
		return errors.New("refresh token is required")
	}
	if refreshToken.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if len(refreshToken.TokenHash) == 0 {
		return errors.New("token_hash is required")
	}
	if refreshToken.ExpiresAt.IsZero() || refreshToken.ExpiresAt.Before(time.Now()) {
		return errors.New("expires_at must be set and in the future")
	}

	return s.authRepo.CreateRefreshToken(ctx, refreshToken)
}

func (s *AuthService) SignInWithEmail(ctx context.Context, req *dto.SignInWithEmailRequest) (*authEntity.AuthenticatedUser, error) {
	if req.Email == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	ok, err := s.validatePassword(ctx, user.ID, req.Password)
	if err != nil || !ok {
		return nil, ErrInvalidCredentials
	}

	if !s.isEmailVerified(user) {
		return nil, ErrInvalidCredentials
	}

	jwtGen := apputils.NewJWTGenerator(apputils.JWTConfig{
		SecretKey:          s.secretKey,
		AccessTokenExpiry:  s.accessTokenExpiry,
		RefreshTokenExpiry: s.refreshTokenExpiry,
		SigninAlgo:         s.signingAlg,
		Issuer:             s.baseURL,
	})

	audience := "client-app"
	if md, ok := ctx.Value("headers").(map[string]string); ok {
		if aud, exists := md["X-App-Audience"]; exists && aud != "" {
			audience = aud
		}
	}

	refreshTokenUUID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	refreshTokenStr, err := jwtGen.GenerateRefreshTokenJWT(ctx, user.ID.String(), audience, refreshTokenUUID.String())
	if err != nil {
		return nil, err
	}
	refreshTokenHash := jwtGen.GetHash(refreshTokenStr)

	session := &authEntity.SessionEntity{
		ID:        uuid.New(),
		UserID:    user.GetID(),
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().Add(jwtGen.AccessTokenExpiry()),
	}
	if err := s.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	accessTokenPayload := dto.AccessTokenPayload{
		UserID: user.ID.String(),
		Email:  user.Email,
		SID:    session.ID.String(),
	}
	accessToken, err := jwtGen.Sign(ctx, accessTokenPayload, user.GetID().String())
	if err != nil {
		return nil, err
	}

	refreshToken := &authEntity.RefreshToken{
		ID:        refreshTokenUUID,
		UserID:    user.GetID(),
		SessionID: &session.ID,
		TokenHash: []byte(refreshTokenHash),
		ExpiresAt: time.Now().Add(jwtGen.RefreshTokenExpiry()),
	}
	if err := s.CreateRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	authUser := &authEntity.AuthenticatedUser{
		UserWithCredentials: authEntity.UserWithCredentials{
			User:         user.AsUserModel(),
			AccessToken:  accessToken,
			RefreshToken: refreshTokenStr,
		},
		SessionID:   &session.ID,
		TokenExpiry: session.ExpiresAt,
	}

	return authUser, nil
}

func (s *AuthService) SetUserPassword(ctx context.Context, userPassword *authEntity.UserPasswordEntity) error {
	if len(userPassword.PasswordHash) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "password can't be empty")
	}

	if !s.userService.IsUserExistsByID(ctx, userPassword.UserID) {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("user with id %s cannot be fount", userPassword.UserID.String()))
	}

	hasher := apputils.NewPasswordHasher()
	hashed, err := hasher.Hash(string(userPassword.PasswordHash))
	if err != nil {
		return err
	}
	userPassword.PasswordHash = []byte(hashed)

	return s.authRepo.CreateUserPassword(ctx, userPassword)
}

func (s *AuthService) validatePassword(ctx context.Context, userID uuid.UUID, password string) (bool, error) {
	userPassword, err := s.authRepo.GetUserPasswordByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("error getting user password %w", err)
	}
	if userPassword == nil {
		return false, fmt.Errorf("cannot find match user password by user id %s", userID)
	}

	hasher := apputils.NewPasswordHasher()
	return hasher.Validate(password, string(userPassword.PasswordHash))
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
