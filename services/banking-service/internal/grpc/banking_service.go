package grpc

import (
	"context"
	stderrors "errors"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/repository"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/service"
)

type BankingService struct {
	pb.UnimplementedBankingServiceServer
	accountRepo    repository.AccountRepository
	paymentService *service.PaymentService
}

func NewBankingService(accountRepo repository.AccountRepository, paymentService *service.PaymentService) *BankingService {
	return &BankingService{
		accountRepo:    accountRepo,
		paymentService: paymentService,
	}
}

func (s *BankingService) GetAccountByNumber(ctx context.Context, req *pb.GetAccountByNumberRequest) (*pb.GetAccountByNumberResponse, error) {
	account, err := s.accountRepo.FindByAccountNumber(ctx, req.AccountNumber)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch account: %v", err)
	}
	if account == nil {
		return nil, status.Errorf(codes.NotFound, "account %s not found", req.AccountNumber)
	}
	return &pb.GetAccountByNumberResponse{
		AccountNumber:    account.AccountNumber,
		ClientId:         uint64(account.ClientID),
		AccountType:      string(account.AccountType),
		CurrencyCode:     string(account.Currency.Code),
		AvailableBalance: account.AvailableBalance,
	}, nil
}

func (s *BankingService) CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	payment, err := s.paymentService.CreatePayment(ctx, dto.CreatePaymentRequest{
		PayerAccountNumber:     req.PayerAccountNumber,
		RecipientAccountNumber: req.RecipientAccountNumber,
		RecipientName:          req.RecipientName,
		Amount:                 req.Amount,
		ReferenceNumber:        req.ReferenceNumber,
		PaymentCode:            req.PaymentCode,
		Purpose:                req.Purpose,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.CreatePaymentResponse{
		PaymentId:     uint64(payment.PaymentID),
		TransactionId: uint64(payment.TransactionID),
		Status:        string(payment.Transaction.Status),
	}, nil
}

func mapError(err error) error {
	var appErr *errors.AppError
	if !stderrors.As(err, &appErr) {
		return status.Errorf(codes.Internal, "internal error: %v", err)
	}
	switch appErr.Code {
	case http.StatusNotFound:
		return status.Errorf(codes.NotFound, appErr.Message)
	case http.StatusBadRequest:
		return status.Errorf(codes.InvalidArgument, appErr.Message)
	case http.StatusUnauthorized:
		return status.Errorf(codes.Unauthenticated, appErr.Message)
	case http.StatusForbidden:
		return status.Errorf(codes.PermissionDenied, appErr.Message)
	case http.StatusConflict:
		return status.Errorf(codes.AlreadyExists, appErr.Message)
	case http.StatusServiceUnavailable:
		return status.Errorf(codes.Unavailable, appErr.Message)
	case http.StatusTooManyRequests:
		return status.Errorf(codes.ResourceExhausted, appErr.Message)
	default:
		return status.Errorf(codes.Internal, appErr.Message)
	}
}
