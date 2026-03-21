package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/db"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type transferRepository struct {
	db *gorm.DB
}

func NewTransferRepository(db *gorm.DB) TransferRepository {
	return &transferRepository{db: db}
}

func (r *transferRepository) Create(ctx context.Context, transfer *model.Transfer) error {
	currentDB := db.DBFromContext(ctx, r.db)
	return currentDB.WithContext(ctx).Create(transfer).Error
}

func (r *transferRepository) ListByClientID(ctx context.Context, clientID uint, page, pageSize int) ([]model.Transfer, int64, error) {
	currentDB := db.DBFromContext(ctx, r.db)

	var (
		transfers []model.Transfer
		total     int64
	)

	baseQuery := currentDB.WithContext(ctx).
		Model(&model.Transfer{}).
		Joins("JOIN transactions ON transactions.transaction_id = transfers.transaction_id").
		Joins("JOIN accounts payer_accounts ON payer_accounts.account_number = transactions.payer_account_number").
		Where("payer_accounts.client_id = ?", clientID)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	if err := baseQuery.
		Preload("Transaction").
		Order("transactions.created_at DESC, transactions.transaction_id DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&transfers).Error; err != nil {
		return nil, 0, err
	}

	return transfers, total, nil
}
