package server

import (
	"common/pkg/auth"
	"common/pkg/errors"
	"common/pkg/logging"
	"common/pkg/permission"
	"context"
	stderrors "errors"
	"log"
	"net/http"
	"time"

	"user-service/internal/config"
	"user-service/internal/handler"
	"user-service/internal/validator"

	_ "user-service/docs"

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
	authHandler *handler.AuthHandler,
	empHandler *handler.EmployeeHandler,
	clientHandler *handler.ClientHandler,
	verifier auth.TokenVerifier,
	permissions auth.PermissionProvider,
) {
	r := gin.New()

	InitRouter(r, cfg)
	SetupRoutes(r, healthHandler, authHandler, empHandler, clientHandler, verifier, permissions)

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
	authHandler *handler.AuthHandler,
	empHandler *handler.EmployeeHandler,
	clientHandler *handler.ClientHandler,
	verifier auth.TokenVerifier,
	permissions auth.PermissionProvider,
) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		api.GET("/health", healthHandler.Health)

		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/activate", authHandler.Activate)
			authGroup.POST("/forgot-password", authHandler.ForgotPassword)
			authGroup.POST("/reset-password", authHandler.ResetPassword)
			authGroup.POST("/refresh", authHandler.RefreshToken)
		}

		authProtected := api.Group("/auth")
		authProtected.Use(auth.Middleware(verifier, permissions))
		{
			authProtected.POST("/change-password", authHandler.ChangePassword)
		}

		emp := api.Group("/employees")
		emp.Use(auth.Middleware(verifier, permissions))
		{
			emp.POST("/register", auth.RequirePermission(permission.EmployeeCreate), empHandler.Register)
			emp.GET("/:id", auth.RequirePermission(permission.EmployeeView), empHandler.GetEmployee)
			emp.PATCH("/:id", auth.RequirePermission(permission.EmployeeUpdate), empHandler.UpdateEmployee)
			emp.GET("", auth.RequirePermission(permission.EmployeeView), empHandler.ListEmployees)
		}

		cli := api.Group("/clients")
		cli.Use(auth.Middleware(verifier, permissions))
		{
			cli.GET("", auth.RequirePermission(permission.ClientView), clientHandler.ListClients)
			cli.POST("/register", clientHandler.Register)
			cli.PATCH("/:id", auth.RequirePermission(permission.ClientUpdate), clientHandler.UpdateClient)
		}

		mobileSecret := api.Group("")
		mobileSecret.Use(auth.Middleware(verifier, permissions), auth.RequireIdentityType(auth.IdentityClient))
		{
			mobileSecret.GET("/secret-mobile", clientHandler.GetMobileSecret)
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
