package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type LoanTypeRepository interface {
	FindByID(ctx context.Context, id uint) (*model.LoanType, error)
}
