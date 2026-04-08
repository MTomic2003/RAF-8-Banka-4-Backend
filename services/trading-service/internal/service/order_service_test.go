package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

// ── Fake Order Repository ─────────────────────────────────────────

type fakeOrderRepo struct {
	// FindAll
	orders   []model.Order
	total    int64
	findErr  error

	// FindByID
	orderByID   *model.Order
	findByIDErr error

	// Create / Save
	createErr error
	saveErr   error

	// FindReadyForExecution
	readyOrders []model.Order
	readyErr    error

	// captured
	capturedOrder *model.Order
}

func (r *fakeOrderRepo) FindAll(_ context.Context, _, _ int, _ *uint, _ *model.OrderStatus, _ *model.OrderDirection, _ *bool) ([]model.Order, int64, error) {
	return r.orders, r.total, r.findErr
}

func (r *fakeOrderRepo) FindByID(_ context.Context, _ uint) (*model.Order, error) {
	return r.orderByID, r.findByIDErr
}

func (r *fakeOrderRepo) Create(_ context.Context, order *model.Order) error {
	r.capturedOrder = order
	return r.createErr
}

func (r *fakeOrderRepo) Save(_ context.Context, order *model.Order) error {
	r.capturedOrder = order
	return r.saveErr
}

func (r *fakeOrderRepo) FindReadyForExecution(_ context.Context, _ time.Time, _ int) ([]model.Order, error) {
	return r.readyOrders, r.readyErr
}

// ── Fake Order Transaction Repository ─────────────────────────────

type fakeOrderTransactionRepo struct {
	createErr error
}

func (r *fakeOrderTransactionRepo) Create(_ context.Context, _ *model.OrderTransaction) error {
	return r.createErr
}

// ── Fake Exchange Repository ──────────────────────────────────────

type fakeExchangeRepo struct {
	exchange    *model.Exchange
	findErr     error
	exchanges   []model.Exchange
	findAllErr  error
	total       int64
	toggledExch *model.Exchange
	toggleErr   error
}

func (r *fakeExchangeRepo) FindByMicCode(_ context.Context, _ string) (*model.Exchange, error) {
	return r.exchange, r.findErr
}

func (r *fakeExchangeRepo) FindAll(_ context.Context, _, _ int) ([]model.Exchange, int64, error) {
	return r.exchanges, r.total, r.findAllErr
}

func (r *fakeExchangeRepo) ToggleTradingEnabled(_ context.Context, _ string) (*model.Exchange, error) {
	return r.toggledExch, r.toggleErr
}

// ── Fake Listing Repository ───────────────────────────────────────

type fakeListingRepo struct {
	listing        *model.Listing
	findByIDErr    error
	dailyPriceInfo *model.ListingDailyPriceInfo
	dailyPriceErr  error

	// stubs for the rest of the interface
	allListings []model.Listing
	findAllErr  error
	countVal    int64
	countErr    error
}

func (r *fakeListingRepo) FindByID(_ context.Context, _ uint) (*model.Listing, error) {
	return r.listing, r.findByIDErr
}

func (r *fakeListingRepo) FindLatestDailyPriceInfo(_ context.Context, _ uint) (*model.ListingDailyPriceInfo, error) {
	return r.dailyPriceInfo, r.dailyPriceErr
}

func (r *fakeListingRepo) FindAll(_ context.Context) ([]model.Listing, error) {
	return r.allListings, r.findAllErr
}

func (r *fakeListingRepo) FindStocks(_ context.Context, _ repository.ListingFilter) ([]model.Listing, int64, error) {
	return nil, 0, nil
}

func (r *fakeListingRepo) FindFutures(_ context.Context, _ repository.ListingFilter) ([]model.Listing, int64, error) {
	return nil, 0, nil
}

func (r *fakeListingRepo) FindOptions(_ context.Context, _ repository.ListingFilter) ([]model.Listing, int64, error) {
	return nil, 0, nil
}

func (r *fakeListingRepo) Upsert(_ context.Context, _ *model.Listing) error { return nil }

func (r *fakeListingRepo) UpdatePriceAndAsk(_ context.Context, _ *model.Listing, _, _ float64) error {
	return nil
}

func (r *fakeListingRepo) Count(_ context.Context) (int64, error) {
	return r.countVal, r.countErr
}

func (r *fakeListingRepo) CreateDailyPriceInfo(_ context.Context, _ *model.ListingDailyPriceInfo) error {
	return nil
}

func (r *fakeListingRepo) FindLastDailyPriceInfo(_ context.Context, _ uint, _ time.Time) (*model.ListingDailyPriceInfo, error) {
	return nil, nil
}

func (r *fakeListingRepo) FindByAssetType(_ context.Context, _ model.AssetType) ([]model.Listing, error) {
	return nil, nil
}

func (r *fakeListingRepo) FindByAssetIDs(_ context.Context, _ []uint) ([]model.Listing, error) {
	return nil, nil
}


// ── Fake User Service Client ──────────────────────────────────────

type fakeUserServiceClient struct {
	employeeResp *pb.GetEmployeeByIdResponse
	employeeErr  error
	clientResp   *pb.GetClientByIdResponse
	clientErr    error
}

func (c *fakeUserServiceClient) GetEmployeeById(_ context.Context, _ uint64) (*pb.GetEmployeeByIdResponse, error) {
	return c.employeeResp, c.employeeErr
}

func (c *fakeUserServiceClient) GetClientById(_ context.Context, _ uint64) (*pb.GetClientByIdResponse, error) {
	return c.clientResp, c.clientErr
}

func (c *fakeUserServiceClient) GetAllClients(_ context.Context, _, _ int32, _, _ string) (*pb.GetAllClientsResponse, error) {
	return nil, nil
}

func (c *fakeUserServiceClient) GetAllActuaries(_ context.Context, _, _ int32, _, _ string) (*pb.GetAllActuariesResponse, error) {
	return nil, nil
}

// ── Fake Banking Client (order-specific) ──────────────────────────

type fakeOrderBankingClient struct {
	accountResp    *pb.GetAccountByNumberResponse
	accountErr     error
	settlementResp *pb.ExecuteTradeSettlementResponse
	settlementErr  error
}

func (c *fakeOrderBankingClient) GetAccountByNumber(_ context.Context, _ string) (*pb.GetAccountByNumberResponse, error) {
	return c.accountResp, c.accountErr
}

func (c *fakeOrderBankingClient) CreatePaymentWithoutVerification(_ context.Context, _ *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	return nil, nil
}

func (c *fakeOrderBankingClient) GetAccountsByClientID(_ context.Context, _ uint64) (*pb.GetAccountsByClientIDResponse, error) {
	return nil, nil
}

func (c *fakeOrderBankingClient) ConvertCurrency(_ context.Context, amount float64, _, _ string) (float64, error) {
	return amount, nil
}

func (c *fakeOrderBankingClient) ExecuteTradeSettlement(_ context.Context, _, _ string, _ pb.TradeSettlementDirection, _ float64) (*pb.ExecuteTradeSettlementResponse, error) {
	return c.settlementResp, c.settlementErr
}

// ── Helpers ───────────────────────────────────────────────────────

func clientAuthCtx() context.Context {
	clientID := uint(10)
	return auth.SetAuthOnContext(context.Background(), &auth.AuthContext{
		IdentityID:   1,
		IdentityType: auth.IdentityClient,
		ClientID:     &clientID,
	})
}

func employeeAuthCtx(employeeID uint) context.Context {
	return auth.SetAuthOnContext(context.Background(), &auth.AuthContext{
		IdentityID:   100,
		IdentityType: auth.IdentityEmployee,
		EmployeeID:   &employeeID,
	})
}

func supervisorAuthCtx(employeeID uint) context.Context {
	return auth.SetAuthOnContext(context.Background(), &auth.AuthContext{
		IdentityID:   200,
		IdentityType: auth.IdentityEmployee,
		EmployeeID:   &employeeID,
	})
}

func newTestOrderService(
	orderRepo *fakeOrderRepo,
	orderTxRepo *fakeOrderTransactionRepo,
	exchangeRepo *fakeExchangeRepo,
	listingRepo *fakeListingRepo,
	userClient *fakeUserServiceClient,
	bankingClient *fakeOrderBankingClient,
) *OrderService {
	svc := NewOrderService(orderRepo, orderTxRepo, exchangeRepo, listingRepo, &fakeAssetOwnershipRepo{}, &fakeFuturesRepo{}, &fakeOptionRepo{}, userClient, bankingClient)
	svc.now = func() time.Time {
		// Wednesday 10:00 UTC  (market hours for a UTC-0 exchange open 09:00-16:00)
		return time.Date(2025, 6, 4, 10, 0, 0, 0, time.UTC)
	}
	return svc
}

func defaultExchange() *model.Exchange {
	return &model.Exchange{
		ExchangeID:     1,
		Name:           "Test Exchange",
		Acronym:        "TST",
		MicCode:        "XTST",
		Polity:         "USA",
		Currency:       "USD",
		TimeZone:       0,
		OpenTime:       "09:00",
		CloseTime:      "16:00",
		TradingEnabled: true,
	}
}

func defaultListing() *model.Listing {
	return &model.Listing{
		ListingID:   1,
		ExchangeMIC: "XTST",
		Price:       150.0,
		Ask:         151.0,
		Asset: &model.Asset{
			Ticker:    "AAPL",
			Name:      "Apple Inc",
			AssetType: model.AssetTypeStock,
		},
	}
}

func defaultAccountResp(clientID uint64) *pb.GetAccountByNumberResponse {
	return &pb.GetAccountByNumberResponse{
		ClientId:    clientID,
		AccountType: "Current",
	}
}

// ── GetOrders Tests ───────────────────────────────────────────────

func TestGetOrders_Success(t *testing.T) {
	orders := []model.Order{
		{OrderID: 1, UserID: 1, Status: model.OrderStatusApproved},
		{OrderID: 2, UserID: 2, Status: model.OrderStatusPending},
	}
	repo := &fakeOrderRepo{orders: orders, total: 2}
	svc := newTestOrderService(repo, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	result, total, err := svc.GetOrders(context.Background(), dto.ListOrdersQuery{Page: 0, PageSize: 10})
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, int64(2), total)
}

func TestGetOrders_Empty(t *testing.T) {
	repo := &fakeOrderRepo{orders: []model.Order{}, total: 0}
	svc := newTestOrderService(repo, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	result, total, err := svc.GetOrders(context.Background(), dto.ListOrdersQuery{Page: 0, PageSize: 10})
	require.NoError(t, err)
	require.Empty(t, result)
	require.Equal(t, int64(0), total)
}

func TestGetOrders_RepoError(t *testing.T) {
	repo := &fakeOrderRepo{findErr: errors.New("db error")}
	svc := newTestOrderService(repo, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	result, total, err := svc.GetOrders(context.Background(), dto.ListOrdersQuery{})
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, int64(0), total)
}

// ── CreateOrder Tests ─────────────────────────────────────────────

func TestCreateOrder_MarketBuy_ClientAutoApproved(t *testing.T) {
	listing := defaultListing()
	exchange := defaultExchange()

	svc := newTestOrderService(
		&fakeOrderRepo{},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{exchange: exchange},
		&fakeListingRepo{listing: listing},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{accountResp: defaultAccountResp(10)},
	)

	ctx := clientAuthCtx()
	req := dto.CreateOrderRequest{
		ListingID:     1,
		AccountNumber: "444000100000000110",
		OrderType:     model.OrderTypeMarket,
		Direction:     model.OrderDirectionBuy,
		Quantity:      10,
	}

	order, err := svc.CreateOrder(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, model.OrderStatusApproved, order.Status)
	require.Equal(t, model.OrderTypeMarket, order.OrderType)
	require.Equal(t, model.OrderDirectionBuy, order.Direction)
	require.Equal(t, uint(10), order.Quantity)
	require.True(t, order.Triggered)
	require.True(t, order.CommissionExempt == false) // client is not exempt
}

func TestCreateOrder_LimitSell_Success(t *testing.T) {
	listing := defaultListing()
	exchange := defaultExchange()
	limitVal := 155.0

	svc := newTestOrderService(
		&fakeOrderRepo{},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{exchange: exchange},
		&fakeListingRepo{listing: listing},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{accountResp: defaultAccountResp(10)},
	)

	ctx := clientAuthCtx()
	req := dto.CreateOrderRequest{
		ListingID:     1,
		AccountNumber: "444000100000000110",
		OrderType:     model.OrderTypeLimit,
		Direction:     model.OrderDirectionSell,
		Quantity:      5,
		LimitValue:    &limitVal,
	}

	order, err := svc.CreateOrder(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, model.OrderTypeLimit, order.OrderType)
	require.Equal(t, model.OrderDirectionSell, order.Direction)
	require.NotNil(t, order.LimitValue)
	require.Equal(t, 155.0, *order.LimitValue)
}

func TestCreateOrder_MissingAuthContext(t *testing.T) {
	svc := newTestOrderService(
		&fakeOrderRepo{},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	req := dto.CreateOrderRequest{
		ListingID:     1,
		AccountNumber: "444000100000000110",
		OrderType:     model.OrderTypeMarket,
		Direction:     model.OrderDirectionBuy,
		Quantity:      10,
	}

	order, err := svc.CreateOrder(context.Background(), req)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "not authenticated")
}

func TestCreateOrder_ListingNotFound(t *testing.T) {
	svc := newTestOrderService(
		&fakeOrderRepo{},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{listing: nil},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{accountResp: defaultAccountResp(10)},
	)

	ctx := clientAuthCtx()
	req := dto.CreateOrderRequest{
		ListingID:     999,
		AccountNumber: "444000100000000110",
		OrderType:     model.OrderTypeMarket,
		Direction:     model.OrderDirectionBuy,
		Quantity:      10,
	}

	order, err := svc.CreateOrder(ctx, req)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "listing not found")
}

func TestCreateOrder_AccountValidationFailure_NotFound(t *testing.T) {
	svc := newTestOrderService(
		&fakeOrderRepo{},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{accountErr: status.Error(codes.NotFound, "account not found")},
	)

	ctx := clientAuthCtx()
	req := dto.CreateOrderRequest{
		ListingID:     1,
		AccountNumber: "444000100000000999",
		OrderType:     model.OrderTypeMarket,
		Direction:     model.OrderDirectionBuy,
		Quantity:      10,
	}

	order, err := svc.CreateOrder(ctx, req)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "account not found")
}

func TestCreateOrder_AccountDoesNotBelongToClient(t *testing.T) {
	svc := newTestOrderService(
		&fakeOrderRepo{},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{accountResp: &pb.GetAccountByNumberResponse{ClientId: 999, AccountType: "Current"}},
	)

	ctx := clientAuthCtx() // clientID=10
	req := dto.CreateOrderRequest{
		ListingID:     1,
		AccountNumber: "444000100000000110",
		OrderType:     model.OrderTypeMarket,
		Direction:     model.OrderDirectionBuy,
		Quantity:      10,
	}

	order, err := svc.CreateOrder(ctx, req)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "account does not belong to you")
}

func TestCreateOrder_EmployeeAgent_SufficientLimit_Approved(t *testing.T) {
	listing := defaultListing()
	exchange := defaultExchange()

	svc := newTestOrderService(
		&fakeOrderRepo{},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{exchange: exchange},
		&fakeListingRepo{listing: listing},
		&fakeUserServiceClient{
			employeeResp: &pb.GetEmployeeByIdResponse{
				Id:           5,
				IsAgent:      true,
				NeedApproval: false,
				OrderLimit:   1000000,
				UsedLimit:    0,
			},
		},
		&fakeOrderBankingClient{accountResp: &pb.GetAccountByNumberResponse{AccountType: "Bank"}},
	)

	ctx := employeeAuthCtx(5)
	req := dto.CreateOrderRequest{
		ListingID:     1,
		AccountNumber: "444000100000000110",
		OrderType:     model.OrderTypeMarket,
		Direction:     model.OrderDirectionBuy,
		Quantity:      10,
	}

	order, err := svc.CreateOrder(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, model.OrderStatusApproved, order.Status)
	require.True(t, order.CommissionExempt)
}

func TestCreateOrder_EmployeeAgent_ExceedsLimit_Pending(t *testing.T) {
	listing := defaultListing()
	exchange := defaultExchange()

	svc := newTestOrderService(
		&fakeOrderRepo{},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{exchange: exchange},
		&fakeListingRepo{listing: listing},
		&fakeUserServiceClient{
			employeeResp: &pb.GetEmployeeByIdResponse{
				Id:           5,
				IsAgent:      true,
				NeedApproval: false,
				OrderLimit:   100, // very low limit
				UsedLimit:    99,
			},
		},
		&fakeOrderBankingClient{accountResp: &pb.GetAccountByNumberResponse{AccountType: "Bank"}},
	)

	ctx := employeeAuthCtx(5)
	req := dto.CreateOrderRequest{
		ListingID:     1,
		AccountNumber: "444000100000000110",
		OrderType:     model.OrderTypeMarket,
		Direction:     model.OrderDirectionBuy,
		Quantity:      10,
	}

	order, err := svc.CreateOrder(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, model.OrderStatusPending, order.Status)
}

func TestCreateOrder_EmployeeNeedsApproval_Pending(t *testing.T) {
	listing := defaultListing()
	exchange := defaultExchange()

	svc := newTestOrderService(
		&fakeOrderRepo{},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{exchange: exchange},
		&fakeListingRepo{listing: listing},
		&fakeUserServiceClient{
			employeeResp: &pb.GetEmployeeByIdResponse{
				Id:           5,
				IsAgent:      true,
				NeedApproval: true,
				OrderLimit:   1000000,
				UsedLimit:    0,
			},
		},
		&fakeOrderBankingClient{accountResp: &pb.GetAccountByNumberResponse{AccountType: "Bank"}},
	)

	ctx := employeeAuthCtx(5)
	req := dto.CreateOrderRequest{
		ListingID:     1,
		AccountNumber: "444000100000000110",
		OrderType:     model.OrderTypeMarket,
		Direction:     model.OrderDirectionBuy,
		Quantity:      10,
	}

	order, err := svc.CreateOrder(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, model.OrderStatusPending, order.Status)
}

func TestCreateOrder_EmployeeNotAgent_Pending(t *testing.T) {
	listing := defaultListing()
	exchange := defaultExchange()

	svc := newTestOrderService(
		&fakeOrderRepo{},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{exchange: exchange},
		&fakeListingRepo{listing: listing},
		&fakeUserServiceClient{
			employeeResp: &pb.GetEmployeeByIdResponse{
				Id:      5,
				IsAgent: false,
			},
		},
		&fakeOrderBankingClient{accountResp: &pb.GetAccountByNumberResponse{AccountType: "Bank"}},
	)

	ctx := employeeAuthCtx(5)
	req := dto.CreateOrderRequest{
		ListingID:     1,
		AccountNumber: "444000100000000110",
		OrderType:     model.OrderTypeMarket,
		Direction:     model.OrderDirectionBuy,
		Quantity:      10,
	}

	order, err := svc.CreateOrder(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, model.OrderStatusPending, order.Status)
}

func TestCreateOrder_LimitOrder_MissingLimitValue(t *testing.T) {
	svc := newTestOrderService(
		&fakeOrderRepo{},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	ctx := clientAuthCtx()
	req := dto.CreateOrderRequest{
		ListingID:     1,
		AccountNumber: "444000100000000110",
		OrderType:     model.OrderTypeLimit,
		Direction:     model.OrderDirectionBuy,
		Quantity:      10,
		LimitValue:    nil,
	}

	order, err := svc.CreateOrder(ctx, req)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "limitValue is required")
}

func TestCreateOrder_StopOrder_MissingStopValue(t *testing.T) {
	svc := newTestOrderService(
		&fakeOrderRepo{},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	ctx := clientAuthCtx()
	req := dto.CreateOrderRequest{
		ListingID:     1,
		AccountNumber: "444000100000000110",
		OrderType:     model.OrderTypeStop,
		Direction:     model.OrderDirectionBuy,
		Quantity:      10,
		StopValue:     nil,
	}

	order, err := svc.CreateOrder(ctx, req)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "stopValue is required")
}

func TestCreateOrder_EmployeeMustUseBankAccount(t *testing.T) {
	svc := newTestOrderService(
		&fakeOrderRepo{},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{accountResp: &pb.GetAccountByNumberResponse{AccountType: "Current"}},
	)

	ctx := employeeAuthCtx(5)
	req := dto.CreateOrderRequest{
		ListingID:     1,
		AccountNumber: "444000100000000110",
		OrderType:     model.OrderTypeMarket,
		Direction:     model.OrderDirectionBuy,
		Quantity:      10,
	}

	order, err := svc.CreateOrder(ctx, req)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "employees must use a bank account")
}

func TestCreateOrder_RepoCreateError(t *testing.T) {
	listing := defaultListing()
	exchange := defaultExchange()

	svc := newTestOrderService(
		&fakeOrderRepo{createErr: errors.New("db insert error")},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{exchange: exchange},
		&fakeListingRepo{listing: listing},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{accountResp: defaultAccountResp(10)},
	)

	ctx := clientAuthCtx()
	req := dto.CreateOrderRequest{
		ListingID:     1,
		AccountNumber: "444000100000000110",
		OrderType:     model.OrderTypeMarket,
		Direction:     model.OrderDirectionBuy,
		Quantity:      10,
	}

	order, err := svc.CreateOrder(ctx, req)
	require.Error(t, err)
	require.Nil(t, order)
}

// ── ApproveOrder Tests ────────────────────────────────────────────

func TestApproveOrder_Success(t *testing.T) {
	pendingOrder := &model.Order{
		OrderID: 1,
		Status:  model.OrderStatusPending,
		Listing: model.Listing{ExchangeMIC: "XTST"},
	}

	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: pendingOrder},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{exchange: defaultExchange()},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	ctx := employeeAuthCtx(5)
	order, err := svc.ApproveOrder(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, model.OrderStatusApproved, order.Status)
	require.NotNil(t, order.ApprovedBy)
	require.Equal(t, uint(100), *order.ApprovedBy) // IdentityID from employeeAuthCtx
	require.NotNil(t, order.NextExecutionAt)
}

func TestApproveOrder_NotFound(t *testing.T) {
	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: nil},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	ctx := employeeAuthCtx(5)
	order, err := svc.ApproveOrder(ctx, 999)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "order not found")
}

func TestApproveOrder_NotPending(t *testing.T) {
	approvedOrder := &model.Order{
		OrderID: 1,
		Status:  model.OrderStatusApproved,
	}

	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: approvedOrder},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	ctx := employeeAuthCtx(5)
	order, err := svc.ApproveOrder(ctx, 1)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "only pending orders can be approved")
}

func TestApproveOrder_MissingAuth(t *testing.T) {
	pendingOrder := &model.Order{
		OrderID: 1,
		Status:  model.OrderStatusPending,
	}

	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: pendingOrder},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	order, err := svc.ApproveOrder(context.Background(), 1)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "not authenticated")
}

// ── DeclineOrder Tests ────────────────────────────────────────────

func TestDeclineOrder_Success(t *testing.T) {
	pendingOrder := &model.Order{
		OrderID: 1,
		Status:  model.OrderStatusPending,
	}

	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: pendingOrder},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	ctx := employeeAuthCtx(5)
	order, err := svc.DeclineOrder(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, model.OrderStatusDeclined, order.Status)
	require.True(t, order.IsDone)
	require.Nil(t, order.NextExecutionAt)
	require.NotNil(t, order.ApprovedBy)
}

func TestDeclineOrder_NotFound(t *testing.T) {
	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: nil},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	ctx := employeeAuthCtx(5)
	order, err := svc.DeclineOrder(ctx, 999)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "order not found")
}

func TestDeclineOrder_NotPending(t *testing.T) {
	approvedOrder := &model.Order{
		OrderID: 1,
		Status:  model.OrderStatusApproved,
	}

	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: approvedOrder},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	ctx := employeeAuthCtx(5)
	order, err := svc.DeclineOrder(ctx, 1)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "only pending orders can be declined")
}

func TestDeclineOrder_MissingAuth(t *testing.T) {
	pendingOrder := &model.Order{
		OrderID: 1,
		Status:  model.OrderStatusPending,
	}

	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: pendingOrder},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	order, err := svc.DeclineOrder(context.Background(), 1)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "not authenticated")
}

// ── CancelOrder Tests ─────────────────────────────────────────────

func TestCancelOrder_OwnerCancelsOwn(t *testing.T) {
	pendingOrder := &model.Order{
		OrderID: 1,
		UserID:  1, // matches clientAuthCtx IdentityID
		Status:  model.OrderStatusPending,
		IsDone:  false,
	}

	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: pendingOrder},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	ctx := clientAuthCtx()
	order, err := svc.CancelOrder(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, model.OrderStatusDeclined, order.Status)
	require.True(t, order.IsDone)
	require.Nil(t, order.NextExecutionAt)
}

func TestCancelOrder_SupervisorCancelsOther(t *testing.T) {
	pendingOrder := &model.Order{
		OrderID: 1,
		UserID:  999, // different from supervisor's IdentityID
		Status:  model.OrderStatusApproved,
		IsDone:  false,
	}

	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: pendingOrder},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{
			employeeResp: &pb.GetEmployeeByIdResponse{
				Id:           7,
				IsSupervisor: true,
			},
		},
		&fakeOrderBankingClient{},
	)

	ctx := supervisorAuthCtx(7)
	order, err := svc.CancelOrder(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, order)
	require.Equal(t, model.OrderStatusDeclined, order.Status)
	require.True(t, order.IsDone)
}

func TestCancelOrder_NonOwnerNonSupervisor_Forbidden(t *testing.T) {
	pendingOrder := &model.Order{
		OrderID: 1,
		UserID:  999, // different from employee's IdentityID (100)
		Status:  model.OrderStatusPending,
		IsDone:  false,
	}

	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: pendingOrder},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{
			employeeResp: &pb.GetEmployeeByIdResponse{
				Id:           5,
				IsSupervisor: false,
			},
		},
		&fakeOrderBankingClient{},
	)

	ctx := employeeAuthCtx(5)
	order, err := svc.CancelOrder(ctx, 1)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "only the order owner or a supervisor")
}

func TestCancelOrder_AlreadyDone(t *testing.T) {
	doneOrder := &model.Order{
		OrderID: 1,
		UserID:  1,
		Status:  model.OrderStatusApproved,
		IsDone:  true,
	}

	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: doneOrder},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	ctx := clientAuthCtx()
	order, err := svc.CancelOrder(ctx, 1)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "cannot cancel a completed order")
}

func TestCancelOrder_NotFound(t *testing.T) {
	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: nil},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	ctx := clientAuthCtx()
	order, err := svc.CancelOrder(ctx, 999)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "order not found")
}

func TestCancelOrder_DeclinedStatus_BadRequest(t *testing.T) {
	declinedOrder := &model.Order{
		OrderID: 1,
		UserID:  1,
		Status:  model.OrderStatusDeclined,
		IsDone:  false,
	}

	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: declinedOrder},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	ctx := clientAuthCtx()
	order, err := svc.CancelOrder(ctx, 1)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "only pending or approved orders can be cancelled")
}

func TestCancelOrder_MissingAuth(t *testing.T) {
	pendingOrder := &model.Order{
		OrderID: 1,
		UserID:  1,
		Status:  model.OrderStatusPending,
		IsDone:  false,
	}

	svc := newTestOrderService(
		&fakeOrderRepo{orderByID: pendingOrder},
		&fakeOrderTransactionRepo{},
		&fakeExchangeRepo{},
		&fakeListingRepo{},
		&fakeUserServiceClient{},
		&fakeOrderBankingClient{},
	)

	order, err := svc.CancelOrder(context.Background(), 1)
	require.Error(t, err)
	require.Nil(t, order)
	require.Contains(t, err.Error(), "not authenticated")
}

// ── Pure function tests ───────────────────────────────────────────

func TestValidateOrderTypeFields(t *testing.T) {
	// Market order needs nothing extra
	require.NoError(t, validateOrderTypeFields(dto.CreateOrderRequest{OrderType: model.OrderTypeMarket}))

	// Limit order requires LimitValue
	require.Error(t, validateOrderTypeFields(dto.CreateOrderRequest{OrderType: model.OrderTypeLimit}))
	lv := 100.0
	require.NoError(t, validateOrderTypeFields(dto.CreateOrderRequest{OrderType: model.OrderTypeLimit, LimitValue: &lv}))

	// Stop order requires StopValue
	require.Error(t, validateOrderTypeFields(dto.CreateOrderRequest{OrderType: model.OrderTypeStop}))
	sv := 90.0
	require.NoError(t, validateOrderTypeFields(dto.CreateOrderRequest{OrderType: model.OrderTypeStop, StopValue: &sv}))

	// StopLimit requires both
	require.Error(t, validateOrderTypeFields(dto.CreateOrderRequest{OrderType: model.OrderTypeStopLimit, LimitValue: &lv}))
	require.Error(t, validateOrderTypeFields(dto.CreateOrderRequest{OrderType: model.OrderTypeStopLimit, StopValue: &sv}))
	require.NoError(t, validateOrderTypeFields(dto.CreateOrderRequest{OrderType: model.OrderTypeStopLimit, LimitValue: &lv, StopValue: &sv}))
}

func TestCalculateInitialPricePerUnit(t *testing.T) {
	listing := &model.Listing{Price: 150.0, Ask: 151.0}
	lv := 155.0
	sv := 145.0

	// Market buy -> Ask
	p := calculateInitialPricePerUnit(dto.CreateOrderRequest{OrderType: model.OrderTypeMarket, Direction: model.OrderDirectionBuy}, listing)
	require.NotNil(t, p)
	require.Equal(t, 151.0, *p)

	// Market sell -> Price
	p = calculateInitialPricePerUnit(dto.CreateOrderRequest{OrderType: model.OrderTypeMarket, Direction: model.OrderDirectionSell}, listing)
	require.NotNil(t, p)
	require.Equal(t, 150.0, *p)

	// Limit -> LimitValue
	p = calculateInitialPricePerUnit(dto.CreateOrderRequest{OrderType: model.OrderTypeLimit, LimitValue: &lv}, listing)
	require.NotNil(t, p)
	require.Equal(t, 155.0, *p)

	// Stop -> StopValue
	p = calculateInitialPricePerUnit(dto.CreateOrderRequest{OrderType: model.OrderTypeStop, StopValue: &sv}, listing)
	require.NotNil(t, p)
	require.Equal(t, 145.0, *p)
}

func TestCalculateCommission(t *testing.T) {
	// Market order: min(0.14 * value, 7)
	require.Equal(t, 0.0, calculateCommission(model.OrderTypeMarket, 0))
	require.Equal(t, 0.0, calculateCommission(model.OrderTypeMarket, -10))
	require.InDelta(t, 0.14*10, calculateCommission(model.OrderTypeMarket, 10), 0.001)
	require.Equal(t, 7.0, calculateCommission(model.OrderTypeMarket, 100)) // 0.14*100=14, capped at 7

	// Limit order: min(0.24 * value, 12)
	require.InDelta(t, 0.24*10, calculateCommission(model.OrderTypeLimit, 10), 0.001)
	require.Equal(t, 12.0, calculateCommission(model.OrderTypeLimit, 100)) // 0.24*100=24, capped at 12
}

func TestNormalizeCurrencyCode(t *testing.T) {
	require.Equal(t, "USD", normalizeCurrencyCode("usd"))
	require.Equal(t, "EUR", normalizeCurrencyCode(" eur "))
	require.Equal(t, "RSD", normalizeCurrencyCode("RSD"))
}

func TestDereferencePrice(t *testing.T) {
	require.Equal(t, 0.0, dereferencePrice(nil))
	v := 42.5
	require.Equal(t, 42.5, dereferencePrice(&v))
}

func TestIsStopConditionMet(t *testing.T) {
	listing := &model.Listing{Price: 150.0, Ask: 151.0}

	// No stop value -> always met
	order := &model.Order{StopValue: nil}
	require.True(t, isStopConditionMet(order, listing))

	// Buy: Ask >= StopValue
	sv := 150.0
	order = &model.Order{Direction: model.OrderDirectionBuy, StopValue: &sv}
	require.True(t, isStopConditionMet(order, listing))

	sv = 200.0
	order = &model.Order{Direction: model.OrderDirectionBuy, StopValue: &sv}
	require.False(t, isStopConditionMet(order, listing))

	// Sell: Price <= StopValue
	sv = 150.0
	order = &model.Order{Direction: model.OrderDirectionSell, StopValue: &sv}
	require.True(t, isStopConditionMet(order, listing))

	sv = 100.0
	order = &model.Order{Direction: model.OrderDirectionSell, StopValue: &sv}
	require.False(t, isStopConditionMet(order, listing))
}

func TestResolveExecutionPrice(t *testing.T) {
	listing := &model.Listing{Price: 150.0, Ask: 151.0}

	// Market buy -> Ask
	p, ok := resolveExecutionPrice(&model.Order{OrderType: model.OrderTypeMarket, Direction: model.OrderDirectionBuy}, listing)
	require.True(t, ok)
	require.Equal(t, 151.0, p)

	// Market sell -> Price
	p, ok = resolveExecutionPrice(&model.Order{OrderType: model.OrderTypeMarket, Direction: model.OrderDirectionSell}, listing)
	require.True(t, ok)
	require.Equal(t, 150.0, p)

	// Limit buy, Ask <= LimitValue -> Ask
	lv := 155.0
	p, ok = resolveExecutionPrice(&model.Order{OrderType: model.OrderTypeLimit, Direction: model.OrderDirectionBuy, LimitValue: &lv}, listing)
	require.True(t, ok)
	require.Equal(t, 151.0, p)

	// Limit buy, Ask > LimitValue -> can't execute
	lv = 140.0
	_, ok = resolveExecutionPrice(&model.Order{OrderType: model.OrderTypeLimit, Direction: model.OrderDirectionBuy, LimitValue: &lv}, listing)
	require.False(t, ok)

	// Limit sell, Price >= LimitValue -> Price
	lv = 140.0
	p, ok = resolveExecutionPrice(&model.Order{OrderType: model.OrderTypeLimit, Direction: model.OrderDirectionSell, LimitValue: &lv}, listing)
	require.True(t, ok)
	require.Equal(t, 150.0, p)

	// Limit sell, Price < LimitValue -> can't execute
	lv = 160.0
	_, ok = resolveExecutionPrice(&model.Order{OrderType: model.OrderTypeLimit, Direction: model.OrderDirectionSell, LimitValue: &lv}, listing)
	require.False(t, ok)
}

func TestIsWeekend(t *testing.T) {
	sat := time.Date(2025, 6, 7, 12, 0, 0, 0, time.UTC) // Saturday
	sun := time.Date(2025, 6, 8, 12, 0, 0, 0, time.UTC) // Sunday
	mon := time.Date(2025, 6, 9, 12, 0, 0, 0, time.UTC) // Monday

	require.True(t, isWeekend(sat))
	require.True(t, isWeekend(sun))
	require.False(t, isWeekend(mon))
}

// ── failOrder Tests ──────────────────────────────────────────────

func TestFailOrder_SetsFieldsAndSaves(t *testing.T) {
	repo := &fakeOrderRepo{}
	svc := newTestOrderService(repo, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	nextExec := time.Date(2025, 6, 5, 12, 0, 0, 0, time.UTC)
	order := &model.Order{
		OrderID:         1,
		Status:          model.OrderStatusApproved,
		IsDone:          false,
		NextExecutionAt: &nextExec,
	}

	err := svc.failOrder(context.Background(), order, model.OrderStatusDeclined)
	require.NoError(t, err)
	require.Equal(t, model.OrderStatusDeclined, order.Status)
	require.True(t, order.IsDone)
	require.Nil(t, order.NextExecutionAt)
	require.NotNil(t, repo.capturedOrder)
}

func TestFailOrder_RepoSaveError(t *testing.T) {
	repo := &fakeOrderRepo{saveErr: errors.New("save failed")}
	svc := newTestOrderService(repo, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	order := &model.Order{OrderID: 1, Status: model.OrderStatusApproved}
	err := svc.failOrder(context.Background(), order, model.OrderStatusDeclined)
	require.Error(t, err)
}

// ── resolveFillQuantity Tests ────────────────────────────────────

func TestResolveFillQuantity_RemainingZero(t *testing.T) {
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	order := &model.Order{Quantity: 10, FilledQty: 10}
	require.Equal(t, uint(0), svc.resolveFillQuantity(order))
}

func TestResolveFillQuantity_AllOrNone(t *testing.T) {
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	order := &model.Order{Quantity: 10, FilledQty: 3, AllOrNone: true}
	require.Equal(t, uint(7), svc.resolveFillQuantity(order))
}

func TestResolveFillQuantity_RemainingOne(t *testing.T) {
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	order := &model.Order{Quantity: 5, FilledQty: 4, AllOrNone: false}
	require.Equal(t, uint(1), svc.resolveFillQuantity(order))
}

func TestResolveFillQuantity_RandomInRange(t *testing.T) {
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	order := &model.Order{Quantity: 100, FilledQty: 0, AllOrNone: false}

	// Run multiple times to verify result is always in [1, remaining]
	for i := 0; i < 50; i++ {
		qty := svc.resolveFillQuantity(order)
		require.GreaterOrEqual(t, qty, uint(1))
		require.LessOrEqual(t, qty, uint(100))
	}
}

// ── resolveExchangeSession Tests ─────────────────────────────────

func TestResolveExchangeSession_NilExchange(t *testing.T) {
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	session := svc.resolveExchangeSession(nil)
	require.True(t, session.IsOpen)
	require.False(t, session.IsClosed)
}

func TestResolveExchangeSession_TradingDisabled(t *testing.T) {
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	exchange := &model.Exchange{
		TradingEnabled: false,
		OpenTime:       "09:00",
		CloseTime:      "16:00",
		TimeZone:       0,
	}
	session := svc.resolveExchangeSession(exchange)
	require.True(t, session.IsOpen)
	require.False(t, session.IsClosed)
}

func TestResolveExchangeSession_DuringOpenHours(t *testing.T) {
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})
	// newTestOrderService sets now to Wednesday 2025-06-04 10:00 UTC
	// Exchange with timezone 0, open 09:00-16:00 -> 10:00 is during open hours

	exchange := &model.Exchange{
		TradingEnabled: true,
		OpenTime:       "09:00",
		CloseTime:      "16:00",
		TimeZone:       0,
	}
	session := svc.resolveExchangeSession(exchange)
	require.True(t, session.IsOpen)
	require.False(t, session.IsClosed)
}

func TestResolveExchangeSession_DuringClosedHours_BeforeOpen(t *testing.T) {
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})
	// now = Wednesday 2025-06-04 10:00 UTC
	// Exchange with timezone +5 -> local time = 15:00, open 16:00-23:00 -> before open
	exchange := &model.Exchange{
		TradingEnabled: true,
		OpenTime:       "16:00",
		CloseTime:      "23:00",
		TimeZone:       5,
	}
	session := svc.resolveExchangeSession(exchange)
	require.True(t, session.IsClosed)
	require.False(t, session.IsOpen)
	require.False(t, session.NextOpen.IsZero())
}

func TestResolveExchangeSession_DuringClosedHours_AfterClose(t *testing.T) {
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})
	// now = Wednesday 2025-06-04 10:00 UTC
	// Exchange with timezone 0, open 06:00-09:00 -> local time 10:00 is after close
	exchange := &model.Exchange{
		TradingEnabled: true,
		OpenTime:       "06:00",
		CloseTime:      "09:00",
		TimeZone:       0,
	}
	session := svc.resolveExchangeSession(exchange)
	require.True(t, session.IsClosed)
	require.False(t, session.IsOpen)
	require.False(t, session.NextOpen.IsZero())
}

// ── processOrder basic paths ─────────────────────────────────────

func TestProcessOrder_ListingNotFound_FailsOrder(t *testing.T) {
	orderRepo := &fakeOrderRepo{}
	listingRepo := &fakeListingRepo{listing: nil}
	svc := newTestOrderService(orderRepo, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, listingRepo, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	order := &model.Order{
		OrderID:   1,
		ListingID: 999,
		Status:    model.OrderStatusApproved,
	}

	err := svc.processOrder(context.Background(), order)
	require.NoError(t, err)
	require.True(t, order.IsDone)
	require.Equal(t, model.OrderStatusDeclined, order.Status)
	require.Nil(t, order.NextExecutionAt)
}

func TestProcessOrder_ExchangeNotFound_FailsOrder(t *testing.T) {
	listing := defaultListing()
	orderRepo := &fakeOrderRepo{}
	listingRepo := &fakeListingRepo{listing: listing}
	exchangeRepo := &fakeExchangeRepo{exchange: nil}
	svc := newTestOrderService(orderRepo, &fakeOrderTransactionRepo{}, exchangeRepo, listingRepo, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	order := &model.Order{
		OrderID:   1,
		ListingID: 1,
		Status:    model.OrderStatusApproved,
	}

	err := svc.processOrder(context.Background(), order)
	require.NoError(t, err)
	require.True(t, order.IsDone)
	require.Equal(t, model.OrderStatusDeclined, order.Status)
}

func TestProcessOrder_ListingRepoError(t *testing.T) {
	listingRepo := &fakeListingRepo{findByIDErr: errors.New("db error")}
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, listingRepo, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	order := &model.Order{OrderID: 1, ListingID: 1}
	err := svc.processOrder(context.Background(), order)
	require.Error(t, err)
}

func TestProcessOrder_ExchangeRepoError(t *testing.T) {
	listing := defaultListing()
	listingRepo := &fakeListingRepo{listing: listing}
	exchangeRepo := &fakeExchangeRepo{findErr: errors.New("db error")}
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, exchangeRepo, listingRepo, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	order := &model.Order{OrderID: 1, ListingID: 1}
	err := svc.processOrder(context.Background(), order)
	require.Error(t, err)
}

func TestProcessOrder_MarketOrder_FullFill(t *testing.T) {
	listing := defaultListing()
	exchange := defaultExchange()
	orderRepo := &fakeOrderRepo{}
	listingRepo := &fakeListingRepo{listing: listing}
	exchangeRepo := &fakeExchangeRepo{exchange: exchange}
	txRepo := &fakeOrderTransactionRepo{}
	bankingClient := &fakeOrderBankingClient{
		settlementResp: &pb.ExecuteTradeSettlementResponse{
			SourceAmount:         151.0,
			SourceCurrencyCode:   "USD",
			DestinationAmount:    151.0,
			DestinationCurrencyCode: "USD",
		},
	}
	svc := newTestOrderService(orderRepo, txRepo, exchangeRepo, listingRepo, &fakeUserServiceClient{}, bankingClient)

	order := &model.Order{
		OrderID:      1,
		ListingID:    1,
		OrderType:    model.OrderTypeMarket,
		Direction:    model.OrderDirectionBuy,
		Quantity:     1,
		FilledQty:    0,
		ContractSize: 1,
		Triggered:    true,
		AllOrNone:    true,
		Status:       model.OrderStatusApproved,
		AccountNumber: "444000100000000110",
		CommissionExempt: true,
	}

	err := svc.processOrder(context.Background(), order)
	require.NoError(t, err)
	require.Equal(t, uint(1), order.FilledQty)
	require.True(t, order.IsDone)
	require.Nil(t, order.NextExecutionAt)
}

// ── processDueOrders Tests ───────────────────────────────────────

func TestProcessDueOrders_NoReadyOrders(t *testing.T) {
	orderRepo := &fakeOrderRepo{readyOrders: []model.Order{}}
	svc := newTestOrderService(orderRepo, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	err := svc.processDueOrders(context.Background())
	require.NoError(t, err)
}

func TestProcessDueOrders_RepoError(t *testing.T) {
	orderRepo := &fakeOrderRepo{readyErr: errors.New("db error")}
	svc := newTestOrderService(orderRepo, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	err := svc.processDueOrders(context.Background())
	require.Error(t, err)
}

func TestProcessDueOrders_WithReadyOrders_ProcessesThem(t *testing.T) {
	listing := defaultListing()
	exchange := defaultExchange()

	readyOrders := []model.Order{
		{
			OrderID:      1,
			ListingID:    1,
			OrderType:    model.OrderTypeMarket,
			Direction:    model.OrderDirectionBuy,
			Quantity:     1,
			ContractSize: 1,
			Triggered:    true,
			AllOrNone:    true,
			Status:       model.OrderStatusApproved,
			AccountNumber: "444000100000000110",
			CommissionExempt: true,
		},
	}

	orderRepo := &fakeOrderRepo{readyOrders: readyOrders}
	listingRepo := &fakeListingRepo{listing: listing}
	exchangeRepo := &fakeExchangeRepo{exchange: exchange}
	txRepo := &fakeOrderTransactionRepo{}
	bankingClient := &fakeOrderBankingClient{
		settlementResp: &pb.ExecuteTradeSettlementResponse{
			SourceAmount:         151.0,
			SourceCurrencyCode:   "USD",
			DestinationAmount:    151.0,
			DestinationCurrencyCode: "USD",
		},
	}

	svc := newTestOrderService(orderRepo, txRepo, exchangeRepo, listingRepo, &fakeUserServiceClient{}, bankingClient)

	err := svc.processDueOrders(context.Background())
	require.NoError(t, err)
}

// ── Start / Stop Tests ───────────────────────────────────────────

func TestStartStop_DoesNotPanic(t *testing.T) {
	svc := newTestOrderService(&fakeOrderRepo{readyOrders: []model.Order{}}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	svc.Start()
	// calling Start again should be a no-op (already running)
	svc.Start()
	svc.Stop()
	// calling Stop again should be safe
	svc.Stop()
}

// ── initialExecutionTime Tests ───────────────────────────────────

func TestInitialExecutionTime_AfterHours(t *testing.T) {
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	session := exchangeSession{IsOpen: true}
	result := svc.initialExecutionTime(session, true)
	expected := svc.now().Add(afterHoursExecutionDelay)
	require.Equal(t, expected, result)
}

func TestInitialExecutionTime_OpenSession(t *testing.T) {
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	session := exchangeSession{IsOpen: true}
	result := svc.initialExecutionTime(session, false)
	require.Equal(t, svc.now(), result)
}

func TestInitialExecutionTime_ClosedSession(t *testing.T) {
	svc := newTestOrderService(&fakeOrderRepo{}, &fakeOrderTransactionRepo{}, &fakeExchangeRepo{}, &fakeListingRepo{}, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	nextOpen := time.Date(2025, 6, 5, 9, 0, 0, 0, time.UTC)
	session := exchangeSession{IsOpen: false, NextOpen: nextOpen}
	result := svc.initialExecutionTime(session, false)
	require.Equal(t, nextOpen, result)
}

// ── approximateOrderValue Tests ──────────────────────────────────

func TestApproximateOrderValue(t *testing.T) {
	price := 100.0
	order := &model.Order{Quantity: 10, ContractSize: 1, PricePerUnit: &price}
	require.Equal(t, 1000.0, approximateOrderValue(order, 0))

	// Fallback when PricePerUnit is nil
	order2 := &model.Order{Quantity: 5, ContractSize: 2, PricePerUnit: nil}
	require.Equal(t, 500.0, approximateOrderValue(order2, 50.0))

	// PricePerUnit is zero, uses fallback
	zero := 0.0
	order3 := &model.Order{Quantity: 5, ContractSize: 2, PricePerUnit: &zero}
	require.Equal(t, 500.0, approximateOrderValue(order3, 50.0))
}

// ── processOrder: closed exchange reschedules ────────────────────

func TestProcessOrder_ClosedExchange_Reschedules(t *testing.T) {
	listing := defaultListing()
	// Exchange with timezone +10 -> local time = 20:00 UTC+10, open 09:00-16:00 -> after close
	exchange := &model.Exchange{
		ExchangeID:     1,
		Name:           "Late Exchange",
		MicCode:        "XTST",
		Currency:       "USD",
		TimeZone:       10,
		OpenTime:       "09:00",
		CloseTime:      "16:00",
		TradingEnabled: true,
	}

	orderRepo := &fakeOrderRepo{}
	listingRepo := &fakeListingRepo{listing: listing}
	exchangeRepo := &fakeExchangeRepo{exchange: exchange}
	svc := newTestOrderService(orderRepo, &fakeOrderTransactionRepo{}, exchangeRepo, listingRepo, &fakeUserServiceClient{}, &fakeOrderBankingClient{})

	order := &model.Order{
		OrderID:      1,
		ListingID:    1,
		OrderType:    model.OrderTypeMarket,
		Direction:    model.OrderDirectionBuy,
		Quantity:     5,
		ContractSize: 1,
		Triggered:    true,
		Status:       model.OrderStatusApproved,
		AfterHours:   false,
	}

	err := svc.processOrder(context.Background(), order)
	require.NoError(t, err)
	// Order should be rescheduled, not failed
	require.False(t, order.IsDone)
	require.NotNil(t, order.NextExecutionAt)
}

// ── processOrder: settlement gRPC FailedPrecondition fails order ─

func TestProcessOrder_SettlementFailedPrecondition_FailsOrder(t *testing.T) {
	listing := defaultListing()
	exchange := defaultExchange()

	orderRepo := &fakeOrderRepo{}
	listingRepo := &fakeListingRepo{listing: listing}
	exchangeRepo := &fakeExchangeRepo{exchange: exchange}
	txRepo := &fakeOrderTransactionRepo{}
	bankingClient := &fakeOrderBankingClient{
		settlementErr: status.Error(codes.FailedPrecondition, "insufficient funds"),
	}
	svc := newTestOrderService(orderRepo, txRepo, exchangeRepo, listingRepo, &fakeUserServiceClient{}, bankingClient)

	order := &model.Order{
		OrderID:          1,
		ListingID:        1,
		OrderType:        model.OrderTypeMarket,
		Direction:        model.OrderDirectionBuy,
		Quantity:         1,
		FilledQty:        0,
		ContractSize:     1,
		Triggered:        true,
		AllOrNone:        true,
		Status:           model.OrderStatusApproved,
		AccountNumber:    "444000100000000110",
		CommissionExempt: true,
	}

	err := svc.processOrder(context.Background(), order)
	require.NoError(t, err)
	require.True(t, order.IsDone)
	require.Equal(t, model.OrderStatusDeclined, order.Status)
}
