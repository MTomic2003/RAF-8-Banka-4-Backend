package dto

import "time"

type CreateClientRequest struct {
	FirstName   string    `json:"first_name" binding:"required,max=50"`
	LastName    string    `json:"last_name" binding:"required,max=100"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	Email       string    `json:"email" binding:"required,email"`
	Username    string    `json:"username" binding:"required"`
	PhoneNumber string    `json:"phone_number"`
	Address     string    `json:"address"`
}
