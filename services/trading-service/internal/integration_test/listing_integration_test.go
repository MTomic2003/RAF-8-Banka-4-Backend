//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

func TestGetStocks(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XNYS")
	listing := seedListing(t, db, "AAPL", ex.MicCode, model.AssetTypeStock, 150.0)
	seedStock(t, db, listing.ListingID)
	seedDailyPriceInfo(t, db, listing.ListingID)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, "/api/listings/stocks?page=1&page_size=10", nil, auth)
	requireStatus(t, rec, http.StatusOK)
	require.Contains(t, rec.Body.String(), "AAPL")
}

func TestGetStocks_AsClient(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XNYS")
	listing := seedListing(t, db, "MSFT", ex.MicCode, model.AssetTypeStock, 300.0)
	seedStock(t, db, listing.ListingID)

	auth := authHeaderForClient(t, 50, 1)

	rec := performRequest(t, router, http.MethodGet, "/api/listings/stocks", nil, auth)
	requireStatus(t, rec, http.StatusOK)
}

func TestGetStocks_Unauthorized(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	rec := performRequest(t, router, http.MethodGet, "/api/listings/stocks", nil, "")
	require.NotEqual(t, http.StatusOK, rec.Code)
}

func TestGetStockDetails(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XNYS")
	listing := seedListing(t, db, "GOOG", ex.MicCode, model.AssetTypeStock, 140.0)
	seedStock(t, db, listing.ListingID)
	seedDailyPriceInfo(t, db, listing.ListingID)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, fmt.Sprintf("/api/listings/stocks/%d", listing.ListingID), nil, auth)
	requireStatus(t, rec, http.StatusOK)
	require.Contains(t, rec.Body.String(), "GOOG")
}

func TestGetStockDetails_NotFound(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, "/api/listings/stocks/99999", nil, auth)
	require.NotEqual(t, http.StatusOK, rec.Code)
}

func TestGetStockDetails_InvalidID(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, "/api/listings/stocks/abc", nil, auth)
	require.NotEqual(t, http.StatusOK, rec.Code)
}

func TestGetFutures(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XCME")
	listing := seedListing(t, db, "CLF25", ex.MicCode, model.AssetTypeFuture, 75.0)
	seedFuture(t, db, listing.ListingID)

	auth := authHeaderForAgent(t)

	rec := performRequest(t, router, http.MethodGet, "/api/listings/futures", nil, auth)
	requireStatus(t, rec, http.StatusOK)
}

func TestGetFutureDetails(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XCME")
	listing := seedListing(t, db, "ESH26", ex.MicCode, model.AssetTypeFuture, 5000.0)
	seedFuture(t, db, listing.ListingID)
	seedDailyPriceInfo(t, db, listing.ListingID)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, fmt.Sprintf("/api/listings/futures/%d", listing.ListingID), nil, auth)
	requireStatus(t, rec, http.StatusOK)
	require.Contains(t, rec.Body.String(), "ESH26")
}

func TestGetForex(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XFOR")
	listing := seedListing(t, db, "EURUSD", ex.MicCode, model.AssetTypeForexPair, 1.08)
	seedForex(t, db, listing.ListingID)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, "/api/listings/forex", nil, auth)
	requireStatus(t, rec, http.StatusOK)
}

func TestGetForexDetails(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XFOR")
	listing := seedListing(t, db, "GBPUSD", ex.MicCode, model.AssetTypeForexPair, 1.27)
	seedForex(t, db, listing.ListingID)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, fmt.Sprintf("/api/listings/forex/%d", listing.ListingID), nil, auth)
	requireStatus(t, rec, http.StatusOK)
}

func TestGetOptions(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XNYS")
	stockListing := seedListing(t, db, "TSLA", ex.MicCode, model.AssetTypeStock, 250.0)
	stock := seedStock(t, db, stockListing.ListingID)

	optListing := seedListing(t, db, "TSLA250C", ex.MicCode, model.AssetTypeOption, 15.0)
	seedOption(t, db, optListing.ListingID, stock.StockID)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, "/api/listings/options", nil, auth)
	requireStatus(t, rec, http.StatusOK)
}

func TestGetOptionDetails(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XNYS")
	stockListing := seedListing(t, db, "NVDA", ex.MicCode, model.AssetTypeStock, 800.0)
	stock := seedStock(t, db, stockListing.ListingID)

	optListing := seedListing(t, db, "NVDA850C", ex.MicCode, model.AssetTypeOption, 30.0)
	seedOption(t, db, optListing.ListingID, stock.StockID)
	seedDailyPriceInfo(t, db, optListing.ListingID)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, fmt.Sprintf("/api/listings/options/%d", optListing.ListingID), nil, auth)
	requireStatus(t, rec, http.StatusOK)
}

func TestGetForex_Unauthorized(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForClient(t, 50, 1)
	rec := performRequest(t, router, http.MethodGet, "/api/listings/forex", nil, auth)
	require.NotEqual(t, http.StatusOK, rec.Code)
}
