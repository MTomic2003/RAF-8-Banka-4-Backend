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

// ── stringsContainsColon Tests ───────────────────────────────────

func TestStringsContainsColon_WithColon(t *testing.T) {
	if !stringsContainsColon("abc:def") {
		t.Fatal("expected true for string with colon")
	}
}

func TestStringsContainsColon_WithoutColon(t *testing.T) {
	if stringsContainsColon("abcdef") {
		t.Fatal("expected false for string without colon")
	}
}

func TestStringsContainsColon_Empty(t *testing.T) {
	if stringsContainsColon("") {
		t.Fatal("expected false for empty string")
	}
}

func TestStringsContainsColon_ColonOnly(t *testing.T) {
	if !stringsContainsColon(":") {
		t.Fatal("expected true for colon-only string")
	}
}

// ── stringsContainsDot Tests ─────────────────────────────────────

func TestStringsContainsDot_WithDot(t *testing.T) {
	if !stringsContainsDot("abc.def") {
		t.Fatal("expected true for string with dot")
	}
}

func TestStringsContainsDot_WithoutDot(t *testing.T) {
	if stringsContainsDot("abcdef") {
		t.Fatal("expected false for string without dot")
	}
}

// ── roundToInt Tests ─────────────────────────────────────────────

func TestRoundToInt(t *testing.T) {
	tests := []struct {
		input    float64
		expected int
	}{
		{1.0, 1},
		{1.4, 1},
		{1.5, 2},
		{1.9, 2},
		{0.0, 0},
		{99.5, 100},
		{99.49, 99},
	}
	for _, tc := range tests {
		result := roundToInt(tc.input)
		if result != tc.expected {
			t.Fatalf("roundToInt(%v) = %d, want %d", tc.input, result, tc.expected)
		}
	}
}

// ── generateExpirationDates Tests ────────────────────────────────

func TestGenerateExpirationDates_AllFuture(t *testing.T) {
	dates := generateExpirationDates()
	if len(dates) == 0 {
		t.Fatal("expected non-empty dates")
	}

	now := time.Now()
	for i, d := range dates {
		if !d.After(now) {
			t.Fatalf("date[%d] = %v is not after now %v", i, d, now)
		}
	}

	// Verify dates are in ascending order
	for i := 1; i < len(dates); i++ {
		if !dates[i].After(dates[i-1]) {
			t.Fatalf("dates not ascending: dates[%d]=%v <= dates[%d]=%v", i, dates[i], i-1, dates[i-1])
		}
	}
}

func TestGenerateExpirationDates_Count(t *testing.T) {
	dates := generateExpirationDates()
	// First batch: i=6,12,18,24,30 -> 5 dates (every 6 days from 6 to 30 inclusive -> (30-6)/6+1=5)
	// Actually: for i := 6; i <= 30; i += 6 -> i=6,12,18,24,30 -> 5 dates
	// Second batch: 6 more dates at 30-day intervals
	// Total: 5 + 6 = 11
	if len(dates) != 11 {
		t.Fatalf("expected 11 dates, got %d", len(dates))
	}
}

// ── waitForNextCall Tests ────────────────────────────────────────

func TestWaitForNextCall_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := waitForNextCall(ctx, 10*time.Second)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

// ── Stop idempotency ────────────────────────────────────────────

func TestStockService_StopWithoutStart(t *testing.T) {
	svc := newStockService(nil, nil, nil, nil, &mockStockClient{})
	// Should not panic
	svc.Stop()
}

func TestStockService_StartStop(t *testing.T) {
	svc := newStockService(nil, nil, nil, nil, &mockStockClient{})
	svc.Start()
	svc.Start() // double start should be no-op
	svc.Stop()
	svc.Stop() // double stop should be safe
}

// ── NewStockService Tests ────────────────────────────────────────

func TestNewStockService_ReturnsNonNil(t *testing.T) {
	db := setupStockServiceTestDB(t)
	svc := NewStockService(
		repository.NewListingRepository(db),
		repository.NewStockRepository(db),
		repository.NewOptionRepository(db),
		repository.NewExchangeRepository(db),
		nil, // client not needed for this test
	)
	if svc == nil {
		t.Fatal("expected non-nil StockService")
	}
}

// ── Initialize Tests ─────────────────────────────────────────────

func TestInitialize_StocksAlreadyExist_Skips(t *testing.T) {
	db := setupStockServiceTestDB(t)

	// Seed a stock so count > 0
	listing := model.Listing{
		Ticker:      "MSFT",
		Name:        "Microsoft",
		ExchangeMIC: "XNAS",
		Price:       300.0,
		Ask:         301.0,
		LastRefresh: time.Now(),
		ListingType: model.ListingTypeStock,
	}
	db.Create(&listing)
	stock := model.Stock{ListingID: listing.ListingID, OutstandingShares: 1000, DividendYield: 0.1}
	db.Omit("Listing").Create(&stock)

	mockClient := &mockStockClient{}
	svc := newStockService(
		repository.NewListingRepository(db),
		repository.NewStockRepository(db),
		repository.NewOptionRepository(db),
		repository.NewExchangeRepository(db),
		mockClient,
	)

	// Should skip seeding without error
	svc.Initialize(context.Background())
}

func TestInitialize_EmptyDB_SeedsStocks(t *testing.T) {
	db := setupStockServiceTestDB(t)

	financials := &client.BasicFinancials{}
	financials.Metric.DividendYieldIndicatedAnnual = 0.01

	mockClient := &mockStockClient{
		symbols: []client.Symbol{
			{Symbol: "TSLA", MIC: "XNAS"},
		},
		profiles: map[string]*client.Profile{
			"TSLA": {Name: "Tesla Inc", Exchange: "NASDAQ"},
		},
		quotes: map[string]*client.Quote{
			"TSLA": {CurrentPrice: 200.0, High: 201.0},
		},
		financials: map[string]*client.BasicFinancials{
			"TSLA": financials,
		},
	}

	svc := newStockService(
		repository.NewListingRepository(db),
		repository.NewStockRepository(db),
		repository.NewOptionRepository(db),
		repository.NewExchangeRepository(db),
		mockClient,
	)

	// Empty DB -> should seed
	svc.Initialize(context.Background())

	var count int64
	db.Model(&model.Stock{}).Count(&count)
	if count == 0 {
		t.Fatal("expected at least one stock seeded")
	}
}

// ── RefreshPrices Tests ───────────────────────────────────────────

func TestRefreshPrices_UpdatesListings(t *testing.T) {
	db := setupStockServiceTestDB(t)

	// seed a listing
	listing := model.Listing{
		Ticker:      "AAPL",
		Name:        "Apple Inc",
		ExchangeMIC: "XNAS",
		Price:       150.0,
		Ask:         151.0,
		LastRefresh: time.Now(),
		ListingType: model.ListingTypeStock,
	}
	db.Create(&listing)

	mockClient := &mockStockClient{
		quotes: map[string]*client.Quote{
			"AAPL": {CurrentPrice: 155.0, High: 156.0},
		},
	}

	svc := newStockService(
		repository.NewListingRepository(db),
		repository.NewStockRepository(db),
		repository.NewOptionRepository(db),
		repository.NewExchangeRepository(db),
		mockClient,
	)

	err := svc.RefreshPrices(context.Background())
	if err != nil {
		t.Fatalf("RefreshPrices failed: %v", err)
	}

	var updated model.Listing
	db.Where("ticker = ?", "AAPL").First(&updated)
	if updated.Price != 155.0 {
		t.Errorf("expected price 155.0, got %f", updated.Price)
	}
}

func TestRefreshPrices_SkipsZeroPrice(t *testing.T) {
	db := setupStockServiceTestDB(t)

	listing := model.Listing{
		Ticker:      "ZERO",
		Name:        "Zero Price Stock",
		ExchangeMIC: "XNAS",
		Price:       100.0,
		Ask:         101.0,
		LastRefresh: time.Now(),
		ListingType: model.ListingTypeStock,
	}
	db.Create(&listing)

	mockClient := &mockStockClient{
		quotes: map[string]*client.Quote{
			"ZERO": {CurrentPrice: 0.0, High: 0.0}, // zero price should be skipped
		},
	}

	svc := newStockService(
		repository.NewListingRepository(db),
		repository.NewStockRepository(db),
		repository.NewOptionRepository(db),
		repository.NewExchangeRepository(db),
		mockClient,
	)

	err := svc.RefreshPrices(context.Background())
	if err != nil {
		t.Fatalf("RefreshPrices failed: %v", err)
	}

	// price should remain unchanged
	var check model.Listing
	db.Where("ticker = ?", "ZERO").First(&check)
	if check.Price != 100.0 {
		t.Errorf("expected price to stay 100.0, got %f", check.Price)
	}
}

func TestRefreshPrices_EmptyListings_NoError(t *testing.T) {
	db := setupStockServiceTestDB(t)

	mockClient := &mockStockClient{}
	svc := newStockService(
		repository.NewListingRepository(db),
		repository.NewStockRepository(db),
		repository.NewOptionRepository(db),
		repository.NewExchangeRepository(db),
		mockClient,
	)

	err := svc.RefreshPrices(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// ── RefreshOptions Tests ──────────────────────────────────────────

func TestRefreshOptions_CreatesOptions(t *testing.T) {
	db := setupStockServiceTestDB(t)

	// seed a stock (non-option ticker, no colon)
	listing := model.Listing{
		Ticker:      "GOOG",
		Name:        "Alphabet Inc",
		ExchangeMIC: "XNAS",
		Price:       2800.0,
		Ask:         2801.0,
		LastRefresh: time.Now(),
		ListingType: model.ListingTypeStock,
	}
	db.Create(&listing)
	stock := model.Stock{ListingID: listing.ListingID, OutstandingShares: 500000, DividendYield: 0.0}
	db.Omit("Listing").Create(&stock)

	mockClient := &mockStockClient{}
	svc := newStockService(
		repository.NewListingRepository(db),
		repository.NewStockRepository(db),
		repository.NewOptionRepository(db),
		repository.NewExchangeRepository(db),
		mockClient,
	)

	err := svc.RefreshOptions(context.Background())
	if err != nil {
		t.Fatalf("RefreshOptions failed: %v", err)
	}

	var count int64
	db.Model(&model.Option{}).Count(&count)
	if count == 0 {
		t.Fatal("expected options to be created")
	}
}

func TestRefreshOptions_SkipsOptionTickers(t *testing.T) {
	db := setupStockServiceTestDB(t)

	// seed an option-like stock ticker (has colon)
	listing := model.Listing{
		Ticker:      "AAPL:CALL:150.00",
		Name:        "AAPL option",
		ExchangeMIC: model.SimulatedExchangeMIC,
		Price:       5.0,
		Ask:         5.1,
		LastRefresh: time.Now(),
		ListingType: model.ListingTypeOption,
	}
	db.Create(&listing)
	stock := model.Stock{ListingID: listing.ListingID, OutstandingShares: 0, DividendYield: 0.0}
	db.Omit("Listing").Create(&stock)

	mockClient := &mockStockClient{}
	svc := newStockService(
		repository.NewListingRepository(db),
		repository.NewStockRepository(db),
		repository.NewOptionRepository(db),
		repository.NewExchangeRepository(db),
		mockClient,
	)

	err := svc.RefreshOptions(context.Background())
	if err != nil {
		t.Fatalf("RefreshOptions failed: %v", err)
	}
	// no options should have been created since the ticker has a colon
	var count int64
	db.Model(&model.Option{}).Count(&count)
	if count != 0 {
		t.Errorf("expected no options for colon ticker, got %d", count)
	}
}

func TestRefreshOptions_EmptyStocks_NoError(t *testing.T) {
	db := setupStockServiceTestDB(t)

	mockClient := &mockStockClient{}
	svc := newStockService(
		repository.NewListingRepository(db),
		repository.NewStockRepository(db),
		repository.NewOptionRepository(db),
		repository.NewExchangeRepository(db),
		mockClient,
	)

	err := svc.RefreshOptions(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
