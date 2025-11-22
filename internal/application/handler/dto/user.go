package dto

import userEntity "github.com/rayhan889/neatspace/internal/domain/user/entities"

type (
	CreateUser struct {
		DisplayName string `json:"display_name" validate:"required,min=10"`
		Email       string `json:"email" validate:"required,email"`
	}
	UserPagination struct {
		DisplayName     string                  `json:"display_name"`
		Username        string                  `json:"username"`
		Metadata        userEntity.UserMetadata `json:"metadata"`
		Email           string                  `json:"email"`
		EmailVerifiedAt *string                 `json:"email_verified_at,omitempty"`
		LastLoginAt     *string                 `json:"last_login_at,omitempty"`
	}
)
