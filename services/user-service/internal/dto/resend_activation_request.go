package dto

type ResendActivationRequest struct {
	Email string `json:"email" binding:"required,email"`
}
