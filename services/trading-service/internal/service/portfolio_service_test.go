package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/stretchr/testify/require"
)

var errTest = errors.New("repo error")

// --- Fake repos ---

type fakeOwnershipRepo struct {
	ownerships []model.OrderOwnership
	err        error
}

func (r *fakeOwnershipRepo) FindByIdentity(_ uint, _ model.OwnerType) ([]model.OrderOwnership, error) {
	return r.ownerships, r.err
}

type fakeStockRepo struct {
	stocks []model.Stock
	err    error
}

func (r *fakeStockRepo) Upsert(_ context.Context, _ *model.Stock) error { return nil }

func (r *fakeStockRepo) FindAll(_ context.Context) ([]model.Stock, error) { return nil, nil }

func (r *fakeStockRepo) FindByListingIDs(_ []uint) ([]model.Stock, error) {
	return r.stocks, r.err
}

type fakeOptionRepo struct {
	options []model.Option
	err     error
}

func (r *fakeOptionRepo) Upsert(_ context.Context, _ *model.Option) error { return nil }

func (r *fakeOptionRepo) FindByListingIDs(_ []uint) ([]model.Option, error) {
	return r.options, r.err
}

type fakeFuturesRepo struct {
	futures []model.FuturesContract
	err     error
}

func (r *fakeFuturesRepo) FindByListingIDs(_ []uint) ([]model.FuturesContract, error) {
	return r.futures, r.err
}

// --- Helpers ---

func ptrF(f float64) *float64 { return &f }

func makeOrder(id, listingID uint, dir model.OrderDirection, status model.OrderStatus, qty uint, filled uint, price float64, contractSize float64) model.Order {
	return model.Order{
		OrderID:      id,
		ListingID:    listingID,
		Direction:    dir,
		Status:       status,
		Quantity:     qty,
		FilledQty:    filled,
		PricePerUnit: ptrF(price),
		ContractSize: contractSize,
		Listing: model.Listing{
			ListingID: listingID,
			Ticker:    "TST",
			Price:     150.0,
		},
		UpdatedAt: time.Now(),
	}
}

func makeOwnership(order model.Order) model.OrderOwnership {
	return model.OrderOwnership{
		OrderID:       order.OrderID,
		Order:         order,
		IdentityID:    1,
		OwnerType:     model.OwnerTypeClient,
		AccountNumber: "444000100000000110",
	}
}

// --- Tests ---

func TestGetPortfolio_HappyPath_Stock(t *testing.T) {
	ord := makeOrder(1, 10, model.OrderDirectionBuy, model.OrderStatusApproved, 10, 10, 100.0, 1.0)
	ord.Listing.Ticker = "AAPL"
	ord.Listing.Price = 150.0

	svc := NewPortfolioService(
		&fakeOwnershipRepo{ownerships: []model.OrderOwnership{makeOwnership(ord)}},
		&fakeStockRepo{stocks: []model.Stock{{StockID: 1, ListingID: 10, OutstandingShares: 1_000_000}}},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
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
	// options: contract size 100, 2 contracts → amount = 200
	ord := makeOrder(2, 20, model.OrderDirectionBuy, model.OrderStatusApproved, 2, 2, 5.0, 100.0)
	ord.Listing.Ticker = "MSFT220404C00180000"
	ord.Listing.Price = 8.0

	svc := NewPortfolioService(
		&fakeOwnershipRepo{ownerships: []model.OrderOwnership{makeOwnership(ord)}},
		&fakeStockRepo{},
		&fakeOptionRepo{options: []model.Option{{OptionID: 1, ListingID: 20}}},
		&fakeFuturesRepo{},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Len(t, result, 1)

	a := result[0]
	require.Equal(t, dto.AssetTypeOption, a.Type)
	require.Equal(t, float64(200), a.Amount) // 2 contracts × 100
	require.InDelta(t, (8.0-5.0)*200, a.Profit, 0.001)
	require.Nil(t, a.OutstandingShares)
}

func TestGetPortfolio_HappyPath_Futures(t *testing.T) {
	ord := makeOrder(3, 30, model.OrderDirectionBuy, model.OrderStatusApproved, 5, 5, 200.0, 1.0)
	ord.Listing.Ticker = "CLJ22"
	ord.Listing.Price = 210.0

	svc := NewPortfolioService(
		&fakeOwnershipRepo{ownerships: []model.OrderOwnership{makeOwnership(ord)}},
		&fakeStockRepo{},
		&fakeOptionRepo{},
		&fakeFuturesRepo{futures: []model.FuturesContract{{FuturesContractID: 1, ListingID: 30}}},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Len(t, result, 1)

	a := result[0]
	require.Equal(t, dto.AssetTypeFutures, a.Type)
	require.Equal(t, float64(5), a.Amount)
	require.InDelta(t, (210.0-200.0)*5, a.Profit, 0.001)
}

func TestGetPortfolio_SkipsRejectedAndPending(t *testing.T) {
	rejected := makeOrder(1, 10, model.OrderDirectionBuy, model.OrderStatusRejected, 10, 10, 100.0, 1.0)
	pending := makeOrder(2, 10, model.OrderDirectionBuy, model.OrderStatusPendingApproval, 10, 10, 100.0, 1.0)

	svc := NewPortfolioService(
		&fakeOwnershipRepo{ownerships: []model.OrderOwnership{makeOwnership(rejected), makeOwnership(pending)}},
		&fakeStockRepo{stocks: []model.Stock{{StockID: 1, ListingID: 10}}},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
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
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	// bought 10, sold 10 → net 0 → not in portfolio
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
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, float64(6), result[0].Amount) // 10 - 4
}

func TestGetPortfolio_ForexExcluded(t *testing.T) {
	// listing 40 not in stocks/options/futures → should be excluded
	ord := makeOrder(1, 40, model.OrderDirectionBuy, model.OrderStatusApproved, 5, 5, 1.2, 1000.0)
	ord.Listing.Ticker = "EUR/USD"
	ord.Listing.Price = 1.25

	svc := NewPortfolioService(
		&fakeOwnershipRepo{ownerships: []model.OrderOwnership{makeOwnership(ord)}},
		&fakeStockRepo{},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestGetPortfolio_EmptyOwnerships(t *testing.T) {
	svc := NewPortfolioService(
		&fakeOwnershipRepo{ownerships: []model.OrderOwnership{}},
		&fakeStockRepo{},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestGetPortfolio_RepoError(t *testing.T) {
	svc := NewPortfolioService(
		&fakeOwnershipRepo{err: errTest},
		&fakeStockRepo{},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
	)

	_, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.Error(t, err)
}

func TestGetPortfolio_WeightedAvgBuyPrice(t *testing.T) {
	// Two buy orders at different prices → avg should be weighted
	buy1 := makeOrder(1, 10, model.OrderDirectionBuy, model.OrderStatusApproved, 10, 10, 100.0, 1.0)
	buy1.Listing.Ticker = "AAPL"
	buy1.Listing.Price = 150.0
	buy2 := makeOrder(2, 10, model.OrderDirectionBuy, model.OrderStatusApproved, 10, 10, 200.0, 1.0)
	buy2.Listing = buy1.Listing

	svc := NewPortfolioService(
		&fakeOwnershipRepo{ownerships: []model.OrderOwnership{makeOwnership(buy1), makeOwnership(buy2)}},
		&fakeStockRepo{stocks: []model.Stock{{StockID: 1, ListingID: 10}}},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Len(t, result, 1)

	// avgBuyPrice = (100*10 + 200*10) / 20 = 150; currentPrice = 150 → profit = 0
	require.InDelta(t, 0.0, result[0].Profit, 0.001)
	require.Equal(t, float64(20), result[0].Amount)
}
