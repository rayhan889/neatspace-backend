package dto

type (
	InitiateEmailVerification struct {
		Email      string `json:"email" validate:"required,email"`
		RedirectTo string `json:"redirect_to" validate:"omitempty,url"`
	}
	ValidateEmailVerification struct {
		Token string `json:"token" validate:"required"`
	}
	SignInWithEmail struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"min=8,max=24"`
	}
	AccessTokenPayload struct {
		UserID string `json:"user_id"` // User ID
		Email  string `json:"email"`   // User Email
		SID    string `json:"sid"`     // Session ID
	}
)
