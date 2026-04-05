package service

import (
	"context"
	"testing"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- Mock client za testove ---
type mockExchangeClient struct {
	data *client.ExchangeRateAPIResponse
}

func (m *mockExchangeClient) FetchRates(ctx context.Context) (*client.ExchangeRateAPIResponse, error) {
	return m.data, nil
}

// --- Helper funkcija za in-memory CGO-free DB (unikatna baza po testu) ---
func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "file:testdb_" + time.Now().Format("150405.000") + "?mode=memory&_pragma=foreign_keys(1)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&model.Exchange{}, &model.Listing{}, &model.ForexPair{}); err != nil {
		t.Fatal(err)
	}

	exchange := model.Exchange{
		Name:           "Simulation Exchange",
		Acronym:        "SIM",
		MicCode:        model.SimulatedExchangeMIC,
		Polity:         "International",
		Currency:       "USD",
		TimeZone:       0,
		OpenTime:       "00:00",
		CloseTime:      "23:59",
		TradingEnabled: true,
	}

	if err := db.Create(&exchange).Error; err != nil {
		t.Fatal(err)
	}

	return db
}

// --- Test za refreshFromAPI ---
func TestRefreshFromAPI(t *testing.T) {
	db := setupTestDB(t)

	mockResp := &client.ExchangeRateAPIResponse{
		BaseCode:           "RSD",
		TimeLastUpdateUnix: time.Now().Unix(),
		TimeNextUpdateUnix: time.Now().Add(time.Hour).Unix(),
		ConversionRates: map[string]float64{
			"RSD": 1,
			"EUR": 0.0080,
			"USD": 0.0085,
			"CHF": 0.0079,
			"GBP": 0.0069,
			"JPY": 1.2,
			"CAD": 0.011,
			"AUD": 0.012,
		},
	}

	mockClient := &mockExchangeClient{data: mockResp}
	repo := repository.NewForexRepository(db)
	listingRepo := repository.NewListingRepository(db)
	service := NewForexService(repo, listingRepo, mockClient)

	if err := service.refreshFromAPI(context.Background()); err != nil {
		t.Fatalf("refreshFromAPI failed: %v", err)
	}

	var pairs []model.ForexPair
	if err := db.Find(&pairs).Error; err != nil {
		t.Fatalf("query failed: %v", err)
	}

	// 8 valuta → 8*7 = 56 parova
	if len(pairs) != 56 {
		t.Fatalf("expected 56 forex pairs, got %d", len(pairs))
	}

	// opcionalna provera parova, samo nekoliko primera
	for _, pair := range pairs {
		if pair.Base == pair.Quote {
			t.Errorf("base and quote should not be same: %s/%s", pair.Base, pair.Quote)
		}
		if pair.Rate <= 0 {
			t.Errorf("rate should be positive for %s/%s, got %f", pair.Base, pair.Quote, pair.Rate)
		}
	}

	var listing model.Listing
	if err := db.Where("ticker = ?", "EUR/USD").First(&listing).Error; err != nil {
		t.Fatalf("failed loading seeded listing: %v", err)
	}

	if listing.ExchangeMIC != model.SimulatedExchangeMIC {
		t.Fatalf("expected forex listing exchange %s, got %s", model.SimulatedExchangeMIC, listing.ExchangeMIC)
	}
}

// --- Test za Initialize i seeding ---
func TestInitialize_SeedsDB(t *testing.T) {
	db := setupTestDB(t)

	mockResp := &client.ExchangeRateAPIResponse{
		BaseCode:           "RSD",
		TimeLastUpdateUnix: time.Now().Unix(),
		TimeNextUpdateUnix: time.Now().Add(time.Hour).Unix(),
		ConversionRates: map[string]float64{
			"RSD": 1,
			"EUR": 0.0080,
			"USD": 0.0085,
			"CHF": 0.0079,
			"GBP": 0.0069,
			"JPY": 1.2,
			"CAD": 0.011,
			"AUD": 0.012,
		},
	}

	mockClient := &mockExchangeClient{data: mockResp}
	repo := repository.NewForexRepository(db)
	listingRepo := repository.NewListingRepository(db)
	service := NewForexService(repo, listingRepo, mockClient)

	// DB prazna → Initialize seeduje sve parove
	service.Initialize(context.Background())

	var count int64
	if err := db.Model(&model.ForexPair{}).Count(&count).Error; err != nil {
		t.Fatalf("count query failed: %v", err)
	}

	if count != 56 {
		t.Fatalf("expected 56 forex pairs, got %d", count)
	}

	// ponovni Initialize → ne dodaje nove
	service.Initialize(context.Background())

	if err := db.Model(&model.ForexPair{}).Count(&count).Error; err != nil {
		t.Fatalf("count query failed: %v", err)
	}

	if count != 56 {
		t.Fatalf("expected count still 56, got %d", count)
	}
}
