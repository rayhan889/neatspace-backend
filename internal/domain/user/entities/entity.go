package entities

import (
	"time"

	"github.com/google/uuid"
)

const UserTable = "public.users"

type UserEntity struct {
	ID              uuid.UUID     `json:"id" db:"id"`
	DisplayName     string        `json:"display_name" db:"display_name"`
	Username        *string       `json:"username" db:"username"`
	Metadata        *UserMetadata `json:"metadata" db:"metadata"`
	Email           string        `json:"email" db:"email"`
	CreatedAt       time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time    `json:"updated_at" db:"updated_at"`
	EmailVerifiedAt *time.Time    `json:"email_verified_at" db:"email_verified_at"`
}

type UserMetadata struct {
	Timezone string `json:"timezone,omitempty"`
	Role     string `json:"role,omitempty"`
	// Add more metadata fields as needed
}

type FilterUser struct {
	Search *string `json:"search,omitempty" query:"search"`
	Limit  int     `json:"limit,omitempty" query:"limit"`
	Offset int     `json:"offset,omitempty" query:"offset"`
}

type UserWithCredential struct {
	UserEntity
	PasswordHash []byte `json:"password_hash" db:"-"`
}

func (u *UserEntity) GetID() uuid.UUID {
	return u.ID
}
func (u *UserEntity) GetEmail() string {
	return u.Email
}
func (u *UserEntity) AsUserModel() UserEntity {
	return *u
}
func (u *UserEntity) GetEmailVerifiedAt() *time.Time {
	return u.EmailVerifiedAt
}
