package seed

import (
	"testing"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupSeedTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:testdb_seed_" + time.Now().Format("150405.000") + "?mode=memory&_pragma=foreign_keys(1)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&model.Exchange{}, &model.Listing{}, &model.FuturesContract{}); err != nil {
		t.Fatal(err)
	}

	return db
}

func TestSeedFuturesContracts_NormalizesAndFallsBackToSeededExchangeMICs(t *testing.T) {
	db := setupSeedTestDB(t)

	if err := RunExchangeSeed(db); err != nil {
		t.Fatalf("RunExchangeSeed failed: %v", err)
	}

	if err := SeedFuturesContracts(db); err != nil {
		t.Fatalf("SeedFuturesContracts failed: %v", err)
	}

	var cmeListing model.Listing
	if err := db.Where("ticker = ?", "ZCJ26").First(&cmeListing).Error; err != nil {
		t.Fatalf("failed loading CME-mapped listing: %v", err)
	}

	if cmeListing.ExchangeMIC != "XCME" {
		t.Fatalf("expected ZCJ26 exchange XCME, got %s", cmeListing.ExchangeMIC)
	}

	var fallbackListing model.Listing
	if err := db.Where("ticker = ?", "CLM26").First(&fallbackListing).Error; err != nil {
		t.Fatalf("failed loading fallback listing: %v", err)
	}

	if fallbackListing.ExchangeMIC != model.SimulatedExchangeMIC {
		t.Fatalf("expected CLM26 exchange %s, got %s", model.SimulatedExchangeMIC, fallbackListing.ExchangeMIC)
	}
}
