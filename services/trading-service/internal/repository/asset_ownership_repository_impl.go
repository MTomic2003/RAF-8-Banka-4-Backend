package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type assetOwnershipRepository struct {
	db *gorm.DB
}

func NewAssetOwnershipRepository(db *gorm.DB) AssetOwnershipRepository {
	return &assetOwnershipRepository{db: db}
}

func (r *assetOwnershipRepository) FindByIdentity(ctx context.Context, identityID uint, ownerType model.OwnerType) ([]model.AssetOwnership, error) {
	var ownerships []model.AssetOwnership
	if err := r.db.WithContext(ctx).
		Where("identity_id = ? AND owner_type = ?", identityID, ownerType).
		Preload("Asset").
		Find(&ownerships).Error; err != nil {
		return nil, err
	}
	return ownerships, nil
}

func (r *assetOwnershipRepository) Upsert(ctx context.Context, ownership *model.AssetOwnership) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "identity_id"}, {Name: "owner_type"}, {Name: "asset_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"amount", "avg_buy_price_rsd", "updated_at"}),
		}).
		Create(ownership).Error
}
