package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type PayeeRepository interface {
	FindAllByClientID(ctx context.Context, clientID uint) ([]model.Payee, error)
	FindByID(ctx context.Context, id uint) (*model.Payee, error)
	Create(ctx context.Context, payee *model.Payee) error
	Update(ctx context.Context, payee *model.Payee) error
	Delete(ctx context.Context, id uint) error
}
