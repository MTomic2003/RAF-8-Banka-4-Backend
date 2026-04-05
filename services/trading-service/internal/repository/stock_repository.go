package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type StockRepository interface {
	Upsert(ctx context.Context, stock *model.Stock) error
	FindByListingIDs(ctx context.Context, listingIDs []uint) ([]model.Stock, error)
	FindAll(ctx context.Context) ([]model.Stock, error)
	Count(ctx context.Context) (int64, error)
}
