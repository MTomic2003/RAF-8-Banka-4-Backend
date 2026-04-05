//go:build schema

package model

import (
	"context"
	"testing"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestSchemaAutoMigrate_UserService(t *testing.T) {
	ctx := context.Background()
	container, err := tcpostgres.Run(
		ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("user_service_schema_check"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	defer func() {
		_ = container.Terminate(context.Background())
	}()

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("build connection string: %v", err)
	}

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := db.AutoMigrate(
		&Identity{},
		&Employee{},
		&ActuaryInfo{},
		&Client{},
		&Position{},
		&ActivationToken{},
		&ResetToken{},
		&RefreshToken{},
		&EmployeePermission{},
	); err != nil {
		t.Fatalf("auto migrate schema: %v", err)
	}

	if !db.Migrator().HasTable(&Identity{}) || !db.Migrator().HasTable(&Client{}) {
		t.Fatal("expected key user-service tables to exist after AutoMigrate")
	}
}
