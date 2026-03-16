package service

import (
	"banking-service/internal/client"
	"banking-service/internal/dto"
	"banking-service/internal/model"
	"banking-service/internal/repository"
	"common/pkg/errors"
	"context"

	"gorm.io/gorm"
)

type CompanyService struct {
	repo       repository.CompanyRepository
	userClient client.UserClient
	db         *gorm.DB
}

func NewCompanyService(
	repo repository.CompanyRepository,
	userClient client.UserClient,
	db *gorm.DB,
) *CompanyService {
	return &CompanyService{
		repo:       repo,
		userClient: userClient,
		db:         db,
	}
}

func (s *CompanyService) workCodeExists(ctx context.Context, id uint) (bool, error) {
	var count int64

	err := s.db.WithContext(ctx).
		Model(&model.WorkCode{}).
		Where("work_code_id = ?", id).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *CompanyService) registrationNumberExists(ctx context.Context, registrationNumber string) (bool, error) {
	var count int64

	err := s.db.WithContext(ctx).
		Model(&model.Company{}).
		Where("registration_number = ?", registrationNumber).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *CompanyService) taxNumberExists(ctx context.Context, taxNumber string) (bool, error) {
	var count int64

	err := s.db.WithContext(ctx).
		Model(&model.Company{}).
		Where("tax_number = ?", taxNumber).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *CompanyService) Create(ctx context.Context, req dto.CreateCompanyRequest) (*model.Company, error) {
	if _, err := s.userClient.GetClientByID(ctx, req.OwnerID); err != nil {
		return nil, errors.NotFoundErr("owner client not found")
	}

	workCodeExists, err := s.workCodeExists(ctx, req.WorkCodeID)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	if !workCodeExists {
		return nil, errors.NotFoundErr("work code not found")
	}

	regNumExists, err := s.registrationNumberExists(ctx, req.RegistrationNumber)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	if regNumExists {
		return nil, errors.ConflictErr("registration number already exists")
	}

	taxNumExists, err := s.taxNumberExists(ctx, req.TaxNumber)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	if taxNumExists {
		return nil, errors.ConflictErr("tax number already exists")
	}

	company := &model.Company{
		Name:               req.Name,
		RegistrationNumber: req.RegistrationNumber,
		TaxNumber:          req.TaxNumber,
		WorkCodeID:         req.WorkCodeID,
		Address:            req.Address,
		OwnerID:            req.OwnerID,
	}

	if err := s.repo.Create(ctx, company); err != nil {
		return nil, errors.InternalErr(err)
	}

	return company, nil
}
