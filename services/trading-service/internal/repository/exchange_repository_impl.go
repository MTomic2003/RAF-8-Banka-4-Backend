package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type exchangeRepositoryImpl struct {
	db *gorm.DB
}

func NewExchangeRepository(db *gorm.DB) ExchangeRepository {
	return &exchangeRepositoryImpl{db: db}
}

func (r *exchangeRepositoryImpl) FindAll(ctx context.Context, page, pageSize int) ([]model.Exchange, int64, error) {
	var exchanges []model.Exchange
	var count int64

	db := r.db.WithContext(ctx).Model(&model.Exchange{})

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Limit(pageSize).Offset(offset).Find(&exchanges).Error
	return exchanges, count, err
}

func (r *exchangeRepositoryImpl) FindByMicCode(ctx context.Context, micCode string) (*model.Exchange, error) {
	var exchange model.Exchange
	result := r.db.WithContext(ctx).Where("mic_code = ?", micCode).First(&exchange)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &exchange, result.Error
}

func (r *exchangeRepositoryImpl) ToggleTradingEnabled(ctx context.Context, micCode string) (*model.Exchange, error) {
	exchange, err := r.FindByMicCode(ctx, micCode)
	if err != nil {
		return nil, err
	}
	if exchange == nil {
		return nil, nil
	}

	exchange.TradingEnabled = !exchange.TradingEnabled
	result := r.db.WithContext(ctx).Save(exchange)
	return exchange, result.Error
}
