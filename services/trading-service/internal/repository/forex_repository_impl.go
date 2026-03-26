package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
)

type forexRepository struct {
	db *gorm.DB
}

func NewForexRepository(db *gorm.DB) ForexRepository {
	return &forexRepository{db: db}
}

func (r *forexRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ForexPair{}).
		Count(&count).Error

	return count, err
}

func (r *forexRepository) Upsert(ctx context.Context, pair model.ForexPair) error {
	return r.db.WithContext(ctx).
		Where("base = ? AND quote = ?", pair.Base, pair.Quote).
		Assign(pair).
		FirstOrCreate(&pair).Error
}
