package service

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

// setupPriceAlertTestDB returns a fresh in-memory sqlite DB with the schema
// needed by PriceAlertService — same pattern watchlist_service_test.go uses.
func setupPriceAlertTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:testdb_pricealert_" + time.Now().Format("150405.000000000") + "?mode=memory&_pragma=foreign_keys(1)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	require.NoError(t, db.AutoMigrate(
		&model.Exchange{},
		&model.Asset{},
		&model.Listing{},
		&model.Stock{},
		&model.ListingDailyPriceInfo{},
		&model.PriceAlert{},
	))
	return db
}

func seedListingForAlert(t *testing.T, db *gorm.DB, ticker string, price float64) *model.Listing {
	t.Helper()
	require.NoError(t, db.Create(&model.Exchange{
		Name: "Test Exchange", Acronym: "TST", MicCode: "TST",
		Currency: "USD", TimeZone: -5, OpenTime: "09:30", CloseTime: "16:00",
		TradingEnabled: true,
	}).Error)
	asset := &model.Asset{Ticker: ticker, Name: ticker, AssetType: model.AssetTypeStock}
	require.NoError(t, db.Create(asset).Error)
	listing := &model.Listing{
		AssetID:     asset.AssetID,
		ExchangeMIC: "TST",
		LastRefresh: time.Now(),
		Price:       price,
		Ask:         price * 1.01,
	}
	require.NoError(t, db.Create(listing).Error)
	return listing
}

func newPriceAlertSvcForTest(t *testing.T, db *gorm.DB) (*PriceAlertService, *recordingMailer) {
	t.Helper()
	alertRepo := repository.NewPriceAlertRepository(db)
	listingRepo := repository.NewListingRepository(db)
	mailer := &recordingMailer{}
	notifier := NewNotificationService(mailer, &notifyUserClient{clientEmail: "alice@example.com"})
	return NewPriceAlertService(alertRepo, listingRepo, notifier), mailer
}

func TestPriceAlertService_CreateAlert_Persists(t *testing.T) {
	t.Parallel()
	db := setupPriceAlertTestDB(t)
	listing := seedListingForAlert(t, db, "AAPL", 150.0)
	svc, _ := newPriceAlertSvcForTest(t, db)

	resp, err := svc.CreateAlert(clientCtx(1), dto.CreatePriceAlertRequest{
		ListingID: listing.ListingID,
		Condition: "ABOVE",
		Threshold: 200,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotZero(t, resp.PriceAlertID)
	assert.Equal(t, "ABOVE", resp.Condition)
	assert.InEpsilon(t, 200.0, resp.Threshold, 1e-9)
	assert.Equal(t, "AAPL", resp.Ticker)
	assert.True(t, resp.IsActive)
	assert.Equal(t, "EMAIL", resp.NotificationType)
}

func TestPriceAlertService_CreateAlert_MissingListing_Returns404(t *testing.T) {
	t.Parallel()
	db := setupPriceAlertTestDB(t)
	svc, _ := newPriceAlertSvcForTest(t, db)

	_, err := svc.CreateAlert(clientCtx(1), dto.CreatePriceAlertRequest{
		ListingID: 9999,
		Condition: "ABOVE",
		Threshold: 100,
	})
	require.Error(t, err)
	// commonErrors.NotFoundErr surfaces via the gin error chain; we only check
	// the message here since the AppError type is exercised end-to-end in the
	// integration tests.
	assert.Contains(t, err.Error(), "listing not found")
}

func TestPriceAlertService_CreateAlert_InvalidCondition(t *testing.T) {
	t.Parallel()
	db := setupPriceAlertTestDB(t)
	listing := seedListingForAlert(t, db, "AAPL", 150.0)
	svc, _ := newPriceAlertSvcForTest(t, db)

	_, err := svc.CreateAlert(clientCtx(1), dto.CreatePriceAlertRequest{
		ListingID: listing.ListingID,
		Condition: "SIDEWAYS",
		Threshold: 100,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ABOVE or BELOW")
}

func TestPriceAlertService_ListMyAlerts_FiltersByOwner(t *testing.T) {
	t.Parallel()
	db := setupPriceAlertTestDB(t)
	listing := seedListingForAlert(t, db, "AAPL", 150.0)
	svc, _ := newPriceAlertSvcForTest(t, db)

	// Client 1's alert.
	_, err := svc.CreateAlert(clientCtx(1), dto.CreatePriceAlertRequest{
		ListingID: listing.ListingID, Condition: "ABOVE", Threshold: 200,
	})
	require.NoError(t, err)
	// Client 2's alert.
	_, err = svc.CreateAlert(clientCtx(2), dto.CreatePriceAlertRequest{
		ListingID: listing.ListingID, Condition: "BELOW", Threshold: 100,
	})
	require.NoError(t, err)

	mine, err := svc.ListMyAlerts(clientCtx(1))
	require.NoError(t, err)
	require.Len(t, mine, 1)
	assert.Equal(t, "ABOVE", mine[0].Condition)
}

func TestPriceAlertService_DeleteAlert_NonOwnerReturns404(t *testing.T) {
	t.Parallel()
	db := setupPriceAlertTestDB(t)
	listing := seedListingForAlert(t, db, "AAPL", 150.0)
	svc, _ := newPriceAlertSvcForTest(t, db)

	resp, err := svc.CreateAlert(clientCtx(1), dto.CreatePriceAlertRequest{
		ListingID: listing.ListingID, Condition: "ABOVE", Threshold: 200,
	})
	require.NoError(t, err)

	// Client 2 cannot delete client 1's alert — hidden as 404 rather than 403.
	err = svc.DeleteAlert(clientCtx(2), resp.PriceAlertID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "price alert not found")

	// Owner can delete it.
	require.NoError(t, svc.DeleteAlert(clientCtx(1), resp.PriceAlertID))
}

func TestPriceAlertService_CheckAndFire_AboveThreshold_FiresAndAutoDeactivates(t *testing.T) {
	t.Parallel()
	db := setupPriceAlertTestDB(t)
	listing := seedListingForAlert(t, db, "AAPL", 150.0)
	svc, mailer := newPriceAlertSvcForTest(t, db)

	// Threshold 100, current price 150 → ABOVE condition met.
	resp, err := svc.CreateAlert(clientCtx(1), dto.CreatePriceAlertRequest{
		ListingID: listing.ListingID, Condition: "ABOVE", Threshold: 100,
	})
	require.NoError(t, err)

	require.NoError(t, svc.CheckAndFire(context.Background()))

	// Email landed.
	sent := waitForSends(t, mailer, 1)
	require.Len(t, sent, 1)
	assert.Contains(t, sent[0].subject, "AAPL")

	// Alert was deactivated and TriggeredAt stamped.
	mine, err := svc.ListMyAlerts(clientCtx(1))
	require.NoError(t, err)
	require.Len(t, mine, 1)
	assert.False(t, mine[0].IsActive)
	require.NotNil(t, mine[0].TriggeredAt)

	// A second check must not re-fire (alert is now inactive).
	mailer.mu.Lock()
	mailer.sent = nil
	mailer.mu.Unlock()
	require.NoError(t, svc.CheckAndFire(context.Background()))
	time.Sleep(20 * time.Millisecond)
	assert.Empty(t, mailer.snapshot(), "inactive alert must not re-fire")

	_ = resp
}

func TestPriceAlertService_CheckAndFire_BelowThreshold_DoesNotFireWhenAbove(t *testing.T) {
	t.Parallel()
	db := setupPriceAlertTestDB(t)
	listing := seedListingForAlert(t, db, "AAPL", 150.0)
	svc, mailer := newPriceAlertSvcForTest(t, db)

	// Current price 150, threshold 100 with BELOW → condition NOT met.
	_, err := svc.CreateAlert(clientCtx(1), dto.CreatePriceAlertRequest{
		ListingID: listing.ListingID, Condition: "BELOW", Threshold: 100,
	})
	require.NoError(t, err)

	require.NoError(t, svc.CheckAndFire(context.Background()))
	time.Sleep(20 * time.Millisecond)
	assert.Empty(t, mailer.snapshot(), "BELOW alert above its threshold must not fire")

	mine, err := svc.ListMyAlerts(clientCtx(1))
	require.NoError(t, err)
	require.Len(t, mine, 1)
	assert.True(t, mine[0].IsActive, "alert must stay active when condition is not met")
}
