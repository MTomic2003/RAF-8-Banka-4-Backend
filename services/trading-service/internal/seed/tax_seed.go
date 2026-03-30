package seed

import (
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
)

func SeedAccumulatedTax(db *gorm.DB) error {
	now := time.Now()

	records := []model.AccumulatedTax{
		{
			AccumulatedTaxID: 1,
			AccountID:        "444000112345678911",
			TaxOwedRSD:       13000,
			LastUpdatedAt:    now,
		},
		{
			AccumulatedTaxID: 2,
			AccountID:        "444000112345678913",
			TaxOwedRSD:       25000,
			LastUpdatedAt:    now,
		},
		{
			AccumulatedTaxID: 3,
			AccountID:        "444000112345678921",
			TaxOwedRSD:       5000,
			LastUpdatedAt:    now,
		},
		{
			AccumulatedTaxID: 4,
			AccountID:        "444000112345678922",
			TaxOwedRSD:       800000,
			LastUpdatedAt:    now,
		},
	}

	for _, r := range records {
		if err := db.FirstOrCreate(&r, model.AccumulatedTax{AccountID: r.AccountID}).Error; err != nil {
			return err
		}
	}

	return nil
}
