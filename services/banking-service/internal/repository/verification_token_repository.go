package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/db"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type VerificationTokenRepository interface {
	Create(ctx context.Context, token *model.VerificationToken) error
	FindByAccountAndClient(ctx context.Context, accountNumber string, clientID uint) (*model.VerificationToken, error)
	DeleteByAccountAndClient(ctx context.Context, accountNumber string, clientID uint) error
}

type verificationTokenRepository struct {
	db *gorm.DB
}

func NewVerificationTokenRepository(db *gorm.DB) VerificationTokenRepository {
	return &verificationTokenRepository{db: db}
}

func (r *verificationTokenRepository) Create(ctx context.Context, token *model.VerificationToken) error {
	db := db.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Create(token).Error
}

func (r *verificationTokenRepository) FindByAccountAndClient(ctx context.Context, accountNumber string, clientID uint) (*model.VerificationToken, error) {
	var token model.VerificationToken
	err := r.db.WithContext(ctx).
		Where("account_number = ? AND client_id = ?", accountNumber, clientID).
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &token, nil
}

func (r *verificationTokenRepository) DeleteByAccountAndClient(ctx context.Context, accountNumber string, clientID uint) error {
	db := db.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).
		Where("account_number = ? AND client_id = ?", accountNumber, clientID).
		Delete(&model.VerificationToken{}).Error
}
