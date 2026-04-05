package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type OrderTransactionRepository interface {
	Create(ctx context.Context, orderTransaction *model.OrderTransaction) error
}
