package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"
)

type ActivationTokenRepository interface {
	Create(ctx context.Context, token *model.ActivationToken) error
	FindByToken(ctx context.Context, token string) (*model.ActivationToken, error)
	Delete(ctx context.Context, token *model.ActivationToken) error
	DeleteByIdentityID(ctx context.Context, identityID uint) error
}
