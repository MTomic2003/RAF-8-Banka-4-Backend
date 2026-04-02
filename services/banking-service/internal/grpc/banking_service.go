package grpc

import (
	"context"

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

func (s *BankingService) CreatePaymentWithoutVerification(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	payment, err := s.paymentService.CreatePaymentWithoutVerification(ctx, dto.CreatePaymentRequest{
		PayerAccountNumber:     req.PayerAccountNumber,
		RecipientAccountNumber: req.RecipientAccountNumber,
		RecipientName:          req.RecipientName,
		Amount:                 req.Amount,
		ReferenceNumber:        req.ReferenceNumber,
		PaymentCode:            req.PaymentCode,
		Purpose:                req.Purpose,
	})
	if err != nil {
		return nil, errors.MapGrpcToHttpError(err)
	}
	return &pb.CreatePaymentResponse{
		PaymentId:     uint64(payment.PaymentID),
		TransactionId: uint64(payment.TransactionID),
		Status:        string(payment.Transaction.Status),
	}, nil
}
