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

// PaginationUser godoc
// @Summary 		Pagination Users
// @Description 	Paginating Through List of Users
// @Tags 			Users
// @Accept 			json
// @Produce 		json
// @Param			page		query	int		false	"Page number (default: 1, min: 1)"				default(1)		minimum(1)
// @Param			per_page	query	int		false	"Items per page (default: 10, max: 100)"		default(10)		minimum(1)	maximum(100)
// @Param			search		query	string	false	"Search by display name or username"
// @Param			role		query	string	false	"Filter by role (user or admin)"				Enums(user, admin)
// @Success      	200   {object}  apputils.PaginationResponse[dto.UserPagination]
// @Failure      	400   {object}  apputils.BaseResponse
// @Failure      	404   {object}  apputils.BaseResponse
// @Failure      	500   {object}  apputils.BaseResponse
// @Router       	/api/v1/users [get]
func (h *UserHandler) PaginationUser(c *fiber.Ctx) error {
	p := apputils.Paginate(c)

	data, total, err := h.userService.PaginationUser(c, p)
	if err != nil {
		return err
	}

	meta := apputils.PaginationMetaBuilder(c, total)

	return c.Status(fiber.StatusOK).JSON(apputils.SuccessResponse(apputils.PaginationBuilder(data, *meta)))
}

// CreateUser godoc
// @Summary 		Create User
// @Description 	Create a new user account
// @Tags 			Users
// @Accept 			json
// @Produce 		json
// @Param			request	body	dto.CreateUser	true	"User creation request"
// @Success      	201
// @Failure      	400   {object}  apputils.BaseResponse
// @Failure      	422   {object}  apputils.BaseResponse
// @Failure      	500   {object}  apputils.BaseResponse
// @Router       	/api/v1/users [post]
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
