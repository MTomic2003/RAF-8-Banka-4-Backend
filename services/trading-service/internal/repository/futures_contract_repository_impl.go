package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
)

type futuresContractRepository struct {
	db *gorm.DB
}

func NewFuturesContractRepository(db *gorm.DB) FuturesContractRepository {
	return &futuresContractRepository{db: db}
}

func (r *futuresContractRepository) FindByAssetIDs(ctx context.Context, assetIDs []uint) ([]model.FuturesContract, error) {
	var contracts []model.FuturesContract
	if err := r.db.WithContext(ctx).Where("asset_id IN ?", assetIDs).Preload("Asset").Preload("Listing").Preload("Listing.Exchange").Find(&contracts).Error; err != nil {
		return nil, err
	}
	return contracts, nil
}
