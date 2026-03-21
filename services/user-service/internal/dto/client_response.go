package dto

import (
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"
)

type ClientResponse struct {
	Id          uint      `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Gender      string    `json:"gender"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	Address     string    `json:"address"`
	Username    string    `json:"username"`
	Active      bool      `json:"active"`
}

type ListClientsResponse struct {
	Data       []ClientResponse `json:"data"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

func ToClientResponse(c *model.Client) *ClientResponse {
	return &ClientResponse{
		Id:          c.ClientID,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		Gender:      c.Gender,
		DateOfBirth: c.DateOfBirth,
		Email:       c.Identity.Email,
		PhoneNumber: c.PhoneNumber,
		Address:     c.Address,
		Username:    c.Identity.Username,
		Active:      c.Identity.Active,
	}
}

func ToClientResponseList(clients []model.Client, total int64, page, pageSize int) *ListClientsResponse {
	responses := make([]ClientResponse, len(clients))
	for i, e := range clients {
		responses[i] = *ToClientResponse(&e)
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	return &ListClientsResponse{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
