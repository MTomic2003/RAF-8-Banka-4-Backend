package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type currencyRepository struct {
	db *gorm.DB
}

func NewCurrencyRepository(db *gorm.DB) CurrencyRepository {
	return &currencyRepository{db: db}
}

func (r *currencyRepository) FindByCode(ctx context.Context, code model.CurrencyCode) (*model.Currency, error) {
	var currency model.Currency
	result := r.db.WithContext(ctx).Where("code = ?", code).First(&currency)
	if result.Error != nil {
		return nil, errors.NotFoundErr("currency not found: " + string(code))
	}
	return &currency, nil
}
