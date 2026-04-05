package seed

import (
	"strings"

	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

func NormalizeExchangeMIC(raw string) string {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	if normalized == "" {
		return ""
	}

	switch {
	case normalized == "FOREX":
		return model.SimulatedExchangeMIC
	case normalized == "CME":
		return "XCME"
	case normalized == "NYSE" || strings.Contains(normalized, "NEW YORK STOCK EXCHANGE"):
		return "XNYS"
	case strings.Contains(normalized, "NASDAQ"):
		return "XNAS"
	default:
		return normalized
	}
}

func resolveExistingExchangeMIC(db *gorm.DB, raw string) string {
	mic := NormalizeExchangeMIC(raw)
	if mic == "" {
		return model.SimulatedExchangeMIC
	}

	if exchangeExists(db, mic) {
		return mic
	}

	return model.SimulatedExchangeMIC
}

func exchangeExists(db *gorm.DB, micCode string) bool {
	var count int64
	if err := db.Model(&model.Exchange{}).Where("mic_code = ?", micCode).Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

func NormalizeListingExchangeMICs(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.Listing{}) {
		return nil
	}

	var listings []model.Listing
	if err := db.Find(&listings).Error; err != nil {
		return err
	}

	for _, listing := range listings {
		resolved := resolveExistingExchangeMIC(db, listing.ExchangeMIC)
		if listing.ExchangeMIC == resolved {
			continue
		}

		if err := db.Model(&model.Listing{}).
			Where("listing_id = ?", listing.ListingID).
			Update("exchange_mic", resolved).Error; err != nil {
			return err
		}
	}

	return nil
}
