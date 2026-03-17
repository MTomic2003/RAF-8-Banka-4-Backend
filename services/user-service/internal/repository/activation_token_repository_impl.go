package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"
)

type activationTokenRepository struct {
	db *gorm.DB
}

func NewActivationTokenRepository(db *gorm.DB) ActivationTokenRepository {
	return &activationTokenRepository{db: db}
}

func (r *activationTokenRepository) Create(ctx context.Context, token *model.ActivationToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *activationTokenRepository) FindByToken(ctx context.Context, token string) (*model.ActivationToken, error) {
	var t model.ActivationToken

	result := r.db.WithContext(ctx).Where("token = ?", token).First(&t)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &t, nil
}

func (r *activationTokenRepository) Delete(ctx context.Context, token *model.ActivationToken) error {
	return r.db.WithContext(ctx).Delete(token).Error
}
