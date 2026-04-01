package repository

import (
	"context"
	"errors"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
)

type listingRepository struct {
	db *gorm.DB
}

func NewListingRepository(db *gorm.DB) ListingRepository {
	return &listingRepository{db: db}
}

func (r *listingRepository) FindAll(ctx context.Context) ([]model.Listing, error) {
	var listings []model.Listing
	if err := r.db.WithContext(ctx).Find(&listings).Error; err != nil {
		return nil, err
	}
	return listings, nil
}

func (r *listingRepository) FindByID(ctx context.Context, id uint) (*model.Listing, error) {
	var listing model.Listing
	result := r.db.WithContext(ctx).First(&listing, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &listing, result.Error
}

func (r *listingRepository) Upsert(ctx context.Context, listing *model.Listing) error {
	return r.db.WithContext(ctx).
		Where(model.Listing{Ticker: listing.Ticker}).
		Assign(*listing).
		FirstOrCreate(listing).Error
}

func (r *listingRepository) UpdatePriceAndAsk(ctx context.Context, listing *model.Listing, price, ask float64) error {
	return r.db.WithContext(ctx).Model(listing).Updates(map[string]interface{}{
		"price":        price,
		"ask":          ask,
		"last_refresh": time.Now(),
	}).Error
}

func (r *listingRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Listing{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func joinLatestDaily(q *gorm.DB) *gorm.DB {
	return q.Joins(`
		LEFT JOIN listing_daily_price_infos AS ldpi
		ON ldpi.listing_id = listings.listing_id
		AND ldpi.date = (
			SELECT MAX(d.date)
			FROM listing_daily_price_infos d
			WHERE d.listing_id = listings.listing_id
		)
	`)
}

func applyListingFilters(q *gorm.DB, filter ListingFilter) *gorm.DB {
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		q = q.Where("listings.ticker LIKE ? OR listings.name LIKE ?", like, like)
	}
	if filter.Exchange != "" {
		q = q.Where("listings.exchange_mic LIKE ?", filter.Exchange+"%")
	}
	if filter.PriceMin > 0 {
		q = q.Where("listings.price >= ?", filter.PriceMin)
	}
	if filter.PriceMax > 0 {
		q = q.Where("listings.price <= ?", filter.PriceMax)
	}
	if filter.AskMin > 0 {
		q = q.Where("listings.ask >= ?", filter.AskMin)
	}
	if filter.AskMax > 0 {
		q = q.Where("listings.ask <= ?", filter.AskMax)
	}
	// FIX: bid i volume su u listing_daily_price_infos, ne u listings
	if filter.BidMin > 0 {
		q = q.Where("ldpi.bid >= ?", filter.BidMin)
	}
	if filter.BidMax > 0 {
		q = q.Where("ldpi.bid <= ?", filter.BidMax)
	}
	if filter.VolumeMin > 0 {
		q = q.Where("ldpi.volume >= ?", filter.VolumeMin)
	}
	if filter.VolumeMax > 0 {
		q = q.Where("ldpi.volume <= ?", filter.VolumeMax)
	}
	return q
}

func sortColumn(filter ListingFilter) string {
	col := "listings.price"
	switch filter.SortBy {
	case "volume":
		col = "ldpi.volume"
	case "maintenance_margin":
		col = "listings.maintenance_margin"
	}
	dir := "ASC"
	if filter.SortDir == "desc" {
		dir = "DESC"
	}
	return col + " " + dir
}

func (r *listingRepository) FindStocks(ctx context.Context, filter ListingFilter) ([]model.Listing, int64, error) {
	var listings []model.Listing
	var count int64

	db := r.db.WithContext(ctx).
		Model(&model.Listing{}).
		Joins("INNER JOIN stocks ON stocks.listing_id = listings.listing_id")

	db = joinLatestDaily(db)
	db = applyListingFilters(db, filter)

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := db.
		Preload("Stock").
		Preload("DailyPriceInfos", func(db *gorm.DB) *gorm.DB {
			return db.Order("date DESC").Limit(1)
		}).
		Order(sortColumn(filter)).
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&listings).Error

	return listings, count, err
}

func (r *listingRepository) FindFutures(ctx context.Context, filter ListingFilter) ([]model.Listing, int64, error) {
	var listings []model.Listing
	var count int64

	db := r.db.WithContext(ctx).
		Model(&model.Listing{}).
		Joins("INNER JOIN futures_contracts ON futures_contracts.listing_id = listings.listing_id")

	db = joinLatestDaily(db)
	db = applyListingFilters(db, filter)

	// FIX: zamena PostgreSQL-specifičnog ::date cast-a sa date range
	if filter.SettlementDate != nil {
		start := filter.SettlementDate.Truncate(24 * time.Hour)
		end := start.Add(24 * time.Hour)
		db = db.Where("futures_contracts.settlement_date >= ? AND futures_contracts.settlement_date < ?", start, end)
	}

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := db.
		Preload("DailyPriceInfos", func(db *gorm.DB) *gorm.DB {
			return db.Order("date DESC").Limit(1)
		}).
		Order(sortColumn(filter)).
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&listings).Error

	return listings, count, err
}

func (r *listingRepository) FindOptions(ctx context.Context, filter ListingFilter) ([]model.Listing, int64, error) {
	var listings []model.Listing
	var count int64

	db := r.db.WithContext(ctx).
		Model(&model.Listing{}).
		Joins("INNER JOIN options ON options.listing_id = listings.listing_id")

	db = joinLatestDaily(db)
	db = applyListingFilters(db, filter)

	if filter.SettlementDate != nil {
		start := filter.SettlementDate.Truncate(24 * time.Hour)
		end := start.Add(24 * time.Hour)
		db = db.Where("options.settlement_date >= ? AND options.settlement_date < ?", start, end)
	}

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := db.
		Preload("DailyPriceInfos", func(db *gorm.DB) *gorm.DB {
			return db.Order("date DESC").Limit(1)
		}).
		Order(sortColumn(filter)).
		Limit(filter.PageSize).
		Offset((filter.Page - 1) * filter.PageSize).
		Find(&listings).Error

	return listings, count, err
}
