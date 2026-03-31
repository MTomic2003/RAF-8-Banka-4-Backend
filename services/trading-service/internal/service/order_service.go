package service

import (
	"context"
	"math"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
)

type OrderService struct {
	orderRepo     repository.OrderRepository
	exchangeRepo  repository.ExchangeRepository
	listingRepo   repository.ListingRepository
	userClient    pb.UserServiceClient
	bankingClient pb.BankingServiceClient
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	exchangeRepo repository.ExchangeRepository,
	listingRepo repository.ListingRepository,
	userClient pb.UserServiceClient,
	bankingClient pb.BankingServiceClient,
) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		exchangeRepo:  exchangeRepo,
		listingRepo:   listingRepo,
		userClient:    userClient,
		bankingClient: bankingClient,
	}
}

func (s *OrderService) GetOrders(ctx context.Context, query dto.ListOrdersQuery) ([]model.Order, int64, error) {
	orders, total, err := s.orderRepo.FindAll(ctx, query.Page, query.PageSize, nil, query.Status, query.Direction, query.IsDone)
	if err != nil {
		return nil, 0, errors.InternalErr(err)
	}

	return orders, total, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (*model.Order, error) {
	if err := validateOrderTypeFields(req); err != nil {
		return nil, err
	}

	authCtx := auth.GetAuthFromContext(ctx)
	if authCtx == nil {
		return nil, errors.UnauthorizedErr("not authenticated")
	}

	if err := s.validateAccount(ctx, req.AccountNumber, authCtx); err != nil {
		return nil, err
	}

	listing, err := s.listingRepo.FindByID(ctx, req.ListingID)
	if err != nil {
		return nil, errors.InternalErr(err)
	}

	if listing == nil {
		return nil, errors.NotFoundErr("listing not found")
	}

	pricePerUnit := calculatePricePerUnit(req, listing)
	afterHours := s.isAfterHours(ctx, listing.ExchangeMIC)
	order := model.Order{
		UserID:        authCtx.IdentityID,
		AccountNumber: req.AccountNumber,
		ListingID:     req.ListingID,
		Listing:       *listing,
		OrderType:     req.OrderType,
		Direction:     req.Direction,
		Quantity:      req.Quantity,
		ContractSize:  1,
		PricePerUnit:  pricePerUnit,
		LimitValue:    req.LimitValue,
		StopValue:     req.StopValue,
		AllOrNone:     req.AllOrNone,
		Margin:        req.Margin,
		AfterHours:    afterHours,
		IsDone:        false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	order.Status = s.resolveOrderStatus(ctx, authCtx, &order)

	if err := s.orderRepo.Create(ctx, &order); err != nil {
		return nil, errors.InternalErr(err)
	}

	return &order, nil
}

func (s *OrderService) ApproveOrder(ctx context.Context, orderID uint) (*model.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, errors.InternalErr(err)
	}

	if order == nil {
		return nil, errors.NotFoundErr("order not found")
	}

	if order.Status != model.OrderStatusPending {
		return nil, errors.BadRequestErr("only pending orders can be approved")
	}

	authCtx := auth.GetAuthFromContext(ctx)
	if authCtx == nil {
		return nil, errors.UnauthorizedErr("not authenticated")
	}

	approverID := authCtx.IdentityID
	order.Status = model.OrderStatusApproved
	order.ApprovedBy = &approverID
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, errors.InternalErr(err)
	}

	return order, nil
}

func (s *OrderService) DeclineOrder(ctx context.Context, orderID uint) (*model.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, errors.InternalErr(err)
	}

	if order == nil {
		return nil, errors.NotFoundErr("order not found")
	}

	if order.Status != model.OrderStatusPending {
		return nil, errors.BadRequestErr("only pending orders can be declined")
	}

	authCtx := auth.GetAuthFromContext(ctx)
	if authCtx == nil {
		return nil, errors.UnauthorizedErr("not authenticated")
	}

	approverID := authCtx.IdentityID
	order.Status = model.OrderStatusDeclined
	order.ApprovedBy = &approverID
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, errors.InternalErr(err)
	}

	return order, nil
}

func (s *OrderService) CancelOrder(ctx context.Context, orderID uint) (*model.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, errors.InternalErr(err)
	}

	if order == nil {
		return nil, errors.NotFoundErr("order not found")
	}

	authCtx := auth.GetAuthFromContext(ctx)
	if authCtx == nil {
		return nil, errors.UnauthorizedErr("not authenticated")
	}

	isOwner := order.UserID == authCtx.IdentityID
	isSupervisor, err := s.checkSupervisor(ctx)
	if err != nil {
		return nil, err
	}

	if !isOwner && !isSupervisor {
		return nil, errors.ForbiddenErr("only the order owner or a supervisor can cancel an order")
	}

	if order.Status != model.OrderStatusPending && order.Status != model.OrderStatusApproved {
		return nil, errors.BadRequestErr("only pending or approved orders can be cancelled")
	}

	if order.IsDone {
		return nil, errors.BadRequestErr("cannot cancel a completed order")
	}

	order.Status = model.OrderStatusDeclined
	order.IsDone = true
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Save(ctx, order); err != nil {
		return nil, errors.InternalErr(err)
	}

	return order, nil
}

func (s *OrderService) resolveOrderStatus(ctx context.Context, authCtx *auth.AuthContext, order *model.Order) model.OrderStatus {
	if authCtx.IdentityType == auth.IdentityClient {
		return model.OrderStatusApproved
	}

	if authCtx.IdentityType != auth.IdentityEmployee || authCtx.EmployeeID == nil {
		return model.OrderStatusPending
	}

	resp, err := s.userClient.GetEmployeeById(ctx, &pb.GetEmployeeByIdRequest{
		Id: uint64(*authCtx.EmployeeID),
	})
	if err != nil {
		return model.OrderStatusPending
	}

	if !resp.IsAgent {
		return model.OrderStatusPending
	}

	if resp.NeedApproval {
		return model.OrderStatusPending
	}

	orderValue := float64(order.Quantity) * order.ContractSize
	if order.PricePerUnit != nil {
		orderValue *= *order.PricePerUnit
	}

	remainingLimit := resp.OrderLimit - resp.UsedLimit
	if orderValue > remainingLimit {
		return model.OrderStatusPending
	}

	return model.OrderStatusApproved
}

func (s *OrderService) validateAccount(ctx context.Context, accountNumber string, authCtx *auth.AuthContext) error {
	account, err := s.bankingClient.GetAccountByNumber(ctx, &pb.GetAccountByNumberRequest{
		AccountNumber: accountNumber,
	})

	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return errors.NotFoundErr("account not found")
		}

		return errors.ServiceUnavailableErr(err)
	}

	if authCtx.IdentityType == auth.IdentityClient {
		if authCtx.ClientID == nil || uint64(*authCtx.ClientID) != account.ClientId {
			return errors.ForbiddenErr("account does not belong to you")
		}
	} else if authCtx.IdentityType == auth.IdentityEmployee {
		if account.AccountType != "Bank" {
			return errors.BadRequestErr("employees must use a bank account")
		}
	}

	return nil
}

func (s *OrderService) checkSupervisor(ctx context.Context) (bool, error) {
	authCtx := auth.GetAuthFromContext(ctx)
	if authCtx == nil || authCtx.IdentityType != auth.IdentityEmployee || authCtx.EmployeeID == nil {
		return false, nil
	}

	resp, err := s.userClient.GetEmployeeById(ctx, &pb.GetEmployeeByIdRequest{
		Id: uint64(*authCtx.EmployeeID),
	})

	if err != nil {
		return false, errors.InternalErr(err)
	}

	return resp.IsSupervisor, nil
}

func (s *OrderService) isAfterHours(ctx context.Context, exchangeMIC string) bool {
	exchange, err := s.exchangeRepo.FindByMicCode(ctx, exchangeMIC)
	if err != nil || exchange == nil {
		return false
	}

	if !exchange.TradingEnabled {
		return false
	}

	now := time.Now().UTC().Add(time.Duration(exchange.TimeZone) * time.Hour)
	closeTime, err := time.Parse("15:04", exchange.CloseTime)
	if err != nil {
		return false
	}

	closeToday := time.Date(now.Year(), now.Month(), now.Day(), closeTime.Hour(), closeTime.Minute(), 0, 0, now.Location())

	hoursUntilClose := closeToday.Sub(now).Hours()

	return hoursUntilClose >= 0 && hoursUntilClose < 4
}

func validateOrderTypeFields(req dto.CreateOrderRequest) error {
	switch req.OrderType {
	case model.OrderTypeLimit:
		if req.LimitValue == nil {
			return errors.BadRequestErr("limitValue is required for LIMIT orders")
		}
	case model.OrderTypeStop:
		if req.StopValue == nil {
			return errors.BadRequestErr("stopValue is required for STOP orders")
		}
	case model.OrderTypeStopLimit:
		if req.LimitValue == nil {
			return errors.BadRequestErr("limitValue is required for STOP_LIMIT orders")
		}

		if req.StopValue == nil {
			return errors.BadRequestErr("stopValue is required for STOP_LIMIT orders")
		}
	}
	return nil
}

func calculateCommission(orderType model.OrderType, pricePerUnit *float64) float64 {
	if pricePerUnit == nil {
		return 0
	}

	price := *pricePerUnit
	switch orderType {
	case model.OrderTypeMarket, model.OrderTypeStop:
		return math.Min(0.14*price, 7)
	case model.OrderTypeLimit, model.OrderTypeStopLimit:
		return math.Min(0.24*price, 12)
	default:
		return 0
	}
}

func calculatePricePerUnit(req dto.CreateOrderRequest, listing *model.Listing) *float64 {
	switch req.OrderType {
	case model.OrderTypeLimit, model.OrderTypeStopLimit:
		return req.LimitValue
	case model.OrderTypeStop:
		return req.StopValue
	case model.OrderTypeMarket:
		var price float64
		if req.Direction == model.OrderDirectionBuy {
			price = listing.Ask
		} else {
			price = listing.Price
		}
		return &price
	default:
		return nil
	}
}
