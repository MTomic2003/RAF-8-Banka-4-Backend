package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/db"
)

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(database *gorm.DB) TransactionManager {
	return &GormTransactionManager{db: database}
}

func (m *GormTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if existing, ok := ctx.Value(db.TxContextKey{}).(*gorm.DB); ok && existing != nil {
		return fn(ctx)
	}
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, db.TxContextKey{}, tx)
		return fn(txCtx)
	})
}
