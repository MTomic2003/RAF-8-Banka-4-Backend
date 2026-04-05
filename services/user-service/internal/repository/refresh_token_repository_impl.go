package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/db"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	db := db.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepository) FindByToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	var t model.RefreshToken
	result := r.db.WithContext(ctx).Where("token = ?", token).First(&t)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &t, nil
}

func (r *refreshTokenRepository) DeleteByIdentityID(ctx context.Context, identityID uint) error {
	db := db.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Where("identity_id = ?", identityID).Delete(&model.RefreshToken{}).Error
}
