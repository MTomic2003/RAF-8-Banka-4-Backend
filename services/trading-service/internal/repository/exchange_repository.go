package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type ExchangeRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]model.Exchange, int64, error)
	FindByMicCode(ctx context.Context, micCode string) (*model.Exchange, error)
	ToggleTradingEnabled(ctx context.Context, micCode string) (*model.Exchange, error)
}
