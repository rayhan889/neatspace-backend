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
	EmailVerification(c *fiber.Ctx) error
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
	g.Post("/email-verification", middlewares.ValidateRequestJSON[dto.EmailVerification](), h.EmailVerification)
}

func (h *AuthHandler) EmailVerification(c *fiber.Ctx) error {
	req := c.Locals(constants.RequestBodyJSONKey).(*dto.EmailVerification)

	err := h.authService.EmailVerification(c.Context(), req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(apputils.SuccessResponse(nil))
}
