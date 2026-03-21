package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
)

type CompanyRepository interface {
	Create(ctx context.Context, company *model.Company) error
	WorkCodeExists(ctx context.Context, id uint) (bool, error)
	RegistrationNumberExists(ctx context.Context, registrationNumber string) (bool, error)
	TaxNumberExists(ctx context.Context, taxNumber string) (bool, error)
}
