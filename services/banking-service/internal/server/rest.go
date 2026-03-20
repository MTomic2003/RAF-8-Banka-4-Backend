package server

import (
	"banking-service/internal/config"
	"banking-service/internal/handler"
	"banking-service/internal/validator"
	"common/pkg/auth"
	"common/pkg/errors"
	"common/pkg/logging"
	"context"
	stderrors "errors"
	"log"
	"net/http"
	"time"

	_ "banking-service/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
)

func NewServer(
	lc fx.Lifecycle,
	cfg *config.Configuration,
	healthHandler *handler.HealthHandler,
	accountHandler *handler.AccountHandler,
	companyHandler *handler.CompanyHandler,
	payeeHandler *handler.PayeeHandler,
	exchangeHandler *handler.ExchangeHandler,
	paymentHandler *handler.PaymentHandler,
	cardHandler *handler.CardHandler,
	loanHandler *handler.LoanHandler,
	transferHandler *handler.TransferHandler,
	verifier auth.TokenVerifier,
	permissions auth.PermissionProvider,
) {
	r := gin.New()

	InitRouter(r, cfg)
	SetupRoutes(r,
		healthHandler,
		accountHandler,
		companyHandler,
		transferHandler,
		payeeHandler,
		exchangeHandler,
		paymentHandler,
		cardHandler,
		loanHandler,
		verifier,
		permissions,
	)

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

func SetupRoutes(
	r *gin.Engine,
	healthHandler *handler.HealthHandler,
	accountHandler *handler.AccountHandler,
	companyHandler *handler.CompanyHandler,
	transferHandler *handler.TransferHandler,
	payeeHandler *handler.PayeeHandler,
	exchangeHandler *handler.ExchangeHandler,
	paymentHandler *handler.PaymentHandler,
	cardHandler *handler.CardHandler,
	loanHandler *handler.LoanHandler,
	verifier auth.TokenVerifier,
	permissions auth.PermissionProvider,
) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		api.GET("/health", healthHandler.Health)

		exchange := api.Group("/exchange")
		{
			exchange.GET("/rates", exchangeHandler.GetRates)
			exchange.GET("/calculate", exchangeHandler.Calculate)
		}

		accounts := api.Group("/accounts")
		accounts.Use(auth.Middleware(verifier, permissions))
		{
			accounts.GET("", auth.RequireIdentityType(auth.IdentityEmployee), accountHandler.ListAccounts)
			accounts.POST("", auth.RequireIdentityType(auth.IdentityEmployee), accountHandler.Create)
		}

		client := api.Group("/clients/:clientId")
		client.Use(auth.Middleware(verifier, permissions))
		{
			clientAccounts := client.Group("/accounts")
			{
				clientAccounts.GET("", auth.RequireClientSelf("clientId", true), accountHandler.GetClientAccounts)

				account := clientAccounts.Group("/:accountNumber")
				{
					account.GET("", auth.RequireClientSelf("clientId", true), accountHandler.GetAccountDetails)
					account.GET("/cards", auth.RequireClientSelf("clientId", true), cardHandler.ListCardsByAccount)
					account.GET("/payments", auth.RequireClientSelf("clientId", true), paymentHandler.GetAccountPayments)
					account.PUT("/name", auth.RequireClientSelf("clientId", false), accountHandler.UpdateAccountName)
					account.PUT("/limits", auth.RequireClientSelf("clientId", false), accountHandler.ConfirmLimitsChange)
					account.POST("/limits/request", auth.RequireClientSelf("clientId", false), accountHandler.RequestLimitsChange)
				}
			}

			clientPayments := client.Group("/payments")
			{
				clientPayments.GET("", auth.RequireClientSelf("clientId", true), paymentHandler.GetClientPayments)
				clientPayments.POST("", auth.RequireClientSelf("clientId", false), paymentHandler.CreatePayment)
				clientPayments.GET("/:id", auth.RequireClientSelf("clientId", true), paymentHandler.GetPaymentByID)
				clientPayments.GET("/:id/receipt", auth.RequireClientSelf("clientId", true), paymentHandler.GetReceipt)
				clientPayments.POST("/:id/verify", auth.RequireClientSelf("clientId", false), paymentHandler.VerifyPayment)
			}

			clientLoans := client.Group("/loans")
			{
				clientLoans.GET("", auth.RequireClientSelf("clientId", true), loanHandler.GetLoans)
				clientLoans.GET("/:loanId", auth.RequireClientSelf("clientId", true), loanHandler.GetLoanByID)
				clientLoans.POST("/request", auth.RequireClientSelf("clientId", false), loanHandler.SubmitLoanRequest)
			}

			clientTransfers := client.Group("/transfers")
			{
				clientTransfers.GET("", auth.RequireClientSelf("clientId", true), transferHandler.GetTransferHistory)
				clientTransfers.POST("", auth.RequireClientSelf("clientId", false), transferHandler.ExecuteTransfer)
			}
		}

		companies := api.Group("/companies")
		companies.Use(auth.Middleware(verifier, permissions))
		{
			companies.POST("", auth.RequireIdentityType(auth.IdentityEmployee), companyHandler.Create)
		}

		payees := api.Group("/payees")
		payees.Use(auth.Middleware(verifier, permissions))
		{
			payees.GET("", auth.RequireIdentityType(auth.IdentityClient), payeeHandler.GetAll)
			payees.POST("", auth.RequireIdentityType(auth.IdentityClient), payeeHandler.Create)
			payees.PATCH("/:id", auth.RequireIdentityType(auth.IdentityClient), payeeHandler.Update)
			payees.DELETE("/:id", auth.RequireIdentityType(auth.IdentityClient), payeeHandler.Delete)
		}

		cards := api.Group("/cards")
		cards.Use(auth.Middleware(verifier, permissions))
		{
			cards.POST("/request", auth.RequireIdentityType(auth.IdentityClient), cardHandler.RequestCard)
			cards.POST("/request/confirm", auth.RequireIdentityType(auth.IdentityClient), cardHandler.ConfirmCardRequest)
			cards.PUT("/:cardId/block", auth.RequireIdentityType(auth.IdentityClient, auth.IdentityEmployee), cardHandler.BlockCard)
			cards.PUT("/:cardId/unblock", auth.RequireIdentityType(auth.IdentityEmployee), cardHandler.UnblockCard)
			cards.PUT("/:cardId/deactivate", auth.RequireIdentityType(auth.IdentityEmployee), cardHandler.DeactivateCard)
		}

		loanRequests := api.Group("/loan-requests")
		loanRequests.Use(auth.Middleware(verifier, permissions))
		{
			loanRequests.GET("", auth.RequireIdentityType(auth.IdentityEmployee), loanHandler.ListLoanRequests)
			loanRequests.PATCH("/:id/approve", auth.RequireIdentityType(auth.IdentityEmployee), loanHandler.ApproveLoanRequest)
			loanRequests.PATCH("/:id/reject", auth.RequireIdentityType(auth.IdentityEmployee), loanHandler.RejectLoanRequest)
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
