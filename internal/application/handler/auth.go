package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/rayhan889/neatspace/internal/application/constants"
	"github.com/rayhan889/neatspace/internal/application/handler/dto"
	"github.com/rayhan889/neatspace/internal/application/middlewares"
	"github.com/rayhan889/neatspace/internal/application/services"
	authEntity "github.com/rayhan889/neatspace/internal/domain/auth/entities"
	"github.com/rayhan889/neatspace/pkg/apputils"
)

type AuthHandlerInterface interface {
	InitiateEmailVerification(c *fiber.Ctx) error
	ValidateEmailVerification(c *fiber.Ctx) error
	SignInWithEmail(c *fiber.Ctx) error
	SetUserPassword(c *fiber.Ctx) error
	UpdateUserPassword(c *fiber.Ctx) error
}

var _ AuthHandlerInterface = (*AuthHandler)(nil)

type AuthHandler struct {
	authService services.AuthServiceInterface
}

type AuthHandlerOpts struct {
	RouteGroup   fiber.Router
	AuthService  services.AuthServiceInterface
	JWTSecretKey []byte
	SigningAlg   jwa.SignatureAlgorithm
}

func NewAuthHandler(opts AuthHandlerOpts) {
	h := &AuthHandler{
		authService: opts.AuthService,
	}

	publicGroup := opts.RouteGroup.Group("/auth")
	publicGroup.Post("/verification/email/initiate", middlewares.ValidateRequestJSON[dto.InitiateEmailVerificationRequest](), h.InitiateEmailVerification)
	publicGroup.Post("/verification/email/validate", middlewares.ValidateRequestJSON[dto.ValidateEmailVerificationRequest](), h.ValidateEmailVerification)
	publicGroup.Post("/signin/email", middlewares.ValidateRequestJSON[dto.SignInWithEmailRequest](), h.SignInWithEmail)

	privateGroup := publicGroup.Group("", middlewares.JWTMiddleware(opts.JWTSecretKey, opts.SigningAlg))
	privateGroup.Post("/password", middlewares.ValidateRequestJSON[dto.SetUserPasswordRequest](), h.SetUserPassword)
	privateGroup.Patch("/password/:userId", middlewares.ValidateRequestJSON[dto.UpdatePasswordRequest](), h.UpdateUserPassword)
}

// InitiateEmailVerification godoc
// @Summary		Initiate Email Verification
// @Description	Send email verification link to user's email address
// @Tags			Authentication
// @Accept			json
// @Produce			json
// @Param			body	body	dto.InitiateEmailVerificationRequest	true	"Email verification request"
// @Success		200	{object}	apputils.BaseResponse
// @Failure		400	{object}	apputils.BaseResponse
// @Failure		404	{object}	apputils.BaseResponse
// @Failure		500	{object}	apputils.BaseResponse
// @Router			/api/v1/auth/verification/email/initiate [post]
func (h *AuthHandler) InitiateEmailVerification(c *fiber.Ctx) error {
	req := c.Locals(constants.RequestBodyJSONKey).(*dto.InitiateEmailVerificationRequest)

	err := h.authService.InitiateEmailVerification(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(apputils.SuccessResponse(map[string]any{
		"message": "Email verification sent successfully",
	}))
}

// ValidateEmailVerification godoc
// @Summary		Validate Email Verification
// @Description	Validate email verification token and mark email as verified
// @Tags			Authentication
// @Accept			json
// @Produce			json
// @Param			body	body	dto.ValidateEmailVerificationRequest	true	"Email verification token"
// @Success		200	{object}	apputils.BaseResponse
// @Failure		400	{object}	apputils.BaseResponse
// @Failure		401	{object}	apputils.BaseResponse
// @Failure		500	{object}	apputils.BaseResponse
// @Router			/api/v1/auth/verification/email/validate [post]
func (h *AuthHandler) ValidateEmailVerification(c *fiber.Ctx) error {
	req := c.Locals(constants.RequestBodyJSONKey).(*dto.ValidateEmailVerificationRequest)

	err := h.authService.ValidateEmailVerification(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(apputils.SuccessResponse(map[string]any{
		"message": "Email verified successfully",
	}))
}

// SignInWithEmail godoc
// @Summary		Sign In with Email
// @Description	Authenticate user with email and password, returns access and refresh tokens
// @Tags			Authentication
// @Accept			json
// @Produce			json
// @Param			body	body	dto.SignInWithEmailRequest	true	"Sign in credentials"
// @Success		200	{object}	apputils.BaseResponse{data=authEntity.AuthenticatedUser}
// @Failure		400	{object}	apputils.BaseResponse
// @Failure		401	{object}	apputils.BaseResponse
// @Failure		500	{object}	apputils.BaseResponse
// @Router			/api/v1/auth/signin/email [post]
func (h *AuthHandler) SignInWithEmail(c *fiber.Ctx) error {
	req := c.Locals(constants.RequestBodyJSONKey).(*dto.SignInWithEmailRequest)

	authUser, err := h.authService.SignInWithEmail(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(apputils.SuccessResponse(map[string]any{
		"message": "Successfully siginin with email",
		"data":    authUser,
	}))
}

// SetUserPassword godoc
// @Summary		Set User Password
// @Description	Set password for a user account (requires authentication)
// @Tags			Authentication
// @Accept			json
// @Produce			json
// @Security		BearerAuth
// @Param			body	body	dto.SetUserPasswordRequest	true	"Password setup request"
// @Success		200	{object}	apputils.BaseResponse
// @Failure		400	{object}	apputils.BaseResponse
// @Failure		401	{object}	apputils.BaseResponse
// @Failure		404	{object}	apputils.BaseResponse
// @Failure		500	{object}	apputils.BaseResponse
// @Router			/api/v1/auth/password [post]
func (h *AuthHandler) SetUserPassword(c *fiber.Ctx) error {
	req := c.Locals(constants.RequestBodyJSONKey).(*dto.SetUserPasswordRequest)

	userID := apputils.UUIDChecker(req.UserID)
	userPassword := &authEntity.UserPasswordEntity{
		UserID:       userID,
		PasswordHash: []byte(req.Password),
		CreatedAt:    time.Now(),
	}

	err := h.authService.SetUserPassword(c.Context(), userPassword)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(apputils.SuccessResponse(map[string]any{
		"message": "Successfully set user password",
	}))
}

// UpdateUserPassword godoc
// @Summary		Update User Password
// @Description	Update password for a specific user (requires authentication)
// @Tags			Authentication
// @Accept			json
// @Produce			json
// @Security		BearerAuth
// @Param			userId	path	string	true	"User ID (UUID)"
// @Param			body	body	dto.UpdatePasswordRequest	true	"Password update request"
// @Success		200	{object}	apputils.BaseResponse
// @Failure		400	{object}	apputils.BaseResponse
// @Failure		401	{object}	apputils.BaseResponse
// @Failure		404	{object}	apputils.BaseResponse
// @Failure		500	{object}	apputils.BaseResponse
// @Router			/api/v1/auth/password/{userId} [patch]
func (h *AuthHandler) UpdateUserPassword(c *fiber.Ctx) error {
	req := c.Locals(constants.RequestBodyJSONKey).(*dto.UpdatePasswordRequest)

	userID := c.Params("userId")
	userUUID := apputils.UUIDChecker(userID)

	err := h.authService.UpdateUserPassword(c.Context(), req.Password, userUUID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(apputils.SuccessResponse(map[string]any{
		"message": "Successfully update user password",
	}))
}
