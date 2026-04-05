package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type LoanRequestRepository interface {
	FindAll(ctx context.Context, query *dto.ListLoanRequestsQuery) ([]model.LoanRequest, int64, error)
	FindByID(ctx context.Context, id uint) (*model.LoanRequest, error)
	Update(ctx context.Context, request *model.LoanRequest) error
	CreateRequest(ctx context.Context, request *model.LoanRequest) error
}
