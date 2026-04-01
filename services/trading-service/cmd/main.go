package main

import (
	"context"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/handler"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/config"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/permission"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/repository"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/seed"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/server"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/service"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/db"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/jwt"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/logging"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
)

// @title Trading Service API
// @version 1.0
// @description API for managing portfolios, executing trades, and handling financial market operations.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme.
func main() {
	fx.New(
		fx.Provide(
			config.Load,
			func(cfg *config.Configuration) (*gorm.DB, error) {
				return db.New(cfg.DB.DSN())
			},
			func(cfg *config.Configuration) auth.TokenVerifier {
				return jwt.NewJWTVerifier(cfg.JWTSecret)
			},
			client.NewUserServiceConnection,
			func(conn *grpc.ClientConn) pb.PermissionServiceClient {
				return pb.NewPermissionServiceClient(conn)
			},
			func(conn *grpc.ClientConn) pb.UserServiceClient {
				return pb.NewUserServiceClient(conn)
			},
			client.NewBankingServiceConnection,
			func(conn *client.BankingConn) pb.BankingServiceClient {
				return pb.NewBankingServiceClient(conn.ClientConn)
			},
			func(c pb.PermissionServiceClient) auth.PermissionProvider {
				return permission.NewGrpcPermissionProvider(c)
			},
			handler.NewHealthHandler,
			repository.NewForexRepository,
			func(cfg *config.Configuration) client.ExchangeRateClient {
				return client.NewExchangeRateClient(cfg.ExchangeRateAPIKey)
			},
			service.NewForexService,

			func(cfg *config.Configuration) *client.StockClient {
				return client.NewStockClient(cfg.FinnhubAPIKey)
			},
			repository.NewListingRepository,
			repository.NewStockRepository,
			repository.NewOptionRepository,
			service.NewStockService,
			repository.NewExchangeRepository,
			service.NewExchangeService,
			handler.NewExchangeHandler,
			service.NewListingService,
			handler.NewListingHandler,
			repository.NewOrderOwnershipRepository,
			repository.NewFuturesContractRepository,
			service.NewPortfolioService,
			handler.NewPortfolioHandler,
			repository.NewOrderRepository,
			service.NewOrderService,
			handler.NewOrderHandler,
		),
		fx.Invoke(func(cfg *config.Configuration) error {
			return logging.Init(cfg.Env)
		}),
		fx.Invoke(func(db *gorm.DB) error {
			return db.AutoMigrate(
				&model.Listing{},
				&model.Stock{},
				&model.Option{},
				&model.ListingDailyPriceInfo{},
				&model.Exchange{},
				&model.Order{},
				&model.OrderOwnership{},
				&model.ForexPair{},
				&model.FuturesContract{},
			)
		}),
		fx.Invoke(func(lc fx.Lifecycle, svc *service.StockService) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go svc.Initialize(context.Background())
					svc.Start()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					svc.Stop()
					return nil
				},
			})
		}),
		fx.Invoke(func(db *gorm.DB) error {
			return seed.SeedFuturesContracts(db)
		}),
		fx.Invoke(func(db *gorm.DB) error {
			return seed.RunExchangeSeed(db)
		}),
		fx.Invoke(server.NewServer),
		fx.Invoke(func(lifecycle fx.Lifecycle, forexService *service.ForexService) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					forexService.Initialize(ctx)
					forexService.Start()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					forexService.Stop()
					return nil
				},
			})
		}),
	).Run()
}
