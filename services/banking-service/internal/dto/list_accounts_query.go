package dto

import "banking-service/internal/model"

type ListAccountsQuery struct {
	ClientID    *uint             `form:"client_id"`
	AccountType model.AccountType `form:"account_type"`
	AccountKind model.AccountKind `form:"account_kind"`
	Status      string            `form:"status"`
	CurrencyID  *uint             `form:"currency_id"`
	Page        int               `form:"page"`
	PageSize    int               `form:"page_size"`
}
