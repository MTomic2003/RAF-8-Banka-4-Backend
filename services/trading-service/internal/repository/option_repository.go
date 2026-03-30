package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type OptionRepository interface {
	Upsert(ctx context.Context, option *model.Option) error
	FindByListingIDs(listingIDs []uint) ([]model.Option, error)
}
