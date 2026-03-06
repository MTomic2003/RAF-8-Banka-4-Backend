package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	pb "github.com/RAF-SI-2025/Banka-4-Backend/proto/health"
)

func HealthHandler(client pb.HealthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp, err := client.Check(c.Request.Context(), &pb.HealthRequest{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"health": resp.Status,
		})
	}
}
