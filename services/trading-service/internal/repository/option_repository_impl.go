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

func (r *optionRepository) Upsert(ctx context.Context, option *model.Option) error {
	return r.db.WithContext(ctx).
		Where(model.Option{AssetID: option.AssetID}).
		Assign(*option).
		FirstOrCreate(option).Error
}

func (r *optionRepository) FindByAssetIDs(ctx context.Context, assetIDs []uint) ([]model.Option, error) {
	var options []model.Option
	if err := r.db.WithContext(ctx).Where("asset_id IN ?", assetIDs).Preload("Asset").Preload("Listing").Preload("Listing.Exchange").Find(&options).Error; err != nil {
		return nil, err
	}
	return options, nil
}

func (r *optionRepository) FindByStockID(ctx context.Context, stockID uint) ([]model.Option, error) {
	var options []model.Option

	if err := r.db.WithContext(ctx).
		Where("stock_id = ?", stockID).
		Preload("Asset").
		Preload("Listing").
		Find(&options).Error; err != nil {
		return nil, err
	}

	return options, nil
}
