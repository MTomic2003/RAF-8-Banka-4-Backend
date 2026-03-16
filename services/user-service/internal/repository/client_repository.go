package repository

import (
	"context"
	"user-service/internal/model"
)

type ClientRepository interface {
	Create(ctx context.Context, client *model.Client) error
	FindByIdentityID(ctx context.Context, identityID uint) (*model.Client, error)
	FindByID(ctx context.Context, id uint) (*model.Client, error)
}
