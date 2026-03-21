package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *model.Transaction) error
	GetByID(ctx context.Context, transactionID uint) (*model.Transaction, error)
	Update(ctx context.Context, transaction *model.Transaction) error
	GetByPayerAccountNumber(ctx context.Context, accountNumber string) ([]*model.Transaction, error)
	GetByRecipientAccountNumber(ctx context.Context, accountNumber string) ([]*model.Transaction, error)
}
