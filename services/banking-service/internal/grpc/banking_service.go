package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/repository"
)

type BankingService struct {
	pb.UnimplementedBankingServiceServer
	accountRepo repository.AccountRepository
}

func NewBankingService(accountRepo repository.AccountRepository) *BankingService {
	return &BankingService{accountRepo: accountRepo}
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
