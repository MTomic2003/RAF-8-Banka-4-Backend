package repository

import (
	"banking-service/internal/dto"
	"banking-service/internal/model"
	"context"
)

type AccountRepository interface {
	Create(ctx context.Context, account *model.Account) error
	AccountNumberExists(ctx context.Context, accountNumber string) (bool, error)
	FindByAccountNumber(ctx context.Context, accountNumber string) (*model.Account, error)
	UpdateBalance(ctx context.Context, account *model.Account) error
	FindAll(ctx context.Context, query *dto.ListAccountsQuery) ([]*model.Account, int64, error)
}
