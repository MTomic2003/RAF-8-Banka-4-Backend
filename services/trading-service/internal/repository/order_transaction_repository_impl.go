package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type orderTransactionRepositoryImpl struct {
	db *gorm.DB
}

func NewOrderTransactionRepository(db *gorm.DB) OrderTransactionRepository {
	return &orderTransactionRepositoryImpl{db: db}
}

func (r *orderTransactionRepositoryImpl) Create(ctx context.Context, orderTransaction *model.OrderTransaction) error {
	return r.db.WithContext(ctx).Create(orderTransaction).Error
}
