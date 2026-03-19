package dto

import "time"

type UpdateClientRequest struct {
	FirstName   *string    `json:"first_name" binding:"omitempty,max=20"`
	LastName    *string    `json:"last_name" binding:"omitempty,max=100"`
	Gender      *string    `json:"gender"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	Email       *string    `json:"email" binding:"omitempty,email"`
	PhoneNumber *string    `json:"phone_number"`
	Address     *string    `json:"address"`
	Username    *string    `json:"username"`
}
