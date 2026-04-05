package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

const refreshInterval = 1 * time.Hour

type ForexService struct {
	repo        repository.ForexRepository
	listingRepo repository.ListingRepository
	client      client.ExchangeRateClient
	mu          sync.Mutex
	cancel      context.CancelFunc
}

func NewForexService(repo repository.ForexRepository, listingRepo repository.ListingRepository, client client.ExchangeRateClient) *ForexService {
	return &ForexService{
		repo:        repo,
		listingRepo: listingRepo,
		client:      client,
	}
}

func (s *ForexService) Initialize(ctx context.Context) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		log.Println("failed counting forex pairs:", err)
		return
	}

	if count > 0 {
		log.Println("forex pairs loaded from DB")
		return
	}

	if err := s.refreshFromAPI(ctx); err != nil {
		log.Println("initial forex fetch failed:", err)
	}
}

func (s *ForexService) Start() {
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return // već radi
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()

	ticker := time.NewTicker(refreshInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.refreshFromAPI(ctx); err != nil {
					log.Println("forex refresh failed:", err)
				}
			}
		}
	}()
}

func (s *ForexService) Stop() {
	s.mu.Lock()
	cancel := s.cancel
	s.cancel = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (s *ForexService) refreshFromAPI(ctx context.Context) error {
	resp, err := s.client.FetchRates(ctx)
	if err != nil {
		return err
	}

	providerUpdatedAt := time.Unix(resp.TimeLastUpdateUnix, 0)
	providerNextUpdateAt := time.Unix(resp.TimeNextUpdateUnix, 0)

	// sve valute koje podržava banka
	supported := []string{"EUR", "USD", "CHF", "GBP", "JPY", "CAD", "AUD", "RSD"}

	rates := resp.ConversionRates
	rates[resp.BaseCode] = 1.0

	for _, base := range supported {
		for _, quote := range supported {
			if base == quote {
				continue
			}

			baseRate, ok1 := rates[base]
			quoteRate, ok2 := rates[quote]

			if !ok1 || !ok2 {
				continue
			}

			rate := quoteRate / baseRate
			ticker := base + "/" + quote

			listing := &model.Listing{
				Ticker:      ticker,
				Name:        ticker,
				ExchangeMIC: model.SimulatedExchangeMIC,
				LastRefresh: time.Now(),
				Price:       rate,
				Ask:         rate,
				ListingType: model.ListingTypeForexPair,
			}
			if err := s.listingRepo.Upsert(ctx, listing); err != nil {
				return err
			}

			pair := model.ForexPair{
				ListingID:            listing.ListingID,
				Base:                 base,
				Quote:                quote,
				Rate:                 rate,
				ProviderUpdatedAt:    providerUpdatedAt,
				ProviderNextUpdateAt: providerNextUpdateAt,
			}

			if err := s.repo.Upsert(ctx, pair); err != nil {
				return err
			}
		}
	}

	log.Println("forex pairs refreshed from API")
	return nil
}
