package repository

import (
	"banking-service/internal/model"
	"context"
)

type TransferRepository interface {
	Create(ctx context.Context, transfer *model.Transfer) error
	ListByClientID(ctx context.Context, clientID uint, page, pageSize int) ([]model.Transfer, int64, error)
}
