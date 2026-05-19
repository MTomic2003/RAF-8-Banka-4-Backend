package repository

import "context"

// TransactionManager runs a unit of work inside a database transaction. The
// resulting *gorm.DB is stashed on the context via db.TxContextKey{}, so
// repositories that look it up with db.DBFromContext automatically join the
// same transaction.
type TransactionManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
