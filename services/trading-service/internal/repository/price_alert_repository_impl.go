package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

type priceAlertRepository struct {
	db *gorm.DB
}

func NewPriceAlertRepository(db *gorm.DB) PriceAlertRepository {
	return &priceAlertRepository{db: db}
}

func (r *priceAlertRepository) Create(ctx context.Context, alert *model.PriceAlert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

// FindByID preloads Listing + Asset so callers (notifier, handler) have the
// ticker available without an extra query. Returns (nil, nil) when the row
// does not exist, mirroring the convention used by other repositories.
func (r *priceAlertRepository) FindByID(ctx context.Context, id uint) (*model.PriceAlert, error) {
	var alert model.PriceAlert
	err := r.db.WithContext(ctx).
		Preload("Listing").
		Preload("Listing.Asset").
		First(&alert, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &alert, nil
}

func (r *priceAlertRepository) FindByOwner(ctx context.Context, userID uint, ownerType model.OwnerType) ([]model.PriceAlert, error) {
	var alerts []model.PriceAlert
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND owner_type = ?", userID, ownerType).
		Preload("Listing").
		Preload("Listing.Asset").
		Order("created_at DESC, price_alert_id DESC").
		Find(&alerts).Error
	return alerts, err
}

// FindAllActive returns every alert with IsActive=true, preloaded with Listing
// + Asset so the scheduler can read the current price (listing.Price) and the
// ticker for the email body without further round-trips. The set is bounded by
// the number of active alerts in the system, which is expected to be small.
func (r *priceAlertRepository) FindAllActive(ctx context.Context) ([]model.PriceAlert, error) {
	var alerts []model.PriceAlert
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Preload("Listing").
		Preload("Listing.Asset").
		Find(&alerts).Error
	return alerts, err
}

// MarkTriggered atomically flips IsActive to false and stamps TriggeredAt. Done
// in a single UPDATE so the scheduler can call it without re-fetching.
func (r *priceAlertRepository) MarkTriggered(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.PriceAlert{}).
		Where("price_alert_id = ?", id).
		Updates(map[string]any{
			"is_active":    false,
			"triggered_at": &now,
		}).Error
}

func (r *priceAlertRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.PriceAlert{}, id).Error
}
