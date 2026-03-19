package dto

type TransferRequest struct {
	FromAccountNumber string  `json:"from_account" binding:"required"`
	ToAccountNumber   string  `json:"to_account" binding:"required"`
	Amount            float64 `json:"amount" binding:"required,gt=0"`
}
