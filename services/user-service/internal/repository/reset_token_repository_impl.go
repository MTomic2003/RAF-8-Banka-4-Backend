package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/db"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"
)

type resetTokenRepository struct {
	db *gorm.DB
}

func NewResetTokenRepository(db *gorm.DB) ResetTokenRepository {
	return &resetTokenRepository{db: db}
}

func (r *resetTokenRepository) Create(ctx context.Context, token *model.ResetToken) error {
	db := db.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Create(token).Error
}

func (r *resetTokenRepository) FindByToken(ctx context.Context, token string) (*model.ResetToken, error) {
	var t model.ResetToken
	result := r.db.WithContext(ctx).Where("token = ?", token).First(&t)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &t, nil
}

func (r *resetTokenRepository) DeleteByIdentityID(ctx context.Context, identityID uint) error {
	db := db.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Where("identity_id = ?", identityID).Delete(&model.ResetToken{}).Error
}
