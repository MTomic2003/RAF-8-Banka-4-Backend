package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type companyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}

func (r *companyRepository) Create(ctx context.Context, company *model.Company) error {
	return r.db.WithContext(ctx).Create(company).Error
}

func (r *companyRepository) WorkCodeExists(ctx context.Context, id uint) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.WorkCode{}).
		Where("work_code_id = ?", id).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *companyRepository) RegistrationNumberExists(ctx context.Context, registrationNumber string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.Company{}).
		Where("registration_number = ?", registrationNumber).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *companyRepository) TaxNumberExists(ctx context.Context, taxNumber string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.Company{}).
		Where("tax_number = ?", taxNumber).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
