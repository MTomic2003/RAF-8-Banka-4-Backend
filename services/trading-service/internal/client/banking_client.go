package client

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
)

type BankingClient interface {
	GetAccountByNumber(ctx context.Context, accountNumber string) (*pb.GetAccountByNumberResponse, error)
	CreatePayment(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error)
}
