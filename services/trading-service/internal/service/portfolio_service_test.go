package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
	"github.com/stretchr/testify/require"
)

var errTest = errors.New("repo error")

// --- Fake repos ---

type fakeAssetOwnershipRepo struct {
	ownerships []model.AssetOwnership
	err        error
}

func (r *fakeAssetOwnershipRepo) FindByIdentity(_ context.Context, _ uint, _ model.OwnerType) ([]model.AssetOwnership, error) {
	return r.ownerships, r.err
}

func (r *fakeAssetOwnershipRepo) Upsert(_ context.Context, _ *model.AssetOwnership) error {
	return nil
}

type fakeStockRepo struct {
	stocks []model.Stock
	err    error
}

func (r *fakeStockRepo) Upsert(_ context.Context, _ *model.Stock) error { return nil }
func (r *fakeStockRepo) FindAll(_ context.Context) ([]model.Stock, error) { return nil, nil }
func (r *fakeStockRepo) FindByAssetIDs(_ context.Context, _ []uint) ([]model.Stock, error) {
	return r.stocks, r.err
}
func (r *fakeStockRepo) Count(_ context.Context) (int64, error) { return 0, nil }

type fakeOptionRepo struct {
	options []model.Option
	err     error
}

func (r *fakeOptionRepo) Upsert(_ context.Context, _ *model.Option) error { return nil }
func (r *fakeOptionRepo) FindByAssetIDs(_ context.Context, _ []uint) ([]model.Option, error) {
	return r.options, r.err
}
func (r *fakeOptionRepo) FindByStockID(_ context.Context, _ uint) ([]model.Option, error) {
	return nil, nil
}

type fakeFuturesRepo struct {
	futures []model.FuturesContract
	err     error
}

func (r *fakeFuturesRepo) FindByAssetIDs(_ context.Context, _ []uint) ([]model.FuturesContract, error) {
	return r.futures, r.err
}

type fakeForexRepo struct {
	forex []model.ForexPair
	err   error
}

func (r *fakeForexRepo) FindByAssetIDs(_ context.Context, _ []uint) ([]model.ForexPair, error) {
	return r.forex, r.err
}
func (r *fakeForexRepo) Count(_ context.Context) (int64, error) { return int64(len(r.forex)), nil }
func (r *fakeForexRepo) Upsert(_ context.Context, _ model.ForexPair) error { return nil }
func (r *fakeForexRepo) FindAll(_ context.Context, _ repository.ListingFilter) ([]model.ForexPair, int64, error) {
	return r.forex, int64(len(r.forex)), r.err
}

// --- Helpers ---

func ptrF(f float64) *float64 { return &f }

func makeOwnership(assetID uint, ticker string, amount, avgBuyPrice float64) model.AssetOwnership {
	return model.AssetOwnership{
		IdentityID:    1,
		OwnerType:     model.OwnerTypeClient,
		AssetID:       assetID,
		Asset:         model.Asset{AssetID: assetID, Ticker: ticker, AssetType: model.AssetTypeStock},
		AccountNumber: "444000100000000110",
		Amount:        amount,
		AvgBuyPrice:   avgBuyPrice,
		UpdatedAt:     time.Now(),
	}
}

func makeListing(assetID uint, price float64) *model.Listing {
	return &model.Listing{
		ListingID: assetID,
		AssetID:   assetID,
		Price:     price,
	}
}

// --- Tests ---

func TestGetPortfolio_HappyPath_Stock(t *testing.T) {
	ownership := makeOwnership(10, "AAPL", 10, 100.0)

	svc := NewPortfolioService(
		&fakeAssetOwnershipRepo{ownerships: []model.AssetOwnership{ownership}},
		&fakeStockRepo{stocks: []model.Stock{{StockID: 1, AssetID: 10, OutstandingShares: 1_000_000, Listing: makeListing(10, 150.0)}}},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
		&fakeForexRepo{},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Len(t, result, 1)

	a := result[0]
	require.Equal(t, dto.AssetTypeStock, a.Type)
	require.Equal(t, "AAPL", a.Ticker)
	require.Equal(t, float64(10), a.Amount)
	require.Equal(t, 150.0, a.PricePerUnit)
	require.InDelta(t, (150.0-100.0)*10, a.Profit, 0.001)
	require.NotNil(t, a.OutstandingShares)
	require.Equal(t, float64(1_000_000), *a.OutstandingShares)
}

func TestGetPortfolio_HappyPath_Option(t *testing.T) {
	ownership := makeOwnership(20, "MSFT220404C00180000", 200, 5.0)
	ownership.Asset.AssetType = model.AssetTypeOption

	svc := NewPortfolioService(
		&fakeAssetOwnershipRepo{ownerships: []model.AssetOwnership{ownership}},
		&fakeStockRepo{},
		&fakeOptionRepo{options: []model.Option{{OptionID: 1, AssetID: 20, Listing: makeListing(20, 8.0)}}},
		&fakeFuturesRepo{},
		&fakeForexRepo{},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Len(t, result, 1)

	a := result[0]
	require.Equal(t, dto.AssetTypeOption, a.Type)
	require.Equal(t, float64(200), a.Amount)
	require.InDelta(t, (8.0-5.0)*200, a.Profit, 0.001)
	require.Nil(t, a.OutstandingShares)
}

func TestGetPortfolio_HappyPath_Futures(t *testing.T) {
	ownership := makeOwnership(30, "CLJ22", 5, 200.0)
	ownership.Asset.AssetType = model.AssetTypeFuture

	svc := NewPortfolioService(
		&fakeAssetOwnershipRepo{ownerships: []model.AssetOwnership{ownership}},
		&fakeStockRepo{},
		&fakeOptionRepo{},
		&fakeFuturesRepo{futures: []model.FuturesContract{{FuturesContractID: 1, AssetID: 30, Listing: makeListing(30, 210.0)}}},
		&fakeForexRepo{},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Len(t, result, 1)

	a := result[0]
	require.Equal(t, dto.AssetTypeFutures, a.Type)
	require.Equal(t, float64(5), a.Amount)
	require.InDelta(t, (210.0-200.0)*5, a.Profit, 0.001)
}

func TestGetPortfolio_ZeroAmountFiltered(t *testing.T) {
	// Ownership with 0 amount should be excluded
	ownership := makeOwnership(10, "AAPL", 0, 100.0)

	svc := NewPortfolioService(
		&fakeStockRepo{stocks: []model.Stock{{StockID: 1, ListingID: 10}}},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
		&fakeForexRepo{},
		&fakeAssetOwnershipRepo{ownerships: []model.AssetOwnership{ownership}}
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestGetPortfolio_NetAmountAfterSell(t *testing.T) {
	buy := makeOrder(1, 10, model.OrderDirectionBuy, model.OrderStatusApproved, 10, 10, 100.0, 1.0)
	buy.Listing.Ticker = "AAPL"
	buy.Listing.Price = 150.0
	sell := makeOrder(2, 10, model.OrderDirectionSell, model.OrderStatusApproved, 10, 10, 140.0, 1.0)
	sell.Listing = buy.Listing

	svc := NewPortfolioService(
		&fakeOwnershipRepo{ownerships: []model.OrderOwnership{makeOwnership(buy), makeOwnership(sell)}},
		&fakeStockRepo{stocks: []model.Stock{{StockID: 1, ListingID: 10}}},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
		&fakeForexRepo{},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestGetPortfolio_PartialSell(t *testing.T) {
	buy := makeOrder(1, 10, model.OrderDirectionBuy, model.OrderStatusApproved, 10, 10, 100.0, 1.0)
	buy.Listing.Ticker = "AAPL"
	buy.Listing.Price = 150.0
	sell := makeOrder(2, 10, model.OrderDirectionSell, model.OrderStatusApproved, 4, 4, 130.0, 1.0)
	sell.Listing = buy.Listing

	svc := NewPortfolioService(
		&fakeOwnershipRepo{ownerships: []model.OrderOwnership{makeOwnership(buy), makeOwnership(sell)}},
		&fakeStockRepo{stocks: []model.Stock{{StockID: 1, ListingID: 10}}},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
		&fakeForexRepo{},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, float64(6), result[0].Amount)
}

func TestGetPortfolio_ForexExcluded(t *testing.T) {
	ord := makeOrder(1, 40, model.OrderDirectionBuy, model.OrderStatusApproved, 5, 5, 1.2, 1000.0)
	ord.Listing.Ticker = "EUR/USD"
	ord.Listing.Price = 1.25

	svc := NewPortfolioService(
		&fakeOwnershipRepo{ownerships: []model.OrderOwnership{makeOwnership(ord)}},
		&fakeStockRepo{},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
		&fakeForexRepo{},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestGetPortfolio_EmptyOwnerships(t *testing.T) {
	svc := NewPortfolioService(
		&fakeAssetOwnershipRepo{ownerships: []model.AssetOwnership{}},
		&fakeStockRepo{},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
		&fakeForexRepo{},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestGetPortfolio_RepoError(t *testing.T) {
	svc := NewPortfolioService(
		&fakeAssetOwnershipRepo{err: errTest},
		&fakeStockRepo{},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
		&fakeForexRepo{},
	)

	_, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.Error(t, err)
}

func TestGetPortfolio_NegativeProfit_NoTax(t *testing.T) {
	// Buy at 200, current price 150 -> negative profit, no tax
	ownership := makeOwnership(10, "AAPL", 20, 200.0)

	svc := NewPortfolioService(
		&fakeAssetOwnershipRepo{ownerships: []model.AssetOwnership{ownership}},
		&fakeStockRepo{stocks: []model.Stock{{StockID: 1, AssetID: 10, Listing: makeListing(10, 150.0)}}},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
		&fakeForexRepo{},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.InDelta(t, (150.0-200.0)*20, result[0].Profit, 0.001)
	require.InDelta(t, 0.0, result[0].TaxAmount, 0.001)
	require.Equal(t, float64(20), result[0].Amount)
}
