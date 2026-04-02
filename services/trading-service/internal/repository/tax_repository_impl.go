package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type taxRepositoryImpl struct {
	db *gorm.DB
}

func NewTaxRepository(db *gorm.DB) TaxRepository {
	return &taxRepositoryImpl{db: db}
}

func (r *taxRepositoryImpl) FindAllAccumulatedTax(ctx context.Context, accountNumbers []string, page, pageSize int) ([]model.AccumulatedTax, int64, error) {
	var taxes []model.AccumulatedTax
	var count int64

	query := r.db.WithContext(ctx).Model(&model.AccumulatedTax{})

	if len(accountNumbers) > 0 {
		query = query.Where("account_number IN ?", accountNumbers)
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Limit(pageSize).Offset(offset).Order("accumulated_tax_id ASC").Find(&taxes).Error
	return taxes, count, err
}

func (r *taxRepositoryImpl) FindAccumulatedTaxByAccountNumber(ctx context.Context, accountNumber string) (*model.AccumulatedTax, error) {
	var tax model.AccumulatedTax
	result := r.db.WithContext(ctx).Where("account_number = ?", accountNumber).First(&tax)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &tax, result.Error
}

func (r *taxRepositoryImpl) FindAllPositiveAccumulatedTax(ctx context.Context) ([]model.AccumulatedTax, error) {
	var taxes []model.AccumulatedTax
	err := r.db.WithContext(ctx).Where("tax_owed_rsd > 0").Find(&taxes).Error
	if err != nil {
		return nil, err
	}
	return taxes, nil
}

func (r *taxRepositoryImpl) AddTaxOwed(ctx context.Context, accountNumber string, amount float64) error {
	result := r.db.WithContext(ctx).
		Model(&model.AccumulatedTax{}).
		Where("account_number = ?", accountNumber).
		Updates(map[string]interface{}{
			"tax_owed_rsd":    gorm.Expr("tax_owed_rsd + ?", amount),
			"last_updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		tax := model.AccumulatedTax{
			AccountNumber: accountNumber,
			TaxOwedRSD:    amount,
			LastUpdatedAt: time.Now(),
		}
		return r.db.WithContext(ctx).Create(&tax).Error
	}

	return nil
}

func (r *taxRepositoryImpl) ClearTax(ctx context.Context, accountNumber string, clearedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.AccumulatedTax{}).
		Where("account_number = ?", accountNumber).
		Updates(map[string]interface{}{
			"tax_owed_rsd":    0,
			"last_cleared_at": clearedAt,
			"last_updated_at": clearedAt,
		}).Error
}

func (r *taxRepositoryImpl) RecordCollectionResult(ctx context.Context, collection *model.TaxCollection, clearTax bool, clearedAmount float64, clearedAt time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(collection).Error; err != nil {
			return err
		}
		if clearTax {
			if err := tx.Model(&model.AccumulatedTax{}).
				Where("account_number = ?", collection.AccountNumber).
				Updates(map[string]interface{}{
					// Subtract only what was collected, not zero out
					"tax_owed_rsd":    gorm.Expr("GREATEST(tax_owed_rsd - ?, 0)", clearedAmount),
					"last_cleared_at": clearedAt,
					"last_updated_at": clearedAt,
				}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *taxRepositoryImpl) SaveAccumulatedTax(ctx context.Context, tax *model.AccumulatedTax) error {
	return r.db.WithContext(ctx).Save(tax).Error
}

func (r *taxRepositoryImpl) CreateTaxCollection(ctx context.Context, collection *model.TaxCollection) error {
	return r.db.WithContext(ctx).Create(collection).Error
}

func (r *taxRepositoryImpl) FindTaxCollectionsByAccountNumber(ctx context.Context, accountNumber string) ([]model.TaxCollection, error) {
	var collections []model.TaxCollection
	err := r.db.WithContext(ctx).
		Where("account_number = ?", accountNumber).
		Order("tax_collection_id DESC").
		Find(&collections).Error
	if err != nil {
		return nil, err
	}
	return collections, nil
}

func (r *taxRepositoryImpl) FindLatestTaxCollection(ctx context.Context, accountNumber string) (*model.TaxCollection, error) {
	var collection model.TaxCollection
	result := r.db.WithContext(ctx).
		Where("account_number = ?", accountNumber).
		Order("tax_collection_id DESC").
		First(&collection)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &collection, result.Error
}
