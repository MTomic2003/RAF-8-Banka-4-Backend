package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type CardRequestRepository interface {
	Create(ctx context.Context, request *model.CardRequest) error
	FindByAccountNumberAndCode(ctx context.Context, accountNumber string, code string) (*model.CardRequest, error)
	FindLatestPendingByAccountNumber(ctx context.Context, accountNumber string) (*model.CardRequest, error)
	Update(ctx context.Context, request *model.CardRequest) error
}
