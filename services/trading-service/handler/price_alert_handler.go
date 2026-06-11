package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/service"
)

type PriceAlertHandler struct {
	svc *service.PriceAlertService
}

func NewPriceAlertHandler(svc *service.PriceAlertService) *PriceAlertHandler {
	return &PriceAlertHandler{svc: svc}
}

func parsePriceAlertID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("priceAlertId"), 10, 64)
	if err != nil {
		return 0, errors.BadRequestErr("invalid price alert id")
	}
	return uint(id), nil
}

// GetMyPriceAlerts godoc
// @Summary List my price alerts
// @Description Lists every price alert owned by the authenticated user (active and already-triggered ones).
// @Tags price-alerts
// @Produce json
// @Success 200 {array} dto.PriceAlertResponse
// @Failure 401 {object} errors.AppError
// @Failure 403 {object} errors.AppError
// @Security BearerAuth
// @Router /api/price-alerts [get]
func (h *PriceAlertHandler) GetMyPriceAlerts(c *gin.Context) {
	result, err := h.svc.ListMyAlerts(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// CreatePriceAlert godoc
// @Summary Create a price alert
// @Description Creates a one-shot price alert for the authenticated user. When the listing's current price crosses the threshold in the configured direction, the user receives an email and the alert auto-deactivates.
// @Tags price-alerts
// @Accept json
// @Produce json
// @Param request body dto.CreatePriceAlertRequest true "Alert definition"
// @Success 201 {object} dto.PriceAlertResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Failure 404 {object} errors.AppError
// @Security BearerAuth
// @Router /api/price-alerts [post]
func (h *PriceAlertHandler) CreatePriceAlert(c *gin.Context) {
	var req dto.CreatePriceAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.BadRequestErr(err.Error()))
		return
	}
	result, err := h.svc.CreateAlert(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

// DeletePriceAlert godoc
// @Summary Delete a price alert
// @Description Deletes one of the authenticated user's price alerts. Other users' alerts return 404 (the resource is hidden, not refused with 403).
// @Tags price-alerts
// @Param priceAlertId path int true "Price alert id"
// @Success 204
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Failure 404 {object} errors.AppError
// @Security BearerAuth
// @Router /api/price-alerts/{priceAlertId} [delete]
func (h *PriceAlertHandler) DeletePriceAlert(c *gin.Context) {
	id, err := parsePriceAlertID(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.DeleteAlert(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}
