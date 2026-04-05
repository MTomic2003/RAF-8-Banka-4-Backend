package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/db"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"
)

type identityRepository struct {
	db *gorm.DB
}

func NewIdentityRepository(db *gorm.DB) IdentityRepository {
	return &identityRepository{db: db}
}

func (r *identityRepository) Create(ctx context.Context, identity *model.Identity) error {
	db := db.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Create(identity).Error
}

func (r *identityRepository) FindByID(ctx context.Context, id uint) (*model.Identity, error) {
	var identity model.Identity
	result := r.db.WithContext(ctx).First(&identity, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &identity, result.Error
}

func (r *identityRepository) FindByEmail(ctx context.Context, email string) (*model.Identity, error) {
	var identity model.Identity
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&identity)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &identity, result.Error
}

func (r *identityRepository) FindByUsername(ctx context.Context, username string) (*model.Identity, error) {
	var identity model.Identity
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&identity)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &identity, result.Error
}

func (r *identityRepository) Update(ctx context.Context, identity *model.Identity) error {
	db := db.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Save(identity).Error
}

func (r *identityRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.Identity{}).
		Where("email = ?", email).
		Count(&count).
		Error

	return count > 0, err
}

func (r *identityRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.Identity{}).
		Where("username = ?", username).
		Count(&count).
		Error

	return count > 0, err
}
