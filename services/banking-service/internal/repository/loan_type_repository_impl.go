package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type loanTypeRepository struct {
	db *gorm.DB
}

func NewLoanTypeRepository(db *gorm.DB) LoanTypeRepository {
	return &loanTypeRepository{db: db}
}

func (r *loanTypeRepository) FindByID(ctx context.Context, id uint) (*model.LoanType, error) {
	var loanType model.LoanType
	result := r.db.WithContext(ctx).First(&loanType, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &loanType, nil
}
