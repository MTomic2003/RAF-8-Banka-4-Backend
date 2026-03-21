package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type AuthorizedPersonRepository interface {
	Create(ctx context.Context, person *model.AuthorizedPerson) error
	FindByID(ctx context.Context, id uint) (*model.AuthorizedPerson, error)
	ListByAccountNumber(ctx context.Context, accountNumber string) ([]model.AuthorizedPerson, error)
}
