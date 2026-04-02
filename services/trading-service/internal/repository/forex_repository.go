package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type ForexRepository interface {
	Count(ctx context.Context) (int64, error)
	Upsert(ctx context.Context, pair model.ForexPair) error
	FindAll(ctx context.Context, filter ListingFilter) ([]model.ForexPair, int64, error)
	FindByListingIDs(ctx context.Context, listingIDs []uint) ([]model.ForexPair, error)
}
