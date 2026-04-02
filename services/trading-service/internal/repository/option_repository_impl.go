package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
)

type optionRepository struct {
	db *gorm.DB
}

func NewOptionRepository(db *gorm.DB) OptionRepository {
	return &optionRepository{db: db}
}

// Upsert inserts or updates an Option matched by ListingID.
func (r *optionRepository) Upsert(ctx context.Context, option *model.Option) error {
	return r.db.WithContext(ctx).
		Where(model.Option{ListingID: option.ListingID}).
		Assign(*option).
		FirstOrCreate(option).Error
}

func (r *optionRepository) FindByListingIDs(ctx context.Context, listingIDs []uint) ([]model.Option, error) {
	var options []model.Option
	if err := r.db.WithContext(ctx).Where("listing_id IN ?", listingIDs).Preload("Listing").Find(&options).Error; err != nil {
		return nil, err
	}
	return options, nil
}
func (r *optionRepository) FindByStockID(ctx context.Context, stockID uint) ([]model.Option, error) {
	var options []model.Option

	if err := r.db.WithContext(ctx).
		Where("stock_id = ?", stockID).
		Preload("Listing").
		Find(&options).Error; err != nil {
		return nil, err
	}

	return options, nil
}
