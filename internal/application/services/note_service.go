package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rayhan889/neatspace/internal/application/handler/dto"
	noteEntity "github.com/rayhan889/neatspace/internal/domain/note/entities"
	"github.com/rayhan889/neatspace/internal/domain/note/repositories"
	"github.com/rayhan889/neatspace/pkg/apputils"
)

type NoteServiceInterface interface {
	PaginationNote(c *fiber.Ctx, p *apputils.Pagination) (data []dto.NotePaginationResponse, total int, err error)
	CreateNote(ctx context.Context, note *noteEntity.NoteEntity) error
}

var _ NoteServiceInterface = (*NoteService)(nil)

type NoteService struct {
	noteRepo    repositories.NoteRepositoryInterface
	userService UserServiceInterface
}

type NoteServiceOpts struct {
	NoteRepo    repositories.NoteRepositoryInterface
	UserService UserServiceInterface
}

func NewNoteService(opts NoteServiceOpts) *NoteService {
	return &NoteService{
		noteRepo:    opts.NoteRepo,
		userService: opts.UserService,
	}
}

func (s *NoteService) PaginationNote(c *fiber.Ctx, p *apputils.Pagination) (data []dto.NotePaginationResponse, total int, err error) {
	return s.noteRepo.PaginationNote(c, p)
}

func (s *NoteService) CreateNote(ctx context.Context, note *noteEntity.NoteEntity) error {
	if !s.userService.IsUserExistsByID(ctx, note.UserID) {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("user with id %s not found", note.UserID.String()))
	}

	if len(note.Content.Content) > 0 {
		var doc noteEntity.TiptapContent
		bytes, err := json.Marshal(note.Content)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "error marshalling content")
		}
		err = json.Unmarshal(bytes, &doc)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "error unmarshalling content")
		}

		note.ContentText = s.extractContentToText(doc.Content)
	}

	return s.noteRepo.CreateNote(ctx, note)
}

func (s *NoteService) extractContentToText(nodes []noteEntity.TiptapContent) string {
	var sb strings.Builder

	for _, node := range nodes {
		if node.Type == "text" {
			sb.WriteString(node.Text)
		}

		if len(node.Content) > 0 {
			sb.WriteString(s.extractContentToText(node.Content))
		}

		switch node.Type {
		case "paragraph", "heading", "bulletList", "orderedList", "listItem":
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
