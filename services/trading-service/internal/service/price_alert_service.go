package service

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"go.uber.org/fx"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	commonErrors "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

// priceAlertCheckInterval is how often the scheduler scans active alerts.
// StockService.RefreshPrices updates listings on its own ticker; this one is
// independent and only reads the already-persisted listing.Price, so the two
// loops do not need to be coupled.
const priceAlertCheckInterval = 30 * time.Second

// PriceAlertService owns the lifecycle of personal price alerts and the
// background scheduler that fires them. The scheduler is decoupled from
// StockService: it just polls active alerts and compares against the listing
// row that StockService keeps fresh.
type PriceAlertService struct {
	alertRepo   repository.PriceAlertRepository
	listingRepo repository.ListingRepository
	notifier    *NotificationService

	stop   chan struct{}
	wg     sync.WaitGroup
	now    func() time.Time
	logger func(format string, args ...any)
}

func NewPriceAlertService(alertRepo repository.PriceAlertRepository, listingRepo repository.ListingRepository, notifier *NotificationService) *PriceAlertService {
	return &PriceAlertService{
		alertRepo:   alertRepo,
		listingRepo: listingRepo,
		notifier:    notifier,
		stop:        make(chan struct{}),
		now:         time.Now,
		logger:      log.Printf,
	}
}

// Start kicks off the periodic checker. It is wired through fx.Lifecycle in
// cmd/main.go so it begins at boot and is gracefully stopped on shutdown.
func (s *PriceAlertService) Start() {
	s.wg.Add(1)
	go s.run()
}

// Stop signals the scheduler goroutine and waits for it to exit.
func (s *PriceAlertService) Stop() {
	close(s.stop)
	s.wg.Wait()
}

func (s *PriceAlertService) run() {
	defer s.wg.Done()
	ticker := time.NewTicker(priceAlertCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), priceAlertCheckInterval)
			if err := s.CheckAndFire(ctx); err != nil {
				s.logger("[price-alerts] check tick failed: %v", err)
			}
			cancel()
		}
	}
}

// CheckAndFire loads every active alert and fires those whose condition is met
// against the current listing price. Exported so tests can drive the check
// deterministically without waiting for the ticker.
func (s *PriceAlertService) CheckAndFire(ctx context.Context) error {
	alerts, err := s.alertRepo.FindAllActive(ctx)
	if err != nil {
		return err
	}
	for i := range alerts {
		alert := alerts[i]
		if alert.Listing == nil {
			s.logger("[price-alerts] alert %d skipped: listing %d missing", alert.PriceAlertID, alert.ListingID)
			continue
		}
		currentPrice := alert.Listing.Price
		if !conditionMet(alert.Condition, currentPrice, alert.Threshold) {
			continue
		}
		if err := s.alertRepo.MarkTriggered(ctx, alert.PriceAlertID); err != nil {
			s.logger("[price-alerts] alert %d mark triggered: %v", alert.PriceAlertID, err)
			continue
		}
		s.notifier.NotifyPriceAlert(alert, alert.Listing, currentPrice)
	}
	return nil
}

// --- CRUD --------------------------------------------------------------------

func (s *PriceAlertService) CreateAlert(ctx context.Context, req dto.CreatePriceAlertRequest) (*dto.PriceAlertResponse, error) {
	userID, ownerType, err := priceAlertOwnerIdentity(ctx)
	if err != nil {
		return nil, err
	}

	listing, err := s.listingRepo.FindByID(ctx, req.ListingID, -1)
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}
	if listing == nil {
		return nil, commonErrors.NotFoundErr("listing not found")
	}

	condition := model.PriceAlertCondition(strings.ToUpper(strings.TrimSpace(req.Condition)))
	if condition != model.PriceAlertConditionAbove && condition != model.PriceAlertConditionBelow {
		return nil, commonErrors.BadRequestErr("condition must be ABOVE or BELOW")
	}

	alert := &model.PriceAlert{
		UserID:           userID,
		OwnerType:        ownerType,
		ListingID:        req.ListingID,
		Condition:        condition,
		Threshold:        req.Threshold,
		NotificationType: model.PriceAlertNotificationEmail,
		IsActive:         true,
		CreatedAt:        s.now(),
	}
	if err := s.alertRepo.Create(ctx, alert); err != nil {
		return nil, commonErrors.InternalErr(err)
	}
	alert.Listing = listing
	resp := toPriceAlertResponse(*alert)
	return &resp, nil
}

func (s *PriceAlertService) ListMyAlerts(ctx context.Context) ([]dto.PriceAlertResponse, error) {
	userID, ownerType, err := priceAlertOwnerIdentity(ctx)
	if err != nil {
		return nil, err
	}
	alerts, err := s.alertRepo.FindByOwner(ctx, userID, ownerType)
	if err != nil {
		return nil, commonErrors.InternalErr(err)
	}
	out := make([]dto.PriceAlertResponse, 0, len(alerts))
	for _, a := range alerts {
		out = append(out, toPriceAlertResponse(a))
	}
	return out, nil
}

func (s *PriceAlertService) DeleteAlert(ctx context.Context, id uint) error {
	userID, ownerType, err := priceAlertOwnerIdentity(ctx)
	if err != nil {
		return err
	}
	alert, err := s.alertRepo.FindByID(ctx, id)
	if err != nil {
		return commonErrors.InternalErr(err)
	}
	// Hide the resource from non-owners by reporting NotFound rather than
	// Forbidden — same convention used by the watchlist service.
	if alert == nil || alert.UserID != userID || alert.OwnerType != ownerType {
		return commonErrors.NotFoundErr("price alert not found")
	}
	if err := s.alertRepo.Delete(ctx, id); err != nil {
		return commonErrors.InternalErr(err)
	}
	return nil
}

// --- helpers -----------------------------------------------------------------

func conditionMet(condition model.PriceAlertCondition, current, threshold float64) bool {
	switch condition {
	case model.PriceAlertConditionAbove:
		return current >= threshold
	case model.PriceAlertConditionBelow:
		return current <= threshold
	default:
		return false
	}
}

func toPriceAlertResponse(a model.PriceAlert) dto.PriceAlertResponse {
	ticker := ""
	if a.Listing != nil && a.Listing.Asset != nil {
		ticker = a.Listing.Asset.Ticker
	}
	return dto.PriceAlertResponse{
		PriceAlertID:     a.PriceAlertID,
		ListingID:        a.ListingID,
		Ticker:           ticker,
		Condition:        string(a.Condition),
		Threshold:        a.Threshold,
		NotificationType: string(a.NotificationType),
		IsActive:         a.IsActive,
		CreatedAt:        a.CreatedAt,
		TriggeredAt:      a.TriggeredAt,
	}
}

// priceAlertOwnerIdentity mirrors watchlist's ownerIdentity helper. Kept local
// rather than shared in a util package to keep services self-contained.
func priceAlertOwnerIdentity(ctx context.Context) (uint, model.OwnerType, error) {
	authCtx := auth.GetAuthFromContext(ctx)
	if authCtx == nil {
		return 0, "", commonErrors.UnauthorizedErr("not authenticated")
	}
	switch authCtx.IdentityType {
	case auth.IdentityClient:
		if authCtx.ClientID == nil {
			return 0, "", commonErrors.UnauthorizedErr("client identity missing")
		}
		return *authCtx.ClientID, model.OwnerTypeClient, nil
	case auth.IdentityEmployee:
		if authCtx.EmployeeID == nil {
			return 0, "", commonErrors.UnauthorizedErr("employee identity missing")
		}
		return *authCtx.EmployeeID, model.OwnerTypeBank, nil
	default:
		return 0, "", commonErrors.ForbiddenErr("access denied for this identity type")
	}
}

// RegisterPriceAlertLifecycle wires Start/Stop into fx without requiring main.go
// to know about the scheduler internals.
func RegisterPriceAlertLifecycle(lc fx.Lifecycle, svc *PriceAlertService) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			svc.Start()
			return nil
		},
		OnStop: func(_ context.Context) error {
			svc.Stop()
			return nil
		},
	})
}
