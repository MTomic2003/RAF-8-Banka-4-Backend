package service

import (
	"banking-service/internal/model"
	"context"
)

type CurrencyConverter interface {
	Convert(ctx context.Context, amount float64, from, to model.CurrencyCode) (float64, error)
	CalculateFee(amount float64) float64
}
