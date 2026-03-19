package dto

import "banking-service/internal/model"

type PayeeResponse struct {
	PayeeID       uint   `json:"payee_id"`
	ClientID      uint   `json:"client_d"`
	Name          string `json:"name"`
	AccountNumber string `json:"account_number"`
}

func ToPayeeResponse(p *model.Payee) PayeeResponse {
	return PayeeResponse{
		PayeeID:       p.PayeeID,
		ClientID:      p.ClientID,
		Name:          p.Name,
		AccountNumber: p.AccountNumber,
	}
}
