package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"
)

type EmployeeRepository interface {
	Create(ctx context.Context, employee *model.Employee) error
	FindByID(ctx context.Context, id uint) (*model.Employee, error)
	FindByIdentityID(ctx context.Context, identityID uint) (*model.Employee, error)
	Update(ctx context.Context, employee *model.Employee) error
	GetAll(ctx context.Context, email, firstName, lastName, position string, page, pageSize int) ([]model.Employee, int64, error)
}
