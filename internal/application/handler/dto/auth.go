package dto

type (
	InitiateEmailVerification struct {
		Email      string `json:"email" validate:"required,email"`
		RedirectTo string `json:"redirect_to" validate:"omitempty,url"`
	}
	ValidateEmailVerification struct {
		Token string `json:"token" validate:"required"`
	}
)
