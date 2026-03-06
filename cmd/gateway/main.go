package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	handlers "github.com/RAF-SI-2025/Banka-4-Backend/internal/http/handlers"
	healthClient "github.com/RAF-SI-2025/Banka-4-Backend/internal/clients/health"
)

func main() {
	healthClient, err := healthClient.New(os.Getenv("HEALTH_SERVICE_ADDR"))
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.GET("/health", handlers.HealthHandler(healthClient))

	log.Println("gateway listening on :8080")
	log.Fatal(r.Run(":8080"))
}
