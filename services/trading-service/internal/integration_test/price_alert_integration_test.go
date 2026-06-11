//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

// seedListingForPriceAlerts seeds an exchange + listing + stock so the route
// has a valid ListingID to point at. Mirrors how watchlist_integration_test.go
// composes its fixtures.
func seedListingForPriceAlerts(t *testing.T, db *gorm.DB, ticker string, price float64) *model.Listing {
	t.Helper()
	seedExchange(t, db, "XNYS")
	listing := seedListing(t, db, ticker, "XNYS", model.AssetTypeStock, price)
	seedStock(t, db, listing.ListingID)
	return listing
}

func TestPriceAlert_CreateAndList_AsClient(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)
	auth := authHeaderForClient(t, 50, 1)

	listing := seedListingForPriceAlerts(t, db, "AAPL", 150.0)

	rec := performRequest(t, router, http.MethodPost, "/api/price-alerts",
		dto.CreatePriceAlertRequest{
			ListingID: listing.ListingID,
			Condition: "ABOVE",
			Threshold: 200,
		}, auth)
	requireStatus(t, rec, http.StatusCreated)

	created := decodeResponse[dto.PriceAlertResponse](t, rec)
	assert.NotZero(t, created.PriceAlertID)
	assert.Equal(t, "ABOVE", created.Condition)
	assert.InEpsilon(t, 200.0, created.Threshold, 1e-9)
	assert.Equal(t, "AAPL", created.Ticker)
	assert.True(t, created.IsActive)

	rec = performRequest(t, router, http.MethodGet, "/api/price-alerts", nil, auth)
	requireStatus(t, rec, http.StatusOK)
	list := decodeResponse[[]dto.PriceAlertResponse](t, rec)
	require.Len(t, list, 1)
	assert.Equal(t, created.PriceAlertID, list[0].PriceAlertID)
}

func TestPriceAlert_Create_AsAgent(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)
	auth := authHeaderForAgent(t)

	listing := seedListingForPriceAlerts(t, db, "MSFT", 300.0)

	rec := performRequest(t, router, http.MethodPost, "/api/price-alerts",
		dto.CreatePriceAlertRequest{
			ListingID: listing.ListingID,
			Condition: "BELOW",
			Threshold: 250,
		}, auth)
	requireStatus(t, rec, http.StatusCreated)
}

func TestPriceAlert_Create_AsSupervisor_IsAllowed(t *testing.T) {
	// Same policy as watchlist — anyone with the Trading permission can use
	// the endpoint; the supervisor role is not gated out.
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)
	auth := authHeaderForSupervisor(t)

	listing := seedListingForPriceAlerts(t, db, "NVDA", 500.0)

	rec := performRequest(t, router, http.MethodPost, "/api/price-alerts",
		dto.CreatePriceAlertRequest{
			ListingID: listing.ListingID,
			Condition: "ABOVE",
			Threshold: 600,
		}, auth)
	requireStatus(t, rec, http.StatusCreated)
}

func TestPriceAlert_Create_MissingListing_Returns404(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)
	auth := authHeaderForClient(t, 50, 1)

	rec := performRequest(t, router, http.MethodPost, "/api/price-alerts",
		dto.CreatePriceAlertRequest{
			ListingID: 99999,
			Condition: "ABOVE",
			Threshold: 100,
		}, auth)
	requireStatus(t, rec, http.StatusNotFound)
}

func TestPriceAlert_Create_InvalidCondition_Returns400(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)
	auth := authHeaderForClient(t, 50, 1)

	listing := seedListingForPriceAlerts(t, db, "AAPL", 150.0)

	rec := performRequest(t, router, http.MethodPost, "/api/price-alerts",
		dto.CreatePriceAlertRequest{
			ListingID: listing.ListingID,
			Condition: "SIDEWAYS",
			Threshold: 100,
		}, auth)
	requireStatus(t, rec, http.StatusBadRequest)
}

func TestPriceAlert_Create_NegativeThreshold_Returns400(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)
	auth := authHeaderForClient(t, 50, 1)

	listing := seedListingForPriceAlerts(t, db, "AAPL", 150.0)

	rec := performRequest(t, router, http.MethodPost, "/api/price-alerts",
		dto.CreatePriceAlertRequest{
			ListingID: listing.ListingID,
			Condition: "ABOVE",
			Threshold: -10,
		}, auth)
	requireStatus(t, rec, http.StatusBadRequest)
}

func TestPriceAlert_Delete_OwnAlert_Returns204(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)
	auth := authHeaderForClient(t, 50, 1)

	listing := seedListingForPriceAlerts(t, db, "AAPL", 150.0)

	rec := performRequest(t, router, http.MethodPost, "/api/price-alerts",
		dto.CreatePriceAlertRequest{
			ListingID: listing.ListingID, Condition: "ABOVE", Threshold: 200,
		}, auth)
	requireStatus(t, rec, http.StatusCreated)
	created := decodeResponse[dto.PriceAlertResponse](t, rec)

	rec = performRequest(t, router, http.MethodDelete,
		"/api/price-alerts/"+itoa(created.PriceAlertID), nil, auth)
	requireStatus(t, rec, http.StatusNoContent)

	// Now listing the user's alerts must be empty.
	rec = performRequest(t, router, http.MethodGet, "/api/price-alerts", nil, auth)
	requireStatus(t, rec, http.StatusOK)
	list := decodeResponse[[]dto.PriceAlertResponse](t, rec)
	assert.Empty(t, list)
}

func TestPriceAlert_Delete_OthersAlert_Returns404(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	// Alice (client 1) creates the alert.
	aliceAuth := authHeaderForClient(t, 50, 1)
	listing := seedListingForPriceAlerts(t, db, "AAPL", 150.0)
	rec := performRequest(t, router, http.MethodPost, "/api/price-alerts",
		dto.CreatePriceAlertRequest{
			ListingID: listing.ListingID, Condition: "ABOVE", Threshold: 200,
		}, aliceAuth)
	requireStatus(t, rec, http.StatusCreated)
	created := decodeResponse[dto.PriceAlertResponse](t, rec)

	// Bob (client 2) tries to delete it — must be hidden as 404, not refused
	// with 403, matching the watchlist convention.
	bobAuth := authHeaderForClient(t, 51, 2)
	rec = performRequest(t, router, http.MethodDelete,
		"/api/price-alerts/"+itoa(created.PriceAlertID), nil, bobAuth)
	requireStatus(t, rec, http.StatusNotFound)
}

func TestPriceAlert_List_Unauthenticated_Returns401(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	rec := performRequest(t, router, http.MethodGet, "/api/price-alerts", nil, "")
	requireStatus(t, rec, http.StatusUnauthorized)
}

func TestPriceAlert_Delete_InvalidID_Returns400(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)
	auth := authHeaderForClient(t, 50, 1)

	rec := performRequest(t, router, http.MethodDelete, "/api/price-alerts/abc", nil, auth)
	requireStatus(t, rec, http.StatusBadRequest)
}

// itoa avoids pulling in strconv just for the small handful of conversions in
// this file. Kept local to keep the test file self-contained.
func itoa(n uint) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
