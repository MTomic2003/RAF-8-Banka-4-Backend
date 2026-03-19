package dto

import "banking-service/internal/model"

type ListLoanRequestsQuery struct {
	ClientID uint                    `form:"client_id"`
	Status   model.LoanRequestStatus `form:"status"`
	Page     int                     `form:"page"`
	PageSize int                     `form:"page_size"`
}