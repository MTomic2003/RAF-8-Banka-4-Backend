package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

const (
	maxCallsPerMinute    = 55
	priceRefreshInterval = 15 * time.Minute
)

type StockService struct {
	listingRepo repository.ListingRepository
	stockRepo   repository.StockRepository
	optionRepo  repository.OptionRepository
	client      *client.StockClient

	mu     sync.Mutex
	cancel context.CancelFunc
}

func NewStockService(
	listingRepo repository.ListingRepository,
	stockRepo repository.StockRepository,
	optionRepo repository.OptionRepository,
	client *client.StockClient,
) *StockService {
	return &StockService{
		listingRepo: listingRepo,
		stockRepo:   stockRepo,
		optionRepo:  optionRepo,
		client:      client,
	}
}

func (s *StockService) Initialize(ctx context.Context) {
	count, err := s.stockRepo.Count(ctx)
	if err != nil {
		log.Printf("[seed] failed to count listings: %v", err)
		return
	}

	if count > 0 {
		log.Printf("[seed] listings already exist (%d), skipping seed", count)
		return
	}

	if err := s.SeedStocks(ctx, 10); err != nil {
		log.Printf("[seed] stock seed failed: %v", err)
		return
	}

	if err := s.SeedOptions(ctx, 5); err != nil {
		log.Printf("[seed] option seed failed: %v", err)
	}
}

func (s *StockService) Start() {
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()

	ticker := time.NewTicker(priceRefreshInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.RefreshPrices(ctx); err != nil {
					log.Println("[refresh] failed:", err)
				}
				if err := s.RefreshOptions(ctx); err != nil {
					log.Println("[refresh-options] failed:", err)
				}
			}
		}
	}()
}

func (s *StockService) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (s *StockService) SeedStocks(ctx context.Context, limit int) error {
	symbols, err := s.client.GetSymbols("US")
	if err != nil {
		return fmt.Errorf("failed to fetch symbols: %w", err)
	}

	log.Printf("[seed] fetched %d symbols, seeding up to %d", len(symbols), limit)

	callsThisMinute := 1
	minuteStart := time.Now()
	count := 0

	for _, sym := range symbols {
		if count >= limit {
			break
		}

		if stringsContainsDot(sym.Symbol) {
			continue
		}

		if callsThisMinute+3 > maxCallsPerMinute {
			elapsed := time.Since(minuteStart)
			if elapsed < time.Minute {
				wait := time.Minute - elapsed
				log.Printf("[seed] rate limit reached, waiting %s...", wait.Round(time.Second))
				if err := waitForNextCall(ctx, wait); err != nil {
					return err
				}
			}
			callsThisMinute = 0
			minuteStart = time.Now()
		}

		profile, err := s.client.GetProfile(sym.Symbol)
		callsThisMinute++
		if err != nil || profile.Name == "" {
			continue
		}

		quote, err := s.client.GetQuote(sym.Symbol)
		callsThisMinute++
		if err != nil || quote.CurrentPrice == 0 {
			continue
		}

		financials, err := s.client.GetBasicFinancials(sym.Symbol)
		callsThisMinute++
		if err != nil {
			continue
		}

		listing := &model.Listing{
			Ticker:      sym.Symbol,
			Name:        profile.Name,
			ExchangeMIC: profile.Exchange,
			LastRefresh: time.Now(),
			Price:       quote.CurrentPrice,
			Ask:         quote.High,
			ListingType: model.ListingTypeStock,
		}
		if err := s.listingRepo.Upsert(ctx, listing); err != nil {
			continue
		}

		stock := &model.Stock{
			ListingID:         listing.ListingID,
			OutstandingShares: profile.ShareOutstanding,
			DividendYield:     financials.Metric.DividendYieldIndicatedAnnual,
		}
		if err := s.stockRepo.Upsert(ctx, stock); err != nil {
			continue
		}

		count++
		log.Printf("[seed] [%d/%d] seeded %s", count, limit, sym.Symbol)
	}

	log.Printf("[seed] done. seeded %d stocks.", count)
	return nil
}

func stringsContainsDot(s string) bool {
	for _, c := range s {
		if c == '.' {
			return true
		}
	}
	return false
}

func (s *StockService) SeedOptions(ctx context.Context, limit int) error {
  stocks, err := s.stockRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load stocks: %w", err)
	}

	stockCount := 0
	optionCount := 0

	for _, stock := range stocks {
		if stockCount >= limit {
			break
		}
		if stock.Listing.Price == 0 {
			continue
		}

		currentPrice := stock.Listing.Price
		basePrice := roundToInt(currentPrice)

		var strikes []float64
		for i := -5; i <= 5; i++ {
			strikes = append(strikes, float64(basePrice+i))
		}

		expirations := generateExpirationDates()

		seeded := 0
		for _, exp := range expirations {
			for _, strike := range strikes {
				s.seedGeneratedOption(ctx, stock.Listing, strike, exp, model.OptionTypeCall, stock.StockID)
				s.seedGeneratedOption(ctx, stock.Listing, strike, exp, model.OptionTypePut, stock.StockID)
				seeded += 2
			}
		}

		optionCount += seeded
		stockCount++
		log.Printf("[seed-options] [%d/%d] seeded %d options for %s", stockCount, limit, seeded, stock.Listing.Ticker)
	}

	log.Printf("[seed-options] done. seeded %d options across %d stocks.", optionCount, stockCount)
	return nil
}

func roundToInt(price float64) int {
	if price-float64(int(price)) >= 0.5 {
		return int(price) + 1
	}
	return int(price)
}

func generateExpirationDates() []time.Time {
	var dates []time.Time
	now := time.Now()
	for i := 6; i <= 30; i += 6 {
		dates = append(dates, now.AddDate(0, 0, i))
	}
	start := dates[len(dates)-1].AddDate(0, 0, 30)
	for i := 0; i < 6; i++ {
		dates = append(dates, start.AddDate(0, 0, 30*i))
	}
	return dates
}

func (s *StockService) seedGeneratedOption(
	ctx context.Context,
	stockListing model.Listing,
	strike float64,
	expiration time.Time,
	optType model.OptionType,
	stockID uint,
) {
	ticker := fmt.Sprintf("%s:%s:%.2f", stockListing.Ticker, optType, strike)

	listing := &model.Listing{
		Ticker:      ticker,
		Name:        fmt.Sprintf("%s %s %.2f %s", stockListing.Ticker, optType, strike, expiration.Format("2006-01-02")),
		ExchangeMIC: stockListing.ExchangeMIC,
		LastRefresh: time.Now(),
		// TODO: replace with actual price calculation from black scholes
		Price:       strike,
		Ask:         strike,
		ListingType: model.ListingTypeOption,
	}
	if err := s.listingRepo.Upsert(ctx, listing); err != nil {
		return
	}

	option := &model.Option{
		ListingID:         listing.ListingID,
		StockID:           stockID, 
		OptionType:        optType,
		StrikePrice:       strike,
		ContractSize:      100,
		SettlementDate:    expiration,
		ImpliedVolatility: 1.0,
		OpenInterest:      0,
	}
	if err := s.optionRepo.Upsert(ctx, option); err != nil {
		return
	}
}

func (s *StockService) RefreshPrices(ctx context.Context) error {
	listings, err := s.listingRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load listings: %w", err)
	}

	callsThisMinute := 0
	minuteStart := time.Now()

	for _, listing := range listings {
		if callsThisMinute+1 > maxCallsPerMinute {
			elapsed := time.Since(minuteStart)
			if elapsed < time.Minute {
				wait := time.Minute - elapsed
				if err := waitForNextCall(ctx, wait); err != nil {
					return err
				}
			}
			callsThisMinute = 0
			minuteStart = time.Now()
		}

		quote, err := s.client.GetQuote(listing.Ticker)
		callsThisMinute++
		if err != nil || quote.CurrentPrice == 0 {
			continue
		}

		if err := s.listingRepo.UpdatePriceAndAsk(ctx, &listing, quote.CurrentPrice, quote.High); err != nil {
			continue
		}
	}

	return nil
}

func (s *StockService) RefreshOptions(ctx context.Context) error {
	stocks, err := s.stockRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load stocks: %w", err)
	}

	for _, stock := range stocks {
		if stringsContainsColon(stock.Listing.Ticker) {
			continue
		}

		expirations := generateExpirationDates()
		basePrice := roundToInt(stock.Listing.Price)

		var strikes []float64
		for i := -5; i <= 5; i++ {
			strikes = append(strikes, float64(basePrice+i))
		}

		for _, exp := range expirations {
			for _, strike := range strikes {
				s.seedGeneratedOption(ctx, stock.Listing, strike, exp, model.OptionTypeCall, stock.StockID)
				s.seedGeneratedOption(ctx, stock.Listing, strike, exp, model.OptionTypePut, stock.StockID)
			}
		}
	}

	return nil
}

func stringsContainsColon(s string) bool {
	for _, c := range s {
		if c == ':' {
			return true
		}
	}
	return false
}

func waitForNextCall(ctx context.Context, wait time.Duration) error {
	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
