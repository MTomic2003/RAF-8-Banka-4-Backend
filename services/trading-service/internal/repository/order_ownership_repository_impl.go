package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
)

type orderOwnershipRepository struct {
	db *gorm.DB
}

func NewOrderOwnershipRepository(db *gorm.DB) OrderOwnershipRepository {
	return &orderOwnershipRepository{db: db}
}

func (r *orderOwnershipRepository) FindByIdentity(ctx context.Context, identityID uint, ownerType model.OwnerType) ([]model.OrderOwnership, error) {
	var ownerships []model.OrderOwnership
	if err := r.db.WithContext(ctx).
		Where("identity_id = ? AND owner_type = ?", identityID, ownerType).
		Preload("Order").
		Preload("Order.Listing").
		Find(&ownerships).Error; err != nil {
		return nil, err
	}
	return ownerships, nil
}
