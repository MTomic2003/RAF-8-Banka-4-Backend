package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type orderRepositoryImpl struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepositoryImpl{db: db}
}

func (r *orderRepositoryImpl) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepositoryImpl) FindByID(ctx context.Context, id uint) (*model.Order, error) {
	var order model.Order
	result := r.db.WithContext(ctx).Preload("Listing").First(&order, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &order, result.Error
}

func (r *orderRepositoryImpl) Save(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderRepositoryImpl) FindAll(ctx context.Context, page, pageSize int, userID *uint, status *model.OrderStatus, direction *model.OrderDirection, isDone *bool) ([]model.Order, int64, error) {
	var orders []model.Order
	var count int64

	db := r.db.WithContext(ctx).Model(&model.Order{})

	if userID != nil {
		db = db.Where("user_id = ?", *userID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if direction != nil {
		db = db.Where("direction = ?", *direction)
	}
	if isDone != nil {
		db = db.Where("is_done = ?", *isDone)
	}

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Preload("Listing").Limit(pageSize).Offset(offset).Find(&orders).Error
	return orders, count, err
}
