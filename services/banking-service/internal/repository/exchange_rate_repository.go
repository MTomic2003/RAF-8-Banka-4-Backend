package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type ExchangeRateRepository interface {
	UpsertAll(ctx context.Context, rates []model.ExchangeRate) error
	GetAll(ctx context.Context) ([]model.ExchangeRate, error)
}
