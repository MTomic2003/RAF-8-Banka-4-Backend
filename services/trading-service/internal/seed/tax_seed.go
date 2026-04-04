package seed

import (
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
)

func AccumulatedTax(db *gorm.DB) error {
	now := time.Now()

	emp4 := uint(4) // Marko Markovic
	emp5 := uint(5) // Jelena Jovanovic
	emp6 := uint(6) // Nikola Nikolic

	records := []model.AccumulatedTax{
		{
			AccountNumber: "444000112345678911",
			TaxOwed:       5000,
			CurrencyCode:  "RSD",
			LastUpdatedAt: now,
		},
		{
			AccountNumber: "444000112345678913",
			TaxOwed:       45000,
			CurrencyCode:  "RSD",
			LastUpdatedAt: now,
		},

		{
			AccountNumber: "444000112345678922",
			TaxOwed:       2000,
			CurrencyCode:  "USD",
			LastUpdatedAt: now,
		},
		// Employee 4 (Marko) - traded via RSD, EUR, USD accounts
		{
			AccountNumber: "444000000000000000", // Bank RSD Account
			EmployeeID:    &emp4,
			TaxOwed:       12000,
			CurrencyCode:  "RSD",
			LastUpdatedAt: now,
		},
		{
			AccountNumber: "444000000000000001", // Bank EUR Account
			EmployeeID:    &emp4,
			TaxOwed:       80,
			CurrencyCode:  "EUR",
			LastUpdatedAt: now,
		},
		{
			AccountNumber: "444000000000000002", // Bank USD Account
			EmployeeID:    &emp4,
			TaxOwed:       150,
			CurrencyCode:  "USD",
			LastUpdatedAt: now,
		},

		// Employee 5 (Jelena) - traded via RSD, GBP, CHF accounts
		{
			AccountNumber: "444000000000000000", // Bank RSD Account
			EmployeeID:    &emp5,
			TaxOwed:       30000,
			CurrencyCode:  "RSD",
			LastUpdatedAt: now,
		},
		{
			AccountNumber: "444000000000000004", // Bank GBP Account
			EmployeeID:    &emp5,
			TaxOwed:       60,
			CurrencyCode:  "GBP",
			LastUpdatedAt: now,
		},
		{
			AccountNumber: "444000000000000003", // Bank CHF Account
			EmployeeID:    &emp5,
			TaxOwed:       200,
			CurrencyCode:  "CHF",
			LastUpdatedAt: now,
		},

		// Employee 6 (Nikola) - traded via USD, JPY, CAD, AUD accounts
		{
			AccountNumber: "444000000000000002", // Bank USD Account
			EmployeeID:    &emp6,
			TaxOwed:       500,
			CurrencyCode:  "USD",
			LastUpdatedAt: now,
		},
		{
			AccountNumber: "444000000000000005", // Bank JPY Account
			EmployeeID:    &emp6,
			TaxOwed:       50000,
			CurrencyCode:  "JPY",
			LastUpdatedAt: now,
		},
		{
			AccountNumber: "444000000000000006", // Bank CAD Account
			EmployeeID:    &emp6,
			TaxOwed:       300,
			CurrencyCode:  "CAD",
			LastUpdatedAt: now,
		},
		{
			AccountNumber: "444000000000000007", // Bank AUD Account
			EmployeeID:    &emp6,
			TaxOwed:       400,
			CurrencyCode:  "AUD",
			LastUpdatedAt: now,
		},
	}

	for _, r := range records {
		if err := db.FirstOrCreate(&r, model.AccumulatedTax{
			AccountNumber: r.AccountNumber,
			EmployeeID:    r.EmployeeID,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}
