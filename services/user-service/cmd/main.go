package main

import (
	"common/pkg/auth"
	"common/pkg/db"
	"common/pkg/jwt"
	"common/pkg/logging"
	"user-service/internal/config"
	"user-service/internal/grpc"
	"user-service/internal/handler"
	"user-service/internal/model"
	"user-service/internal/permission"
	"user-service/internal/repository"
	"user-service/internal/seed"
	"user-service/internal/server"
	"user-service/internal/service"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

// @title User Service API
// @version 1.0
// @description API for managing employees, clients, authentication, and permissions.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer" followed by a space and your token. Example: "Bearer eyJhbGci..."
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
			func(database *gorm.DB) auth.PermissionProvider {
				return permission.NewDBPermissionProvider(database)
			},

			repository.NewIdentityRepository,
			repository.NewEmployeeRepository,
			repository.NewClientRepository,
			repository.NewActivationTokenRepository,
			repository.NewResetTokenRepository,
			repository.NewRefreshTokenRepository,
			repository.NewPositionRepository,
			service.NewAuthService,
			service.NewEmployeeService,
			service.NewClientService,
			service.NewEmailService,
			handler.NewAuthHandler,
			handler.NewEmployeeHandler,
			handler.NewClientHandler,
			handler.NewHealthHandler,
			grpc.NewPermissionService,
			grpc.NewUserService,
		),
		fx.Invoke(func(cfg *config.Configuration) error {
			return logging.Init(cfg.Env)
		}),
		fx.Invoke(func(db *gorm.DB) error {
			if err := db.AutoMigrate(
				&model.Identity{},
				&model.Employee{},
				&model.Client{},
				&model.Position{},
				&model.ActivationToken{},
				&model.ResetToken{},
				&model.RefreshToken{},
				&model.EmployeePermission{},
			); err != nil {
				return err
			}
			return seed.Run(db)
		}),
		fx.Invoke(server.NewServer, server.NewGRPCServer),
	).Run()
}
