package service

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type CurrencyConverter interface {
	Convert(ctx context.Context, amount float64, from, to model.CurrencyCode) (float64, error)
	CalculateFee(amount float64) float64
}
