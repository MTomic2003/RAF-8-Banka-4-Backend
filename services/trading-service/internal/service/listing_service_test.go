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

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(
		&model.Exchange{},
		&model.Asset{},
		&model.Listing{},
		&model.Stock{},
		&model.FuturesContract{},
		&model.ForexPair{},
		&model.ListingDailyPriceInfo{},
		&model.Option{},
	); err != nil {
		t.Fatal(err)
	}

	return db
}

func seedListingTestExchanges(t *testing.T, db *gorm.DB) {
	exchanges := []model.Exchange{
		{Name: "Nasdaq", Acronym: "NASDAQ", MicCode: "XNAS", Polity: "USA", Currency: "USD", TimeZone: -4, OpenTime: "09:30", CloseTime: "16:00", TradingEnabled: true},
		{Name: "Chicago Mercantile Exchange", Acronym: "CME", MicCode: "XCME", Polity: "USA", Currency: "USD", TimeZone: -4, OpenTime: "09:30", CloseTime: "16:00", TradingEnabled: true},
		{Name: "Simulation Exchange", Acronym: "SIM", MicCode: model.SimulatedExchangeMIC, Polity: "International", Currency: "USD", TimeZone: 0, OpenTime: "00:00", CloseTime: "23:59", TradingEnabled: true},
	}

	for i := range exchanges {
		if err := db.Create(&exchanges[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func seedListingTestData(t *testing.T, db *gorm.DB) {
	seedListingTestExchanges(t, db)

	// Create assets
	aaplAsset := model.Asset{Ticker: "AAPL", Name: "Apple Inc", AssetType: model.AssetTypeStock}
	googAsset := model.Asset{Ticker: "GOOG", Name: "Alphabet Inc", AssetType: model.AssetTypeStock}
	db.Create(&aaplAsset)
	db.Create(&googAsset)

	// Create listings referencing assets
	listings := []model.Listing{
		{AssetID: aaplAsset.AssetID, ExchangeMIC: "XNAS", Price: 150.0, Ask: 151.0, MaintenanceMargin: 10.0, LastRefresh: time.Now()},
		{AssetID: googAsset.AssetID, ExchangeMIC: "XNAS", Price: 2800.0, Ask: 2801.0, MaintenanceMargin: 20.0, LastRefresh: time.Now()},
	}
	for i := range listings {
		if err := db.Create(&listings[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	stocks := []model.Stock{
		{AssetID: aaplAsset.AssetID, OutstandingShares: 1000000, DividendYield: 0.5},
		{AssetID: googAsset.AssetID, OutstandingShares: 500000, DividendYield: 0.0},
	}
	for _, s := range stocks {
		if err := db.Omit("Asset").Create(&s).Error; err != nil {
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

	// Futures
	futuresAsset := model.Asset{Ticker: "CLJ26", Name: "Crude Oil", AssetType: model.AssetTypeFuture}
	db.Create(&futuresAsset)

	futuresListing := model.Listing{
		AssetID: futuresAsset.AssetID, ExchangeMIC: "XCME",
		Price: 75.0, Ask: 75.5, MaintenanceMargin: 5.0, LastRefresh: time.Now(),
	}
	db.Create(&futuresListing)

	futuresContract := model.FuturesContract{
		AssetID:        futuresAsset.AssetID,
		ContractSize:   1000,
		ContractUnit:   "barrels",
		SettlementDate: time.Now().AddDate(0, 3, 0),
	}
	db.Create(&futuresContract)

	// Forex
	eurUsdAsset := model.Asset{Ticker: "EUR/USD", Name: "EUR/USD", AssetType: model.AssetTypeForexPair}
	usdRsdAsset := model.Asset{Ticker: "USD/RSD", Name: "USD/RSD", AssetType: model.AssetTypeForexPair}
	db.Create(&eurUsdAsset)
	db.Create(&usdRsdAsset)

	forexListings := []model.Listing{
		{AssetID: eurUsdAsset.AssetID, ExchangeMIC: model.SimulatedExchangeMIC, Price: 1.08, LastRefresh: time.Now()},
		{AssetID: usdRsdAsset.AssetID, ExchangeMIC: model.SimulatedExchangeMIC, Price: 117.0, LastRefresh: time.Now()},
	}
	for i := range forexListings {
		db.Create(&forexListings[i])
	}

	forexPairs := []model.ForexPair{
		{AssetID: eurUsdAsset.AssetID, Base: "EUR", Quote: "USD", Rate: 1.08},
		{AssetID: usdRsdAsset.AssetID, Base: "USD", Quote: "RSD", Rate: 117.0},
	}
	for _, p := range forexPairs {
		db.Omit("Asset").Create(&p)
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

func TestGetStockDetails_Success(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	// Create an option for the AAPL stock
	optionAsset := model.Asset{
		Ticker: "AAPL220404C00180000", Name: "AAPL Call", AssetType: model.AssetTypeOption,
	}
	db.Create(&optionAsset)

	optionListing := model.Listing{
		AssetID: optionAsset.AssetID, ExchangeMIC: "XNAS",
		Price: 5.50, Ask: 5.60, LastRefresh: time.Now(),
	}
	db.Create(&optionListing)

	option := model.Option{
		AssetID: optionAsset.AssetID, StockID: 1, OptionType: model.OptionTypeCall,
		StrikePrice: 180.0, ContractSize: 100, SettlementDate: time.Now().AddDate(0, 1, 0),
	}
	db.Create(&option)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	result, err := svc.GetStockDetails(context.Background(), 1, 0)
	if err != nil {
		t.Fatalf("GetStockDetails failed: %v", err)
	}

	if result.Ticker != "AAPL" {
		t.Errorf("expected ticker AAPL, got %s", result.Ticker)
	}
	if len(result.History) != 1 {
		t.Errorf("expected 1 history record, got %d", len(result.History))
	}
	if len(result.Options) != 1 {
		t.Fatalf("expected 1 option, got %d", len(result.Options))
	}
	if result.Options[0].OptionType != "CALL" {
		t.Errorf("expected option type CALL, got %s", result.Options[0].OptionType)
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

// --- Detailed Views Tests ---

func TestGetFutureDetails_Success(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	var futureListing model.Listing
	db.Joins("INNER JOIN assets ON assets.asset_id = listings.asset_id").
		Where("assets.ticker = ?", "CLJ26").First(&futureListing)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	result, err := svc.GetFutureDetails(context.Background(), futureListing.ListingID, 0)
	if err != nil {
		t.Fatalf("GetFutureDetails failed: %v", err)
	}

	if result.Ticker != "CLJ26" {
		t.Errorf("expected ticker CLJ26, got %s", result.Ticker)
	}
	if result.ContractSize != 1000 {
		t.Errorf("expected contract size 1000, got %f", result.ContractSize)
	}
	if result.ContractUnit != "barrels" {
		t.Errorf("expected contract unit barrels, got %s", result.ContractUnit)
	}
}

func TestGetForexDetails_Success(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	var forexListing model.Listing
	db.Joins("INNER JOIN assets ON assets.asset_id = listings.asset_id").
		Where("assets.ticker = ?", "EUR/USD").First(&forexListing)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	result, err := svc.GetForexDetails(context.Background(), forexListing.ListingID, 0)
	if err != nil {
		t.Fatalf("GetForexDetails failed: %v", err)
	}

	if result.Ticker != "EUR/USD" {
		t.Errorf("expected ticker EUR/USD, got %s", result.Ticker)
	}
	if result.Price != 1.08 {
		t.Errorf("expected price 1.08, got %f", result.Price)
	}
}

// --- GetOptions Tests ---

func TestGetOptions_ReturnsAll(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	// Insert two option assets, listings, and options
	optAsset1 := model.Asset{Ticker: "AAPL:CALL:150.00", Name: "AAPL CALL 150", AssetType: model.AssetTypeOption}
	optAsset2 := model.Asset{Ticker: "AAPL:PUT:140.00", Name: "AAPL PUT 140", AssetType: model.AssetTypeOption}
	db.Create(&optAsset1)
	db.Create(&optAsset2)

	optListing1 := model.Listing{AssetID: optAsset1.AssetID, ExchangeMIC: model.SimulatedExchangeMIC, Price: 5.0, Ask: 5.1, LastRefresh: time.Now()}
	optListing2 := model.Listing{AssetID: optAsset2.AssetID, ExchangeMIC: model.SimulatedExchangeMIC, Price: 3.0, Ask: 3.1, LastRefresh: time.Now()}
	db.Create(&optListing1)
	db.Create(&optListing2)

	opt1 := model.Option{AssetID: optAsset1.AssetID, StockID: 1, OptionType: model.OptionTypeCall, StrikePrice: 150.0, ContractSize: 100, SettlementDate: time.Now().AddDate(0, 1, 0)}
	opt2 := model.Option{AssetID: optAsset2.AssetID, StockID: 1, OptionType: model.OptionTypePut, StrikePrice: 140.0, ContractSize: 100, SettlementDate: time.Now().AddDate(0, 1, 0)}
	db.Create(&opt1)
	db.Create(&opt2)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	result, err := svc.GetOptions(context.Background(), dto.ListingQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("GetOptions failed: %v", err)
	}
	if len(result.Data) != 2 {
		t.Fatalf("expected 2 options, got %d", len(result.Data))
	}
}

func TestGetOptions_InvalidSettlementDate_BadRequest(t *testing.T) {
	db := setupListingTestDB(t)
	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	_, err := svc.GetOptions(context.Background(), dto.ListingQuery{SettlementDate: "not-a-date"})
	if err == nil {
		t.Fatal("expected error for invalid settlement date")
	}
}

// --- GetFutureDetails error path tests ---

func TestGetFutureDetails_NotFound_WrongType(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	// Use ID=1 which is a stock, not a future
	_, err := svc.GetFutureDetails(context.Background(), 1, 0)
	if err == nil {
		t.Fatal("expected error when listing type is not future")
	}
}

func TestGetFutureDetails_ListingNotFound(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	_, err := svc.GetFutureDetails(context.Background(), 99999, 0)
	if err == nil {
		t.Fatal("expected error for nonexistent listing")
	}
}

// --- GetForexDetails error path tests ---

func TestGetForexDetails_NotFound_WrongType(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	// Use ID=1 which is a stock, not a forex pair
	_, err := svc.GetForexDetails(context.Background(), 1, 0)
	if err == nil {
		t.Fatal("expected error when listing type is not forexPair")
	}
}

func TestGetForexDetails_ListingNotFound(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	_, err := svc.GetForexDetails(context.Background(), 99999, 0)
	if err == nil {
		t.Fatal("expected error for nonexistent listing")
	}
}

// --- GetOptionDetails error path tests ---

func TestGetOptionDetails_NotFound_WrongType(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	// Use ID=1 which is a stock, not an option
	_, err := svc.GetOptionDetails(context.Background(), 1, 0)
	if err == nil {
		t.Fatal("expected error when listing type is not option")
	}
}

func TestGetOptionDetails_ListingNotFound(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	_, err := svc.GetOptionDetails(context.Background(), 99999, 0)
	if err == nil {
		t.Fatal("expected error for nonexistent listing")
	}
}

// --- latestDaily Tests ---

func TestLatestDaily_EmptySlice(t *testing.T) {
	result := latestDaily([]model.ListingDailyPriceInfo{})
	if result != nil {
		t.Fatal("expected nil for empty slice")
	}
}

func TestLatestDaily_SingleElement(t *testing.T) {
	now := time.Now()
	infos := []model.ListingDailyPriceInfo{
		{Date: now, Price: 100.0},
	}
	result := latestDaily(infos)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Price != 100.0 {
		t.Errorf("expected price 100.0, got %f", result.Price)
	}
}

func TestLatestDaily_MultipleElements_ReturnsLatest(t *testing.T) {
	now := time.Now()
	infos := []model.ListingDailyPriceInfo{
		{Date: now.Add(-48 * time.Hour), Price: 90.0},
		{Date: now, Price: 100.0},
		{Date: now.Add(-24 * time.Hour), Price: 95.0},
	}
	result := latestDaily(infos)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Price != 100.0 {
		t.Errorf("expected latest price 100.0, got %f", result.Price)
	}
}

// --- mapHistory Tests ---

func TestMapHistory_EmptySlice(t *testing.T) {
	result := mapHistory([]model.ListingDailyPriceInfo{})
	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %d", len(result))
	}
}

func TestMapHistory_MapsAllFields(t *testing.T) {
	now := time.Now()
	infos := []model.ListingDailyPriceInfo{
		{Date: now, Price: 100.0, Ask: 101.0, Bid: 99.0, Change: 1.5, Volume: 500},
	}
	result := mapHistory(infos)
	if len(result) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result))
	}
	r := result[0]
	if r.Price != 100.0 || r.Ask != 101.0 || r.Bid != 99.0 || r.Change != 1.5 || r.Volume != 500 {
		t.Errorf("mapHistory fields not mapped correctly: %+v", r)
	}
}

func TestGetOptionDetails_Success(t *testing.T) {
	db := setupListingTestDB(t)
	seedListingTestData(t, db)

	optionAsset := model.Asset{
		Ticker: "AAPL220404P00180000", Name: "AAPL Put", AssetType: model.AssetTypeOption,
	}
	db.Create(&optionAsset)

	optionListing := model.Listing{
		AssetID: optionAsset.AssetID, ExchangeMIC: "XNAS",
		Price: 3.20, Ask: 3.30, LastRefresh: time.Now(),
	}
	db.Create(&optionListing)

	option := model.Option{
		AssetID: optionAsset.AssetID, StockID: 1, OptionType: model.OptionTypePut,
		StrikePrice: 180.0, ContractSize: 100, SettlementDate: time.Now().AddDate(0, 1, 0),
	}
	db.Create(&option)

	svc := NewListingService(
		repository.NewListingRepository(db),
		repository.NewFuturesContractRepository(db),
		repository.NewForexRepository(db),
		repository.NewOptionRepository(db),
	)

	result, err := svc.GetOptionDetails(context.Background(), optionListing.ListingID, 0)
	if err != nil {
		t.Fatalf("GetOptionDetails failed: %v", err)
	}

	if result.Ticker != "AAPL220404P00180000" {
		t.Errorf("expected ticker AAPL220404P00180000, got %s", result.Ticker)
	}
	if result.OptionType != string(model.OptionTypePut) {
		t.Errorf("expected option type PUT, got %s", result.OptionType)
	}
	if result.Strike != 180.0 {
		t.Errorf("expected strike 180.0, got %f", result.Strike)
	}
}
