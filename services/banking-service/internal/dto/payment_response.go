package dto

import (
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type CreatePaymentResponse struct {
	PaymentID uint `json:"id"`
}

type VerifyPaymentResponse struct {
	PaymentID uint `json:"id"`
}

type PaymentResponse struct {
	PaymentID              uint      `json:"payment_id"`
	RecipientName          string    `json:"recipient_name"`
	ReferenceNumber        string    `json:"reference_number"`
	PaymentCode            string    `json:"payment_code"`
	Purpose                string    `json:"purpose"`
	PayerAccountNumber     string    `json:"payer_account_number"`
	RecipientAccountNumber string    `json:"recipient_account_number"`
	Amount                 float64   `json:"amount"`
	CurrencyCode           string    `json:"currency_code"`
	Status                 string    `json:"status"`
	CreatedAt              time.Time `json:"created_at"`
	Commission             float64   `json:"commission"`
}

func ToPaymentResponse(p *model.Payment) PaymentResponse {
	return PaymentResponse{
		PaymentID:              p.PaymentID,
		RecipientName:          p.RecipientName,
		ReferenceNumber:        p.ReferenceNumber,
		PaymentCode:            p.PaymentCode,
		Purpose:                p.Purpose,
		PayerAccountNumber:     p.Transaction.PayerAccountNumber,
		RecipientAccountNumber: p.Transaction.RecipientAccountNumber,
		Amount:                 p.Transaction.StartAmount - p.Transaction.Commission,
		CurrencyCode:           string(p.Transaction.StartCurrencyCode),
		Status:                 string(p.Transaction.Status),
		CreatedAt:              p.Transaction.CreatedAt,
		Commission:             p.Transaction.Commission,
	}
}
