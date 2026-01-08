package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/rayhan889/neatspace/internal/application/constants"
	"github.com/rayhan889/neatspace/internal/application/handler/dto"
	"github.com/rayhan889/neatspace/internal/application/middlewares"
	"github.com/rayhan889/neatspace/internal/application/services"
	"github.com/rayhan889/neatspace/internal/domain/note/entities"
	"github.com/rayhan889/neatspace/pkg/apputils"
)

type NoteHandlerInterface interface {
	PaginationNote(c *fiber.Ctx) error
	CreateNote(c *fiber.Ctx) error
}

var _ NoteHandlerInterface = (*NoteHandler)(nil)

type NoteHandler struct {
	noteService services.NoteServiceInterface
}

type NoteHandlerOpts struct {
	RouteGroup   fiber.Router
	NoteService  services.NoteServiceInterface
	JWTSecretKey []byte
	SigningAlg   jwa.SignatureAlgorithm
}

func NewNoteHandler(opts NoteHandlerOpts) {
	h := &NoteHandler{
		noteService: opts.NoteService,
	}

	publicGroup := opts.RouteGroup.Group("/notes")

	privateGroup := publicGroup.Group("", middlewares.JWTMiddleware(opts.JWTSecretKey, opts.SigningAlg))
	privateGroup.Get("", h.PaginationNote)
	privateGroup.Post("/new", middlewares.ValidateRequestJSON[dto.CreateNoteRequest](), h.CreateNote)
}

func (h *NoteHandler) PaginationNote(c *fiber.Ctx) error {
	p := apputils.Paginate(c)

	data, total, err := h.noteService.PaginationNote(c, p)
	if err != nil {
		return err
	}

	meta := apputils.PaginationMetaBuilder(c, total)

	return c.Status(fiber.StatusOK).JSON(apputils.SuccessResponse(apputils.PaginationBuilder(data, *meta)))
}

func (h *NoteHandler) CreateNote(c *fiber.Ctx) error {
	req := c.Locals(constants.RequestBodyJSONKey).(*dto.CreateNoteRequest)

	userID := c.Locals("user_id").(string)
	userIDUUID := apputils.UUIDChecker(userID)

	note := &entities.NoteEntity{
		ID:          uuid.New(),
		UserID:      userIDUUID,
		Title:       req.Title,
		Content:     req.Content,
		ContentText: "",
		CreatedAt:   time.Now(),
	}

	err := h.noteService.CreateNote(c.Context(), note)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusCreated)
}
