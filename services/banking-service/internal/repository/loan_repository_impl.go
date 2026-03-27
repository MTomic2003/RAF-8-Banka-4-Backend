package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/db"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type loanRepository struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) LoanRepository {
	return &loanRepository{db: db}
}

func (r *loanRepository) CreateLoan(ctx context.Context, loan *model.Loan) error {
	db := db.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Create(loan).Error
}

func (r *loanRepository) FindByClientID(ctx context.Context, clientID uint, sortByAmountDesc bool) ([]model.Loan, error) {
	var loans []model.Loan
	query := r.db.WithContext(ctx).
		Joins("JOIN loan_requests ON loan_requests.id = loans.loan_request_id").
		Where("loan_requests.client_id = ?", clientID).
		Preload("LoanRequest.LoanType")

	if sortByAmountDesc {
		query = query.Order("loan_requests.amount DESC")
	} else {
		query = query.Order("loan_requests.amount ASC")
	}

	if err := query.Find(&loans).Error; err != nil {
		return nil, err
	}
	return loans, nil
}

func (r *loanRepository) FindByIDAndClientID(ctx context.Context, id uint, clientID uint) (*model.Loan, error) {
	var loan model.Loan
	err := r.db.WithContext(ctx).
		Joins("JOIN loan_requests ON loan_requests.id = loans.loan_request_id").
		Where("loans.id = ? AND loan_requests.client_id = ?", id, clientID).
		Preload("LoanRequest.LoanType").
		First(&loan).Error
	if err != nil {
		return nil, err
	}
	return &loan, nil
}

func (r *loanRepository) FindLoanByRequestID(ctx context.Context, requestID uint) (*model.Loan, error) {
	var loan model.Loan
	err := r.db.WithContext(ctx).
		Preload("Installments").
		Where("loan_request_id = ?", requestID).
		First(&loan).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &loan, err
}

func (r *loanRepository) UpdateLoan(ctx context.Context, loan *model.Loan) error {
	db := db.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Save(loan).Error
}

func (r *loanRepository) CreateInstallments(ctx context.Context, installments []model.LoanInstallment) error {
	db := db.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Create(&installments).Error
}

func (r *loanRepository) FindDueInstallments(ctx context.Context, date time.Time) ([]model.LoanInstallment, error) {
	var installments []model.LoanInstallment
	err := r.db.WithContext(ctx).
		Preload("Loan.LoanRequest").
		Where("status = ? AND due_date <= ?", model.InstallmentStatusPending, date).
		Find(&installments).Error
	return installments, err
}

func (r *loanRepository) FindRetryInstallments(ctx context.Context, now time.Time) ([]model.LoanInstallment, error) {
	var installments []model.LoanInstallment
	err := r.db.WithContext(ctx).
		Preload("Loan.LoanRequest").
		Where("status = ? AND retry_at <= ?", model.InstallmentStatusRetrying, now).
		Find(&installments).Error
	return installments, err
}

func (r *loanRepository) UpdateInstallment(ctx context.Context, installment *model.LoanInstallment) error {
	db := db.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Save(installment).Error
}

func (r *loanRepository) FindActiveVariableRateLoans(ctx context.Context) ([]model.Loan, error) {
	var loans []model.Loan
	err := r.db.WithContext(ctx).
		Preload("Installments").
		Where("status = ? AND is_variable_rate = ?", model.LoanStatusActive, true).
		Find(&loans).Error
	return loans, err
}
