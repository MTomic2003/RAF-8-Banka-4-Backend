package server

import (
	"context"
	stderrors "errors"
	"log"
	"net/http"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/handler"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/config"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/middleware"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/validator"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/logging"
	_ "github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/docs"
)

func NewServer(lc fx.Lifecycle, cfg *config.Configuration, healthHandler *handler.HealthHandler, exchangeHandler *handler.ExchangeHandler, orderHandler *handler.OrderHandler, portfolioHandler *handler.PortfolioHandler, listingHandler *handler.ListingHandler, verifier auth.TokenVerifier, permProvider auth.PermissionProvider, userClient pb.UserServiceClient) {
	r := gin.New()

	InitRouter(r, cfg)

	SetupRoutes(r, healthHandler, exchangeHandler, orderHandler, portfolioHandler, listingHandler, verifier, permProvider, userClient)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	RegisterServerLifecycle(lc, server)
}

func InitRouter(r *gin.Engine, cfg *config.Configuration) {
	r.Use(gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.URLs.FrontendBaseURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(logging.Logger())
	r.Use(errors.ErrorHandler())

	validator.RegisterValidators()
}

func SetupRoutes(r *gin.Engine, healthHandler *handler.HealthHandler, exchangeHandler *handler.ExchangeHandler, orderHandler *handler.OrderHandler, portfolioHandler *handler.PortfolioHandler, listingHandler *handler.ListingHandler, verifier auth.TokenVerifier, permProvider auth.PermissionProvider, userClient pb.UserServiceClient) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		api.GET("/health", healthHandler.Health)

		exchanges := api.Group("/exchanges")
		{
			exchanges.GET("", exchangeHandler.GetAll)
			exchanges.PATCH("/:micCode/toggle", exchangeHandler.ToggleTradingEnabled)
		}
    
		listings := api.Group("/listings")
		listings.Use(auth.Middleware(verifier, permProvider))
		{
			listings.GET("/stocks", auth.AnyOf(middleware.RequireSupervisor(userClient), middleware.RequireAgent(userClient), auth.RequireIdentityType(auth.IdentityClient)), listingHandler.GetStocks)
			listings.GET("/futures", auth.AnyOf(middleware.RequireSupervisor(userClient), middleware.RequireAgent(userClient), auth.RequireIdentityType(auth.IdentityClient)),  listingHandler.GetFutures)
			listings.GET("/forex", auth.AnyOf(middleware.RequireSupervisor(userClient), middleware.RequireAgent(userClient)), listingHandler.GetForex)
			listings.GET("/options", auth.AnyOf(middleware.RequireSupervisor(userClient), middleware.RequireAgent(userClient)), listingHandler.GetOptions)
    }

		authMw := auth.Middleware(verifier, permProvider)

		client := api.Group("/client")
		client.Use(authMw, auth.RequireClientSelf("clientId", true))
		client.GET("/:clientId/assets", portfolioHandler.GetClientPortfolio)

		actuary := api.Group("/actuary")
		actuary.Use(authMw, auth.RequireIdentityType(auth.IdentityEmployee))
		actuary.GET("/:actId/assets", portfolioHandler.GetActuaryPortfolio)
		orders := api.Group("/orders")
		orders.Use(auth.Middleware(verifier, permProvider))
		{
			orders.GET("", middleware.RequireSupervisor(userClient), orderHandler.GetOrders)
			orders.POST("", orderHandler.CreateOrder)
			orders.PATCH("/:id/approve", middleware.RequireSupervisor(userClient), orderHandler.ApproveOrder)
			orders.PATCH("/:id/decline", middleware.RequireSupervisor(userClient), orderHandler.DeclineOrder)
			orders.PATCH("/:id/cancel", orderHandler.CancelOrder)
		}
	}
}

func RegisterServerLifecycle(lc fx.Lifecycle, server *http.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := server.ListenAndServe(); err != nil && !stderrors.Is(err, http.ErrServerClosed) {
					log.Fatal(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
