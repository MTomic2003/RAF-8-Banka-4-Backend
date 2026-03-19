package dto

import (
	"banking-service/internal/model"
	"time"
)

type TransferResponse struct {
	TransferID        uint      `json:"transfer_id"`
	TransactionID     uint      `json:"transaction_id"`
	FromAccountNumber string    `json:"from_account_number"`
	ToAccountNumber   string    `json:"to_account_number"`
	InitialAmount     float64   `json:"initial_amount"`
	FinalAmount       float64   `json:"final_amount"`
	ExchangeRate      *float64  `json:"exchange_rate,omitempty"`
	Commission        float64   `json:"commission"`
	CreatedAt         time.Time `json:"created_at"`
}

type ListTransfersResponse struct {
	Data       []TransferResponse `json:"data"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

func ToTransferResponse(transfer *model.Transfer) TransferResponse {
	return TransferResponse{
		TransferID:        transfer.TransferID,
		TransactionID:     transfer.TransactionID,
		FromAccountNumber: transfer.Transaction.PayerAccountNumber,
		ToAccountNumber:   transfer.Transaction.RecipientAccountNumber,
		InitialAmount:     transfer.Transaction.StartAmount - transfer.Commission,
		FinalAmount:       transfer.Transaction.EndAmount,
		ExchangeRate:      transfer.ExchangeRate,
		Commission:        transfer.Commission,
		CreatedAt:         transfer.Transaction.CreatedAt,
	}
}
