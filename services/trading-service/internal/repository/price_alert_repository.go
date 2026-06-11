package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type PriceAlertRepository interface {
	Create(ctx context.Context, alert *model.PriceAlert) error
	FindByID(ctx context.Context, id uint) (*model.PriceAlert, error)
	FindByOwner(ctx context.Context, userID uint, ownerType model.OwnerType) ([]model.PriceAlert, error)
	FindAllActive(ctx context.Context) ([]model.PriceAlert, error)
	MarkTriggered(ctx context.Context, id uint) error
	Delete(ctx context.Context, id uint) error
}
