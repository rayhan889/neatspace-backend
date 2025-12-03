package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rayhan889/neatspace/internal/application/constants"
	"github.com/rayhan889/neatspace/internal/application/handler/dto"
	"github.com/rayhan889/neatspace/internal/application/middlewares"
	"github.com/rayhan889/neatspace/internal/application/services"
	"github.com/rayhan889/neatspace/pkg/apputils"
)

type AuthHandlerInterface interface {
	InitiateEmailVerification(c *fiber.Ctx) error
}

var _ AuthHandlerInterface = (*AuthHandler)(nil)

type AuthHandler struct {
	authService services.AuthServiceInterface
}

type AuthHandlerOpts struct {
	RouteGroup  fiber.Router
	AuthService services.AuthServiceInterface
}

func NewAuthHandler(opts AuthHandlerOpts) {
	h := &AuthHandler{
		authService: opts.AuthService,
	}

	g := opts.RouteGroup.Group("/auth")
	g.Post("/verification/email/initiate", middlewares.ValidateRequestJSON[dto.InitiateEmailVerification](), h.InitiateEmailVerification)
	g.Post("/verification/email/validate", middlewares.ValidateRequestJSON[dto.ValidateEmailVerification](), h.ValidateEmailVerification)
	g.Post("/signin/email", middlewares.ValidateRequestJSON[dto.SignInWithEmail](), h.SignInWithEmail)
}

func (h *AuthHandler) InitiateEmailVerification(c *fiber.Ctx) error {
	req := c.Locals(constants.RequestBodyJSONKey).(*dto.InitiateEmailVerification)

	err := h.authService.InitiateEmailVerification(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(apputils.SuccessResponse(map[string]any{
		"message": "Email verification sent successfully",
	}))
}

func (h *AuthHandler) ValidateEmailVerification(c *fiber.Ctx) error {
	req := c.Locals(constants.RequestBodyJSONKey).(*dto.ValidateEmailVerification)

	err := h.authService.ValidateEmailVerification(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(apputils.SuccessResponse(map[string]any{
		"message": "Email verified successfully",
	}))
}

func (h *AuthHandler) SignInWithEmail(c *fiber.Ctx) error {
	req := c.Locals(constants.RequestBodyJSONKey).(*dto.SignInWithEmail)

	authUser, err := h.authService.SignInWithEmail(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(apputils.SuccessResponse(map[string]any{
		"message": "Successfully siginin with email",
		"data":    authUser,
	}))
}
