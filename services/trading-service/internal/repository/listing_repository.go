package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type ListingRepository interface {
	FindAll(ctx context.Context) ([]model.Listing, error)
	Upsert(ctx context.Context, listing *model.Listing) error
	UpdatePriceAndAsk(ctx context.Context, listing *model.Listing, price, ask float64) error
	Count(ctx context.Context) (int64, error)
}
