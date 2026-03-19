package repository

import (
	"banking-service/internal/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type payeeRepository struct {
	db *gorm.DB
}

func NewPayeeRepository(db *gorm.DB) PayeeRepository {
	return &payeeRepository{db: db}
}

func (r *payeeRepository) FindAllByClientID(ctx context.Context, clientID uint) ([]model.Payee, error) {
	var payees []model.Payee
	err := r.db.WithContext(ctx).Where("client_id = ?", clientID).Find(&payees).Error
	return payees, err
}

func (r *payeeRepository) FindByID(ctx context.Context, id uint) (*model.Payee, error) {
	var payee model.Payee
	result := r.db.WithContext(ctx).First(&payee, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &payee, result.Error
}

func (r *payeeRepository) Create(ctx context.Context, payee *model.Payee) error {
	return r.db.WithContext(ctx).Create(payee).Error
}

func (r *payeeRepository) Update(ctx context.Context, payee *model.Payee) error {
	return r.db.WithContext(ctx).Save(payee).Error
}

func (r *payeeRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Payee{}, id).Error
}