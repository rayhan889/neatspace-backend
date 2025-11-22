package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rayhan889/neatspace/internal/application/constants"
	"github.com/rayhan889/neatspace/internal/application/handler/dto"
	"github.com/rayhan889/neatspace/internal/application/middlewares"
	"github.com/rayhan889/neatspace/internal/application/services"
	"github.com/rayhan889/neatspace/internal/domain/user/entities"
	"github.com/rayhan889/neatspace/pkg/apputils"
)

type UserHandlerInterface interface {
	PaginationUser(c *fiber.Ctx) error
	CreateUser(c *fiber.Ctx) error
}

var _ UserHandlerInterface = (*UserHandler)(nil)

type UserHandler struct {
	userService services.UserServiceInterface
}

type UserHandlerOpts struct {
	RouteGroup  fiber.Router
	UserService services.UserServiceInterface
}

func NewUserHandler(opts UserHandlerOpts) {
	h := &UserHandler{
		userService: opts.UserService,
	}

	g := opts.RouteGroup.Group("/users")
	g.Get("", h.PaginationUser)
	g.Post("", middlewares.ValidateRequestJSON[dto.CreateUser](), h.CreateUser)
}

func (h *UserHandler) PaginationUser(c *fiber.Ctx) error {
	p := apputils.Paginate(c)

	data, total, err := h.userService.PaginationUser(c, p)
	if err != nil {
		return err
	}

	meta := apputils.PaginationMetaBuilder(c, total)

	return c.Status(fiber.StatusOK).JSON(apputils.SuccessResponse(apputils.PaginationBuilder(data, *meta)))
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	req := c.Locals(constants.RequestBodyJSONKey).(*dto.CreateUser)

	user := &entities.UserEntity{
		ID:              uuid.New(),
		DisplayName:     req.DisplayName,
		Email:           req.Email,
		EmailVerifiedAt: nil,
		CreatedAt:       time.Now(),
	}

	err := h.userService.CreateUser(c.Context(), user)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusCreated)
}
