package repository

import (
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
func (r *optionRepository) Upsert(option *model.Option) error {
	return r.db.
		Where(model.Option{ListingID: option.ListingID}).
		Assign(*option).
		FirstOrCreate(option).Error
}
