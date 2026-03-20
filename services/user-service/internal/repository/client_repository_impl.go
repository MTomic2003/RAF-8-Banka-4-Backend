package repository

import (
	"context"
	"errors"
	"user-service/internal/dto"
	"user-service/internal/model"

	"gorm.io/gorm"
)

type clientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) ClientRepository {
	return &clientRepository{db: db}
}

func (r *clientRepository) Create(ctx context.Context, client *model.Client) error {
	return r.db.WithContext(ctx).Create(client).Error
}

func (r *clientRepository) FindByIdentityID(ctx context.Context, identityID uint) (*model.Client, error) {
	var c model.Client

	result := r.db.WithContext(ctx).
		Preload("Identity").
		Where("identity_id = ?", identityID).
		First(&c)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if result.Error != nil {
		return nil, result.Error
	}

	return &c, nil
}
func (r *clientRepository) FindByID(ctx context.Context, id uint) (*model.Client, error) {
	var c model.Client
	result := r.db.WithContext(ctx).
		Preload("Identity").
		First(&c, id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &c, nil
}
func (r *clientRepository) FindAll(ctx context.Context, query *dto.ListClientsQuery) ([]model.Client, int64, error) {
	var clients []model.Client
	var count int64

	db := r.db.WithContext(ctx).Model(&model.Client{})

	// Filteri
	if query.Email != "" {
		db = db.Joins("JOIN identities ON identities.id = clients.identity_id").
			Where("identities.email ILIKE ?", "%"+query.Email+"%")
	}
	if query.FirstName != "" {
		db = db.Where("first_name ILIKE ?", "%"+query.FirstName+"%")
	}
	if query.LastName != "" {
		db = db.Where("last_name ILIKE ?", "%"+query.LastName+"%")
	}

	// Ukupan broj
	db.Count(&count)
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	// Paginacija
	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("Identity").Limit(query.PageSize).Offset(offset).Find(&clients).Error

	return clients, count, err
}
func (r *clientRepository) Update(ctx context.Context, client *model.Client) error {
	return r.db.WithContext(ctx).Save(client).Error
}
