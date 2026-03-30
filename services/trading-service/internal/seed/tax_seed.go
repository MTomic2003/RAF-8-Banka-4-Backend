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
			AccountNumber:    "444000112345678911",
			TaxOwedRSD:       13000,
			LastUpdatedAt:    now,
		},
		{
			AccumulatedTaxID: 2,
			AccountNumber:    "444000112345678913",
			TaxOwedRSD:       25000,
			LastUpdatedAt:    now,
		},
		{
			AccumulatedTaxID: 3,
			AccountNumber:    "444000112345678921",
			TaxOwedRSD:       5000,
			LastUpdatedAt:    now,
		},
		{
			AccumulatedTaxID: 4,
			AccountNumber:    "444000112345678922",
			TaxOwedRSD:       800000,
			LastUpdatedAt:    now,
		},
	}

	for _, r := range records {
		if err := db.FirstOrCreate(&r, model.AccumulatedTax{AccountNumber: r.AccountNumber}).Error; err != nil {
			return err
		}
	}

	return nil
}
