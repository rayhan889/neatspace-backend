package dto

type (
	EmailVerification struct {
		Email      string `json:"email" validate:"required,email"`
		RedirectTo string `json:"redirect_to" validate:"omitempty,url"`
	}
)
