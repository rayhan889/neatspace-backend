package dto

import (
	"time"

	noteEntity "github.com/rayhan889/neatspace/internal/domain/note/entities"
)

type (
	NotePaginationResponse struct {
		Title     string                   `json:"title"`
		Content   noteEntity.TiptapContent `json:"content"`
		CreatedAt time.Time                `json:"created_at"`
		UpdatedAt *time.Time               `json:"updated_at"`
	}
	CreateNoteRequest struct {
		Title   string                   `json:"title" validate:"required,min=3,max=100"`
		Content noteEntity.TiptapContent `json:"content" validate:"required"`
	}
)
