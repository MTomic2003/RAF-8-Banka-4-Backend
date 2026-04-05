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

type mockStockClient struct {
	symbols    []client.Symbol
	profiles   map[string]*client.Profile
	quotes     map[string]*client.Quote
	financials map[string]*client.BasicFinancials
}

func (m *mockStockClient) GetSymbols(exchange string) ([]client.Symbol, error) {
	return m.symbols, nil
}

func (m *mockStockClient) GetProfile(ticker string) (*client.Profile, error) {
	return m.profiles[ticker], nil
}

func (m *mockStockClient) GetQuote(ticker string) (*client.Quote, error) {
	return m.quotes[ticker], nil
}

func (m *mockStockClient) GetBasicFinancials(ticker string) (*client.BasicFinancials, error) {
	return m.financials[ticker], nil
}

func setupStockServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:testdb_stock_" + time.Now().Format("150405.000") + "?mode=memory&_pragma=foreign_keys(1)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&model.Exchange{}, &model.Listing{}, &model.Stock{}, &model.Option{}); err != nil {
		t.Fatal(err)
	}

	exchanges := []model.Exchange{
		{Name: "Nasdaq", Acronym: "NASDAQ", MicCode: "XNAS", Polity: "USA", Currency: "USD", TimeZone: -4, OpenTime: "09:30", CloseTime: "16:00", TradingEnabled: true},
		{Name: "Simulation Exchange", Acronym: "SIM", MicCode: model.SimulatedExchangeMIC, Polity: "International", Currency: "USD", TimeZone: 0, OpenTime: "00:00", CloseTime: "23:59", TradingEnabled: true},
	}

	for i := range exchanges {
		if err := db.Create(&exchanges[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	return db
}

func TestSeedStocks_UsesSymbolMICAndGeneratedOptionsUseSimulationExchange(t *testing.T) {
	db := setupStockServiceTestDB(t)

	financials := &client.BasicFinancials{}
	financials.Metric.DividendYieldIndicatedAnnual = 0.42

	mockClient := &mockStockClient{
		symbols: []client.Symbol{
			{Symbol: "AAPL", MIC: "XNAS"},
		},
		profiles: map[string]*client.Profile{
			"AAPL": {Name: "Apple Inc", Exchange: "NASDAQ"},
		},
		quotes: map[string]*client.Quote{
			"AAPL": {CurrentPrice: 150.0, High: 151.0},
		},
		financials: map[string]*client.BasicFinancials{
			"AAPL": financials,
		},
	}

	svc := newStockService(
		repository.NewListingRepository(db),
		repository.NewStockRepository(db),
		repository.NewOptionRepository(db),
		repository.NewExchangeRepository(db),
		mockClient,
	)

	if err := svc.SeedStocks(context.Background(), 1); err != nil {
		t.Fatalf("SeedStocks failed: %v", err)
	}

	var stockListing model.Listing
	if err := db.Where("ticker = ?", "AAPL").First(&stockListing).Error; err != nil {
		t.Fatalf("failed to load seeded stock listing: %v", err)
	}

	if stockListing.ExchangeMIC != "XNAS" {
		t.Fatalf("expected stock exchange XNAS, got %s", stockListing.ExchangeMIC)
	}

	if err := svc.SeedOptions(context.Background(), 1); err != nil {
		t.Fatalf("SeedOptions failed: %v", err)
	}

	var optionListing model.Listing
	if err := db.Where("listing_type = ?", model.ListingTypeOption).First(&optionListing).Error; err != nil {
		t.Fatalf("failed to load seeded option listing: %v", err)
	}

	if optionListing.ExchangeMIC != model.SimulatedExchangeMIC {
		t.Fatalf("expected option exchange %s, got %s", model.SimulatedExchangeMIC, optionListing.ExchangeMIC)
	}
}
