package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rayhan889/neatspace/internal/application/handler/dto"
	noteEntity "github.com/rayhan889/neatspace/internal/domain/note/entities"
	"github.com/rayhan889/neatspace/pkg/apputils"
)

type NoteRepositoryInterface interface {
	PaginationNote(c *fiber.Ctx, p *apputils.Pagination) (data []dto.NotePaginationResponse, total int, err error)
	CreateNote(ctx context.Context, note *noteEntity.NoteEntity) error
}

var _ NoteRepositoryInterface = (*NoteRepository)(nil)

type NoteRepository struct {
	pgPool *pgxpool.Pool
	logger *slog.Logger
}

func NewNoteRepository(pgPool *pgxpool.Pool, logger *slog.Logger) *NoteRepository {
	return &NoteRepository{
		pgPool: pgPool,
		logger: logger,
	}
}

func (r *NoteRepository) PaginationNote(c *fiber.Ctx, p *apputils.Pagination) (data []dto.NotePaginationResponse, total int, err error) {
	query := fmt.Sprintf(`
			SELECT title, content, created_at, updated_at 
			FROM %s 
			WHERE 1=1`,
		noteEntity.NoteTable)

	query, args := r.queryFilter(c, query)

	argPos := len(args) + 1
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, p.Limit, p.Offset)

	rows, err := r.pgPool.Query(c.Context(), query, args...)
	if err != nil {
		r.logger.Error("failed to query notes", slog.String("op", "PaginationNote"), slog.String("err", err.Error()))
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var item dto.NotePaginationResponse
		var tiptapContent noteEntity.TiptapContent

		var tiptapContentBytes []byte

		err := rows.Scan(
			&item.Title,
			&tiptapContentBytes,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan note row", slog.String("op", "PaginationNote"), slog.String("err", err.Error()))
			return nil, 0, err
		}
		if len(tiptapContentBytes) > 0 {
			err = json.Unmarshal(tiptapContentBytes, &tiptapContent)
			if err != nil {
				r.logger.Error("failed to unmarshal tiptap content", slog.String("op", "PaginationNote"), slog.String("err", err.Error()))
				return nil, 0, err
			}
		}

		item.Content = tiptapContent

		data = append(data, item)
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM %s
		WHERE 1=1
	`, noteEntity.NoteTable)

	countQuery, countArgs := r.queryFilter(c, countQuery)

	err = r.pgPool.QueryRow(c.Context(), countQuery, countArgs...).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count notes", slog.String("op", "PaginationNote"), slog.String("err", err.Error()))
		return nil, 0, err
	}

	return data, total, nil

}

func (r *NoteRepository) CreateNote(ctx context.Context, note *noteEntity.NoteEntity) error {
	_, err := r.pgPool.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (id, title, user_id, content, content_text, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)`, noteEntity.NoteTable),
		note.ID,
		note.Title,
		note.UserID,
		note.Content,
		note.ContentText,
		note.CreatedAt,
		note.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("failed to create note", slog.String("op", "CreateNote"), slog.String("err", err.Error()))
		return err
	}

	r.logger.Info("note created successfully", slog.String("op", "CreateNote"), slog.String("note_id", note.ID.String()))
	return nil
}

func (r *NoteRepository) queryFilter(c *fiber.Ctx, baseQuery string) (string, []interface{}) {
	if baseQuery == "" {
		baseQuery = "WHERE 1=1"
	}

	var args []interface{}
	argPos := 1

	// TO BE IMPLEMENTED LATER
	// if search := c.Query("search"); search != "" {
	// 	ilike := "%" + search + "%"

	// }

	if userID := c.Query("user_id"); userID != "" {
		baseQuery += fmt.Sprintf(" AND %s.user_id = $%d", noteEntity.NoteTable, argPos)
		args = append(args, userID)
		argPos++
	}

	return baseQuery, args
}
