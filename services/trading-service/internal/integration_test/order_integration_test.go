//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

func TestCreateOrder(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XNYS")
	listing := seedListing(t, db, "AAPL", ex.MicCode, model.AssetTypeStock, 150.0)
	seedStock(t, db, listing.ListingID)

	auth := authHeaderForSupervisor(t)

	body := map[string]any{
		"listing_id":     listing.ListingID,
		"order_type":     "MARKET",
		"direction":      "BUY",
		"quantity":       5,
		"account_number": "444000100000000001",
	}

	rec := performRequest(t, router, http.MethodPost, "/api/orders", body, auth)
	requireStatus(t, rec, http.StatusCreated)

	resp := decodeResponse[map[string]any](t, rec)
	require.Equal(t, "MARKET", resp["order_type"])
	require.Equal(t, "BUY", resp["direction"])
	require.Equal(t, float64(5), resp["quantity"])
}

func TestCreateOrder_LimitOrder(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XNAS")
	listing := seedListing(t, db, "MSFT", ex.MicCode, model.AssetTypeStock, 400.0)
	seedStock(t, db, listing.ListingID)

	auth := authHeaderForSupervisor(t)

	body := map[string]any{
		"listing_id":     listing.ListingID,
		"order_type":     "LIMIT",
		"direction":      "SELL",
		"quantity":       10,
		"limit_value":    405.0,
		"account_number": "444000100000000001",
	}

	rec := performRequest(t, router, http.MethodPost, "/api/orders", body, auth)
	requireStatus(t, rec, http.StatusCreated)

	resp := decodeResponse[map[string]any](t, rec)
	require.Equal(t, "LIMIT", resp["order_type"])
}

func TestCreateOrder_Unauthorized(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	body := map[string]any{
		"listing_id":     1,
		"order_type":     "MARKET",
		"direction":      "BUY",
		"quantity":       5,
		"account_number": "444000100000000001",
	}

	rec := performRequest(t, router, http.MethodPost, "/api/orders", body, "")
	require.NotEqual(t, http.StatusCreated, rec.Code)
}

func TestCreateOrder_InvalidBody(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodPost, "/api/orders", map[string]any{}, auth)
	require.NotEqual(t, http.StatusCreated, rec.Code)
}

func TestGetOrders(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XNYS")
	listing := seedListing(t, db, "GOOG", ex.MicCode, model.AssetTypeStock, 140.0)
	seedStock(t, db, listing.ListingID)
	seedOrder(t, db, 10, listing.ListingID, model.OrderDirectionBuy, model.OrderStatusApproved)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, "/api/orders?page=1&page_size=10", nil, auth)
	requireStatus(t, rec, http.StatusOK)

	resp := decodeResponse[map[string]any](t, rec)
	data := resp["data"].([]any)
	require.GreaterOrEqual(t, len(data), 1)
}

func TestGetOrders_DefaultPagination(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodGet, "/api/orders", nil, auth)
	requireStatus(t, rec, http.StatusOK)
}

func TestGetOrders_ForbiddenForNonSupervisor(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForAgent(t)

	rec := performRequest(t, router, http.MethodGet, "/api/orders", nil, auth)
	require.NotEqual(t, http.StatusOK, rec.Code)
}

func TestApproveOrder(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XNYS")
	listing := seedListing(t, db, "META", ex.MicCode, model.AssetTypeStock, 500.0)
	seedStock(t, db, listing.ListingID)
	order := seedOrder(t, db, 20, listing.ListingID, model.OrderDirectionBuy, model.OrderStatusPending)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodPatch, fmt.Sprintf("/api/orders/%d/approve", order.OrderID), nil, auth)
	requireStatus(t, rec, http.StatusOK)

	resp := decodeResponse[map[string]any](t, rec)
	require.Equal(t, "APPROVED", resp["status"])
}

func TestApproveOrder_NotFound(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodPatch, "/api/orders/99999/approve", nil, auth)
	require.NotEqual(t, http.StatusOK, rec.Code)
}

func TestApproveOrder_InvalidID(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodPatch, "/api/orders/abc/approve", nil, auth)
	require.NotEqual(t, http.StatusOK, rec.Code)
}

func TestDeclineOrder(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XNAS")
	listing := seedListing(t, db, "AMZN", ex.MicCode, model.AssetTypeStock, 180.0)
	seedStock(t, db, listing.ListingID)
	order := seedOrder(t, db, 20, listing.ListingID, model.OrderDirectionSell, model.OrderStatusPending)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodPatch, fmt.Sprintf("/api/orders/%d/decline", order.OrderID), nil, auth)
	requireStatus(t, rec, http.StatusOK)

	resp := decodeResponse[map[string]any](t, rec)
	require.Equal(t, "DECLINED", resp["status"])
}

func TestCancelOrder(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XNYS")
	listing := seedListing(t, db, "NFLX", ex.MicCode, model.AssetTypeStock, 600.0)
	seedStock(t, db, listing.ListingID)
	order := seedOrder(t, db, 10, listing.ListingID, model.OrderDirectionBuy, model.OrderStatusApproved)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodPatch, fmt.Sprintf("/api/orders/%d/cancel", order.OrderID), nil, auth)
	requireStatus(t, rec, http.StatusOK)

	resp := decodeResponse[map[string]any](t, rec)
	require.Equal(t, true, resp["is_done"])
}

func TestCancelOrder_AlreadyDeclined(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db)

	ex := seedExchange(t, db, "XNYS")
	listing := seedListing(t, db, "DIS", ex.MicCode, model.AssetTypeStock, 100.0)
	seedStock(t, db, listing.ListingID)
	order := seedOrder(t, db, 10, listing.ListingID, model.OrderDirectionBuy, model.OrderStatusDeclined)

	auth := authHeaderForSupervisor(t)

	rec := performRequest(t, router, http.MethodPatch, fmt.Sprintf("/api/orders/%d/cancel", order.OrderID), nil, auth)
	require.NotEqual(t, http.StatusOK, rec.Code)
}
