package dto

type (
	InitiateEmailVerificationRequest struct {
		Email      string `json:"email" validate:"required,email"`
		RedirectTo string `json:"redirect_to" validate:"omitempty,url"`
	}
	ValidateEmailVerificationRequest struct {
		Token string `json:"token" validate:"required"`
	}
	SignInWithEmailRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"min=8,max=24"`
	}
	AccessTokenPayload struct {
		UserID string `json:"user_id"` // User ID
		Email  string `json:"email"`   // User Email
		SID    string `json:"sid"`     // Session ID
	}
	SetUserPasswordRequest struct {
		UserID               string `json:"user_id" validate:"required,uuid"`
		Password             string `json:"password" validate:"required,min=8" example:"secure.password"`
		PasswordConfirmation string `json:"password_confirmation" validate:"required,eqfield=Password" example:"secret.password"`
	}
)
