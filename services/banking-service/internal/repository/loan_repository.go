package repository

import (
	"context"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type LoanRepository interface {
	FindByClientID(ctx context.Context, clientID uint, sortByAmountDesc bool) ([]model.Loan, error)
	FindByIDAndClientID(ctx context.Context, id uint, clientID uint) (*model.Loan, error)

	// Metode za aktivne kredite
	CreateLoan(ctx context.Context, loan *model.Loan) error
	FindLoanByRequestID(ctx context.Context, requestID uint) (*model.Loan, error)
	UpdateLoan(ctx context.Context, loan *model.Loan) error

	// Metode za rate
	CreateInstallments(ctx context.Context, installments []model.LoanInstallment) error
	FindDueInstallments(ctx context.Context, date time.Time) ([]model.LoanInstallment, error)
	FindRetryInstallments(ctx context.Context, now time.Time) ([]model.LoanInstallment, error)
	UpdateInstallment(ctx context.Context, installment *model.LoanInstallment) error

	// Vraca sve aktivne kredite sa varijabilnom kamatnom stopom
	FindActiveVariableRateLoans(ctx context.Context) ([]model.Loan, error)
}
