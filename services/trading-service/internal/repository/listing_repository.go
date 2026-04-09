package repository

import (
	"context"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type ListingRepository interface {
	FindAll(ctx context.Context) ([]model.Listing, error)
	FindStocks(ctx context.Context, filter ListingFilter) ([]model.Listing, int64, error)
	FindFutures(ctx context.Context, filter ListingFilter) ([]model.Listing, int64, error)
	FindOptions(ctx context.Context, filter ListingFilter) ([]model.Listing, int64, error)
	FindByID(ctx context.Context, id uint, daysBack int) (*model.Listing, error)
	FindLatestDailyPriceInfo(ctx context.Context, listingID uint) (*model.ListingDailyPriceInfo, error)
	Upsert(ctx context.Context, listing *model.Listing) error
	UpdatePriceAndAsk(ctx context.Context, listing *model.Listing, price, ask float64) error
	Count(ctx context.Context) (int64, error)
	CreateDailyPriceInfo(ctx context.Context, info *model.ListingDailyPriceInfo) error
	FindLastDailyPriceInfo(ctx context.Context, listingID uint, beforeDate time.Time) (*model.ListingDailyPriceInfo, error)
	FindByAssetType(ctx context.Context, assetType model.AssetType) ([]model.Listing, error)
	FindByAssetIDs(ctx context.Context, assetIDs []uint) ([]model.Listing, error)
}
