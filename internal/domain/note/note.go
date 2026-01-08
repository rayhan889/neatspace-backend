package note

import (
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rayhan889/neatspace/internal/application/services"
	"github.com/rayhan889/neatspace/internal/domain/note/repositories"
)

type Options struct {
	PgPool      *pgxpool.Pool
	Logger      *slog.Logger
	UserService services.UserServiceInterface
}

type NoteDomain struct {
	logger      *slog.Logger
	noteService *services.NoteService
	userService services.UserServiceInterface
}

func NewNoteDomain(opts *Options) *NoteDomain {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}))
	}

	noteService := services.NewNoteService(services.NoteServiceOpts{
		NoteRepo:    repositories.NewNoteRepository(opts.PgPool, logger),
		UserService: opts.UserService,
	})

	return &NoteDomain{
		logger:      logger,
		noteService: noteService,
	}
}

func (d *NoteDomain) GetNoteService() services.NoteServiceInterface {
	return d.noteService
}
