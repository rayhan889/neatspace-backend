package entities

import (
	"time"

	"github.com/google/uuid"
)

const NoteTable = "public.notes"

type NoteEntity struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	UserID      uuid.UUID     `json:"user_id" db:"user_id"`
	Title       string        `json:"title" db:"title"`
	Content     TiptapContent `json:"content" db:"content"`
	ContentText string        `json:"content_text" db:"content_text"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time    `json:"updated_at" db:"updated_at"`
}

type TiptapContent struct {
	Type    string          `json:"type"`
	Text    string          `json:"text,omitempty"`
	Content []TiptapContent `json:"content,omitempty"`
}
