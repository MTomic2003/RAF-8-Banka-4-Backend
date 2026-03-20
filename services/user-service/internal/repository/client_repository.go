package repository

import (
	"context"
	"user-service/internal/dto"
	"user-service/internal/model"
)

type ClientRepository interface {
	Create(ctx context.Context, client *model.Client) error
	FindByIdentityID(ctx context.Context, identityID uint) (*model.Client, error)
	FindByID(ctx context.Context, id uint) (*model.Client, error)
	FindAll(ctx context.Context, query *dto.ListClientsQuery) ([]model.Client, int64, error)
	Update(ctx context.Context, client *model.Client) error
}
