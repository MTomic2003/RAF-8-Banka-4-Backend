package repository

import (
	"context"
	"errors"
	"user-service/internal/model"

	"gorm.io/gorm"
)

type clientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) ClientRepository {
	return &clientRepository{db: db}
}

func (r *clientRepository) Create(ctx context.Context, client *model.Client) error {
	return r.db.WithContext(ctx).Create(client).Error
}

func (r *clientRepository) FindByIdentityID(ctx context.Context, identityID uint) (*model.Client, error) {
	var c model.Client

	result := r.db.WithContext(ctx).
		Preload("Identity").
		Where("identity_id = ?", identityID).
		First(&c)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &c, nil
}
