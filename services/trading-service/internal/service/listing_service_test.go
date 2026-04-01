package service

import (
	"context"
	"testing"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupListingTestDB(t *testing.T) *gorm.DB {
	dsn := "file:testdb_listing_" + time.Now().Format("150405.000") + "?mode=memory&_pragma=foreign_keys(1)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(
		&model.Listing{},
		&model.Stock{},
		&model.FuturesContract{},
		&model.ForexPair{},
		&model.ListingDailyPriceInfo{},
	); err != nil {
		t.Fatal(err)
	}

	return db
}

func seedListingTestData(t *testing.T, db *gorm.DB) {
	listings := []model.Listing{
		{Ticker: "AAPL", Name: "Apple Inc", ExchangeMIC: "XNAS", Price: 150.0, Ask: 151.0, MaintenanceMargin: 10.0, LastRefresh: time.Now(), ListingType: model.ListingTypeStock},
		{Ticker: "GOOG", Name: "Alphabet Inc", ExchangeMIC: "XNAS", Price: 2800.0, Ask: 2801.0, MaintenanceMargin: 20.0, LastRefresh: time.Now(), ListingType: model.ListingTypeStock},
	}
	for i := range listings {
		if err := db.Create(&listings[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	stocks := []model.Stock{
		{ListingID: listings[0].ListingID, OutstandingShares: 1000000, DividendYield: 0.5},
		{ListingID: listings[1].ListingID, OutstandingShares: 500000, DividendYield: 0.0},
	}
	for _, s := range stocks {
		if err := db.Omit("Listing").Create(&s).Error; err != nil {
			t.Fatal(err)
		}
	}

	dailyInfos := []model.ListingDailyPriceInfo{
		{ListingID: listings[0].ListingID, Date: time.Now(), Bid: 149.0, Change: 1.5, Volume: 1000},
		{ListingID: listings[1].ListingID, Date: time.Now(), Bid: 2799.0, Change: -5.0, Volume: 500},
	}
	for _, d := range dailyInfos {
		if err := db.Omit("Listing").Create(&d).Error; err != nil {
			t.Fatal(err)
		}
	}

	futuresListing := model.Listing{
		Ticker: "CLJ26", Name: "Crude Oil", ExchangeMIC: "XCME",
		Price: 75.0, Ask: 75.5, MaintenanceMargin: 5.0,
		LastRefresh: time.Now(), ListingType: model.ListingTypeFuture,
	}
	if err := db.Create(&futuresListing).Error; err != nil {
		t.Fatal(err)
	}
	futuresContract := model.FuturesContract{
		ListingID:      futuresListing.ListingID,
		ContractSize:   1000,
		ContractUnit:   "barrels",
		SettlementDate: time.Now().AddDate(0, 3, 0),
	}
	if err := db.Create(&futuresContract).Error; err != nil {
		t.Fatal(err)
	}

	forexListings := []model.Listing{
		{Ticker: "EUR/USD", Name: "EUR/USD", ExchangeMIC: "FOREX", Price: 1.08, LastRefresh: time.Now(), ListingType: model.ListingTypeForexPair},
		{Ticker: "USD/RSD", Name: "USD/RSD", ExchangeMIC: "FOREX", Price: 117.0, LastRefresh: time.Now(), ListingType: model.ListingTypeForexPair},
	}
	for i := range forexListings {
		if err := db.Create(&forexListings[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	forexPairs := []model.ForexPair{
		{ListingID: forexListings[0].ListingID, Base: "EUR", Quote: "USD", Rate: 1.08},
		{ListingID: forexListings[1].ListingID, Base: "USD", Quote: "RSD", Rate: 117.0},
	}
	for _, p := range forexPairs {
		if err := db.Omit("Listing").Create(&p).Error; err != nil {
			t.Fatal(err)
		}
	}
}

// --- Stocks ---

func TestGetStocks_ReturnsAll(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	result, err := svc.GetStocks(context.Background(), dto.ListingQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetStocks failed: %v", err)
	}

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 stocks, got %d", len(result.Data))
	}
	if result.Total != 2 {
		t.Fatalf("expected total 2, got %d", result.Total)
	}
}

func TestGetStocks_FilterByExchange(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	result, err := svc.GetStocks(context.Background(), dto.ListingQuery{
		Exchange: "XNAS",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("GetStocks failed: %v", err)
	}

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 stocks for XNAS, got %d", len(result.Data))
	}
}

func TestGetStocks_FilterBySearch(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	result, err := svc.GetStocks(context.Background(), dto.ListingQuery{
		Search:   "AAPL",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("GetStocks failed: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 stock for AAPL, got %d", len(result.Data))
	}
	if result.Data[0].Ticker != "AAPL" {
		t.Errorf("expected ticker AAPL, got %s", result.Data[0].Ticker)
	}
}

func TestGetStocks_InitialMarginCost(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	result, err := svc.GetStocks(context.Background(), dto.ListingQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetStocks failed: %v", err)
	}

	for _, s := range result.Data {
		expected := s.MaintenanceMargin * 1.1
		if s.InitialMarginCost != expected {
			t.Errorf("expected InitialMarginCost %f, got %f", expected, s.InitialMarginCost)
		}
	}
}

// --- Futures ---

func TestGetFutures_ReturnsAll(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	result, err := svc.GetFutures(context.Background(), dto.ListingQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetFutures failed: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 future, got %d", len(result.Data))
	}
	if result.Data[0].Ticker != "CLJ26" {
		t.Errorf("expected ticker CLJ26, got %s", result.Data[0].Ticker)
	}
}

func TestGetFutures_ContractDataPresent(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	result, err := svc.GetFutures(context.Background(), dto.ListingQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetFutures failed: %v", err)
	}

	f := result.Data[0]
	if f.ContractSize != 1000 {
		t.Errorf("expected ContractSize 1000, got %f", f.ContractSize)
	}
	if f.ContractUnit != "barrels" {
		t.Errorf("expected ContractUnit barrels, got %s", f.ContractUnit)
	}
	if f.SettlementDate.IsZero() {
		t.Error("expected non-zero SettlementDate")
	}
}

// --- Forex ---

func TestGetForex_ReturnsAll(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	result, err := svc.GetForex(context.Background(), dto.ListingQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetForex failed: %v", err)
	}

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 forex pairs, got %d", len(result.Data))
	}
}

func TestGetForex_TickerFormat(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	result, err := svc.GetForex(context.Background(), dto.ListingQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetForex failed: %v", err)
	}

	for _, p := range result.Data {
		if p.Ticker != p.Base+"/"+p.Quote {
			t.Errorf("expected ticker %s/%s, got %s", p.Base, p.Quote, p.Ticker)
		}
	}
}
