package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type TransferRepository interface {
	Create(ctx context.Context, transfer *model.Transfer) error
	ListByClientID(ctx context.Context, clientID uint, page, pageSize int) ([]model.Transfer, int64, error)
}
