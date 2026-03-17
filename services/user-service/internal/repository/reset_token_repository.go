package repository

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"
)

type ResetTokenRepository interface {
	Create(ctx context.Context, token *model.ResetToken) error
	FindByToken(ctx context.Context, token string) (*model.ResetToken, error)
	DeleteByIdentityID(ctx context.Context, identityID uint) error
}
