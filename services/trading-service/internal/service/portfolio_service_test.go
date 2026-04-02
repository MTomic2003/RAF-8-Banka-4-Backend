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

type fakeOwnershipRepo struct {
	ownerships []model.OrderOwnership
	err        error
}

func (r *fakeOwnershipRepo) FindByIdentity(_ context.Context, _ uint, _ model.OwnerType) ([]model.OrderOwnership, error) {
	return r.ownerships, r.err
}

type fakeStockRepo struct {
	stocks []model.Stock
	err    error
}

func (r *fakeStockRepo) Upsert(_ context.Context, _ *model.Stock) error { return nil }

func (r *fakeStockRepo) FindAll(_ context.Context) ([]model.Stock, error) { return nil, nil }

func (r *fakeStockRepo) FindByListingIDs(_ context.Context, _ []uint) ([]model.Stock, error) {
	return r.stocks, r.err
}

func (r *fakeStockRepo) Count(_ context.Context) (int64, error) {
	return 0, nil
}

type fakeOptionRepo struct {
	options []model.Option
	err     error
}

func (r *fakeOptionRepo) Upsert(_ context.Context, _ *model.Option) error { return nil }

func (r *fakeOptionRepo) FindByListingIDs(_ context.Context, _ []uint) ([]model.Option, error) {
	return r.options, r.err
}

type fakeFuturesRepo struct {
	futures []model.FuturesContract
	err     error
}

func (r *fakeFuturesRepo) FindByListingIDs(_ context.Context, _ []uint) ([]model.FuturesContract, error) {
	return r.futures, r.err
}

type fakeForexRepo struct {
	forex []model.ForexPair
	err   error
}

func (r *fakeForexRepo) FindByListingIDs(_ context.Context, _ []uint) ([]model.ForexPair, error) {
	return r.forex, r.err
}

func (r *fakeForexRepo) Count(_ context.Context) (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return int64(len(r.forex)), nil
}

func (r *fakeForexRepo) Upsert(_ context.Context, pair model.ForexPair) error {
	if r.err != nil {
		return r.err
	}
	// dodajemo ili zamenjujemo pair u slice za jednostavan fake
	for i, f := range r.forex {
		if f.ListingID == pair.ListingID {
			r.forex[i] = pair
			return nil
		}
	}
	r.forex = append(r.forex, pair)
	return nil
}

func (r *fakeForexRepo) FindAll(_ context.Context, _ repository.ListingFilter) ([]model.ForexPair, int64, error) {
	return r.forex, int64(len(r.forex)), r.err
}

func (m *fakeOptionRepo) FindByStockID(ctx context.Context, stockID uint) ([]model.Option, error) {
	return nil, nil
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
	expectedProfit := (150.0 - 100.0) * 10
	require.InDelta(t, expectedProfit*0.15, a.TaxAmount, 0.001)
	require.NotNil(t, a.OutstandingShares)
	require.Equal(t, float64(1_000_000), *a.OutstandingShares)
}

func TestGetPortfolio_HappyPath_Option(t *testing.T) {
	ord := makeOrder(2, 20, model.OrderDirectionBuy, model.OrderStatusApproved, 2, 2, 5.0, 100.0)
	ord.Listing.Ticker = "MSFT220404C00180000"
	ord.Listing.Price = 8.0

	svc := NewPortfolioService(
		&fakeOwnershipRepo{ownerships: []model.OrderOwnership{makeOwnership(ord)}},
		&fakeStockRepo{},
		&fakeOptionRepo{options: []model.Option{{OptionID: 1, ListingID: 20}}},
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
	require.InDelta(t, ((8.0-5.0)*200)*0.15, a.TaxAmount, 0.001)
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

func TestGetPortfolio_SkipsRejectedAndPending(t *testing.T) {
	rejected := makeOrder(1, 10, model.OrderDirectionBuy, model.OrderStatusDeclined, 10, 10, 100.0, 1.0)
	pending := makeOrder(2, 10, model.OrderDirectionBuy, model.OrderStatusPending, 10, 10, 100.0, 1.0)

	svc := NewPortfolioService(
		&fakeOwnershipRepo{ownerships: []model.OrderOwnership{makeOwnership(rejected), makeOwnership(pending)}},
		&fakeStockRepo{stocks: []model.Stock{{StockID: 1, ListingID: 10}}},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
		&fakeForexRepo{},
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
	expectedProfit := (150.0 - 100.0) * 6
	require.InDelta(t, expectedProfit*0.15, result[0].TaxAmount, 0.001)
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
		&fakeOwnershipRepo{ownerships: []model.OrderOwnership{}},
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
		&fakeOwnershipRepo{err: errTest},
		&fakeStockRepo{},
		&fakeOptionRepo{},
		&fakeFuturesRepo{},
		&fakeForexRepo{},
	)

	_, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.Error(t, err)
}

func TestGetPortfolio_WeightedAvgBuyPrice(t *testing.T) {
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
		&fakeForexRepo{},
	)

	result, err := svc.GetPortfolio(context.Background(), 1, model.OwnerTypeClient)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.InDelta(t, 0.0, result[0].Profit, 0.001)
	require.InDelta(t, 0.0, result[0].TaxAmount, 0.001)
	require.Equal(t, float64(20), result[0].Amount)
}
