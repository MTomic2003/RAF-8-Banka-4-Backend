package dto

type CreatePayeeRequest struct {
	Name          string `json:"name"          binding:"required"`
	AccountNumber string `json:"account_number" binding:"required"`
}

type UpdatePayeeRequest struct {
	Name          string `json:"name"          binding:"omitempty"`
	AccountNumber string `json:"account_number" binding:"omitempty"`
}
