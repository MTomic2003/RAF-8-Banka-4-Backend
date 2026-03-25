package repository

import (
	"context"
	"errors"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
	"gorm.io/gorm"
)

type loanRequestRepository struct {
	db *gorm.DB
}

func NewLoanRequestRepository(db *gorm.DB) LoanRequestRepository {
	return &loanRequestRepository{db: db}
}

func (r *loanRequestRepository) FindAll(ctx context.Context, query *dto.ListLoanRequestsQuery) ([]model.LoanRequest, int64, error) {
	var loans []model.LoanRequest
	var count int64
	db := r.db.WithContext(ctx).Model(&model.LoanRequest{})

	if query.ClientID != 0 {
		db = db.Where("client_id = ?", query.ClientID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("LoanType").
		Limit(query.PageSize).Offset(offset).Find(&loans).Error

	return loans, count, err
}

func (r *loanRequestRepository) FindByID(ctx context.Context, id uint) (*model.LoanRequest, error) {
	var loan model.LoanRequest
	result := r.db.WithContext(ctx).Preload("LoanType").First(&loan, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &loan, nil
}

func (r *loanRequestRepository) Update(ctx context.Context, request *model.LoanRequest) error {
	return r.db.WithContext(ctx).Save(request).Error
}

func (r *loanRequestRepository) CreateRequest(ctx context.Context, request *model.LoanRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}
