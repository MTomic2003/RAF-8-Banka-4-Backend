package repository

import (
	"context"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type TaxRepository interface {
	FindAllAccumulatedTax(ctx context.Context, accountNumbers []string, page, pageSize int) ([]model.AccumulatedTax, int64, error)

	FindAccumulatedTaxByAccountNumber(ctx context.Context, accountNumber string) (*model.AccumulatedTax, error)

	FindAllPositiveAccumulatedTax(ctx context.Context) ([]model.AccumulatedTax, error)

	AddTaxOwed(ctx context.Context, accountNumber string, amount float64) error

	ClearTax(ctx context.Context, accountNumber string, clearedAt time.Time) error

	RecordCollectionResult(ctx context.Context, collection *model.TaxCollection, clearTax bool, clearedAmount float64, clearedAt time.Time) error

	SaveAccumulatedTax(ctx context.Context, tax *model.AccumulatedTax) error

	CreateTaxCollection(ctx context.Context, collection *model.TaxCollection) error

	FindTaxCollectionsByAccountNumber(ctx context.Context, accountNumber string) ([]model.TaxCollection, error)

	FindLatestTaxCollection(ctx context.Context, accountNumber string) (*model.TaxCollection, error)
}
