package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"
)

type IdentityRepository interface {
	Create(ctx context.Context, identity *model.Identity) error
	FindByID(ctx context.Context, id uint) (*model.Identity, error)
	FindByEmail(ctx context.Context, email string) (*model.Identity, error)
	FindByUsername(ctx context.Context, username string) (*model.Identity, error)
	Update(ctx context.Context, identity *model.Identity) error
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
}
