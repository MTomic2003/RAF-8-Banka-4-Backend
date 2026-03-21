package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type CurrencyRepository interface {
	FindByCode(ctx context.Context, code model.CurrencyCode) (*model.Currency, error)
}
