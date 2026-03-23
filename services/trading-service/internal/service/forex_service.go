package service

import (
	"context"
	"log"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"gorm.io/gorm"
)

const refreshInterval = 1 * time.Hour

type ForexService struct {
	db     *gorm.DB
	client client.ExchangeRateClient
}

func NewForexService(db *gorm.DB, client client.ExchangeRateClient) *ForexService {
	return &ForexService{
		db:     db,
		client: client,
	}
}

func (s *ForexService) Initialize(ctx context.Context) {
	var count int64

	if err := s.db.WithContext(ctx).
		Model(&model.ForexPair{}).
		Count(&count).Error; err != nil {
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

func (s *ForexService) StartBackgroundRefresh(ctx context.Context) {
	ticker := time.NewTicker(refreshInterval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := s.refreshFromAPI(ctx); err != nil {
					log.Println("forex refresh failed:", err)
				}
			}
		}
	}()
}

func (s *ForexService) refreshFromAPI(ctx context.Context) error {
	resp, err := s.client.FetchRates(ctx)
	if err != nil {
		return err
	}

	providerUpdatedAt := time.Unix(resp.TimeLastUpdateUnix, 0)
	providerNextUpdateAt := time.Unix(resp.TimeNextUpdateUnix, 0)

	for quote, rate := range resp.ConversionRates {

		pair := model.ForexPair{
			Base:                 resp.BaseCode,
			Quote:                quote,
			Rate:                 rate,
			ProviderUpdatedAt:    providerUpdatedAt,
			ProviderNextUpdateAt: providerNextUpdateAt,
		}

		if err := s.db.WithContext(ctx).
			Where("base = ? AND quote = ?", pair.Base, pair.Quote).
			Assign(pair).
			FirstOrCreate(&pair).Error; err != nil {
			return err
		}
	}

	log.Println("forex pairs refreshed from API")
	return nil
}
