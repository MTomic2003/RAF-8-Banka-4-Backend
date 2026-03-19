package repository

import (
	"common/pkg/db"
	"context"

	"gorm.io/gorm"
)

type GormTransactionManager struct {
	db *gorm.DB
}

func NewGormTransactionManager(db *gorm.DB) TransactionManager {
	return &GormTransactionManager{db: db}
}

func (m *GormTransactionManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if existingTx, ok := ctx.Value(db.TxContextKey{}).(*gorm.DB); ok && existingTx != nil {
		return fn(ctx)
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, db.TxContextKey{}, tx)
		return fn(txCtx)
	})
}
