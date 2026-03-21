package dto

import "github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"

type ListLoanRequestsQuery struct {
	ClientID uint                    `form:"client_id"`
	Status   model.LoanRequestStatus `form:"status"`
	Page     int                     `form:"page"`
	PageSize int                     `form:"page_size"`
}
