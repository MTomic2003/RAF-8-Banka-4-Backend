package repository

import (
	"context"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type PaymentFilter struct {
	DateFrom  *time.Time
	DateTo    *time.Time
	AmountMin *float64
	AmountMax *float64
	Status    *model.TransactionStatus
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *model.Payment) error
	GetByID(ctx context.Context, id uint) (*model.Payment, error)
	Update(ctx context.Context, payment *model.Payment) error
	FindByClient(ctx context.Context, clientID uint, filters *dto.PaymentFilters) ([]model.Payment, int64, error)
	FindByAccount(ctx context.Context, accountNumber string, filters *dto.PaymentFilters) ([]model.Payment, int64, error)
}
