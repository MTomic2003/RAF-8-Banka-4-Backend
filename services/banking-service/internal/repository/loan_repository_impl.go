package repository

import (
	"banking-service/internal/dto"
	"banking-service/internal/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type loanRepository struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) LoanRepository {
	return &loanRepository{db: db}
}

func (r *loanRepository) FindByClientID(ctx context.Context, clientID uint, sortByAmountDesc bool) ([]model.LoanRequest, error) {
	var loans []model.LoanRequest

	query := r.db.WithContext(ctx).Where("client_id = ?", clientID).Preload("LoanType")

	if sortByAmountDesc {
		query = query.Order("amount DESC")
	} else {
		query = query.Order("amount ASC")
	}

	if err := query.Find(&loans).Error; err != nil {
		return nil, err
	}
	return loans, nil
}

func (r *loanRepository) FindByIDAndClientID(ctx context.Context, id uint, clientID uint) (*model.LoanRequest, error) {
	var loan model.LoanRequest
	if err := r.db.WithContext(ctx).Where("id = ? AND client_id = ?", id, clientID).Preload("LoanType").First(&loan).Error; err != nil {
		return nil, err
	}
	return &loan, nil
}

func (r *loanRepository) CreateRequest(ctx context.Context, request *model.LoanRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}
func (r *loanRepository) FindAll(ctx context.Context, query *dto.ListLoanRequestsQuery) ([]model.LoanRequest, int64, error) {
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
func (r *loanRepository) FindByID(ctx context.Context, id uint) (*model.LoanRequest, error) {
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

func (r *loanRepository) Update(ctx context.Context, request *model.LoanRequest) error {
	return r.db.WithContext(ctx).Save(request).Error
}