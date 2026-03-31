package client

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/config"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewBankingServiceConnection(lc fx.Lifecycle, cfg *config.Configuration) (*BankingConn, error) {
	conn, err := grpc.NewClient(
		cfg.BankingServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return conn.Close()
		},
	})

	return &BankingConn{ClientConn: conn}, nil
}

// BankingConn wraps grpc.ClientConn to distinguish it from the user-service connection in DI.
type BankingConn struct {
	*grpc.ClientConn
}
