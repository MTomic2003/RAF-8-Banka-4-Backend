package grpc

import (
	"context"
	"fmt"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
)

type BankingServiceClient struct {
	stub pb.BankingServiceClient
}

func NewBankingServiceClient(conn *client.BankingConn) *BankingServiceClient {
	return &BankingServiceClient{
		stub: pb.NewBankingServiceClient(conn.ClientConn),
	}
}

func (c *BankingServiceClient) GetAccountByNumber(ctx context.Context, accountNumber string) (*pb.GetAccountByNumberResponse, error) {
	resp, err := c.stub.GetAccountByNumber(ctx, &pb.GetAccountByNumberRequest{
		AccountNumber: accountNumber,
	})
	if err != nil {
		return nil, fmt.Errorf("banking client GetAccountByNumber: %w", err)
	}
	return resp, nil
}

func (c *BankingServiceClient) CreatePaymentWithoutVerification(ctx context.Context, req *pb.CreatePaymentRequest) (*pb.CreatePaymentResponse, error) {
	resp, err := c.stub.CreatePayment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("banking client CreatePayment: %w", err)
	}
	return resp, nil
}
