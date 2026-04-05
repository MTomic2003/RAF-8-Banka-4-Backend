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

func (r *futuresContractRepository) FindByListingIDs(ctx context.Context, listingIDs []uint) ([]model.FuturesContract, error) {
	var contracts []model.FuturesContract
	if err := r.db.WithContext(ctx).Where("listing_id IN ?", listingIDs).Preload("Listing").Find(&contracts).Error; err != nil {
		return nil, err
	}
	return contracts, nil
}
