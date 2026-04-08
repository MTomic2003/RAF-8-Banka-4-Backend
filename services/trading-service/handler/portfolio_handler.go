package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	pkgerrors "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/service"
)

type PortfolioHandler struct {
	service *service.PortfolioService
}

func NewPortfolioHandler(service *service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{service: service}
}

// GetClientPortfolio godoc
// @Summary Get portfolio for a client
// @Description Returns all currently held asset positions for a client, aggregated from all orders. Only approved orders with fills are counted. Includes stocks, futures, options, and forex pairs.
// @Tags portfolio
// @Security BearerAuth
// @Produce json
// @Param clientId path int true "Client ID"
// @Success 200 {array} dto.PortfolioAssetResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Failure 403 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /api/client/{clientId}/assets [get]
func (h *PortfolioHandler) GetClientPortfolio(c *gin.Context) {
	clientID, err := strconv.ParseUint(c.Param("clientId"), 10, 64)
	if err != nil {
		c.Error(pkgerrors.BadRequestErr("invalid client id"))
		return
	}

	assets, err := h.service.GetPortfolio(c.Request.Context(), uint(clientID), model.OwnerTypeClient)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, assets)
}

// GetActuaryPortfolio godoc
// @Summary Get portfolio for an actuary/agent
// @Description Returns all currently held asset positions for an actuary (employee agent/supervisor), aggregated from all orders. Only approved orders with fills are counted. Includes stocks, futures, options, and forex pairs.
// @Tags portfolio
// @Security BearerAuth
// @Produce json
// @Param actId path int true "Actuary ID"
// @Success 200 {array} dto.PortfolioAssetResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Failure 403 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /api/actuary/{actId}/assets [get]
func (h *PortfolioHandler) GetActuaryPortfolio(c *gin.Context) {
	actID, err := strconv.ParseUint(c.Param("actId"), 10, 64)
	if err != nil {
		c.Error(pkgerrors.BadRequestErr("invalid actuary id"))
		return
	}

	assets, err := h.service.GetPortfolio(c.Request.Context(), uint(actID), model.OwnerTypeActuary)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, assets)
}

// GetClientPortfolioProfit godoc
// @Summary Get total profit for a client portfolio
// @Description Returns the total accumulated profit across all currently held asset positions for a client.
// @Tags portfolio
// @Security BearerAuth
// @Produce json
// @Param clientId path int true "Client ID"
// @Success 200 {object} dto.PortfolioProfitResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Failure 403 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /api/client/{clientId}/assets/profit [get]
func (h *PortfolioHandler) GetClientPortfolioProfit(c *gin.Context) {
	clientID, err := strconv.ParseUint(c.Param("clientId"), 10, 64)
	if err != nil {
		c.Error(pkgerrors.BadRequestErr("invalid client id"))
		return
	}

	assets, err := h.service.GetPortfolio(c.Request.Context(), uint(clientID), model.OwnerTypeClient)
	if err != nil {
		c.Error(err)
		return
	}

	var total float64
	for _, a := range assets {
		total += a.Profit
	}

	c.JSON(http.StatusOK, dto.PortfolioProfitResponse{TotalProfitRSD: total})
}

// GetActuaryPortfolioProfit godoc
// @Summary Get total profit for an actuary portfolio
// @Description Returns the total accumulated profit across all currently held asset positions for an actuary.
// @Tags portfolio
// @Security BearerAuth
// @Produce json
// @Param actId path int true "Actuary ID"
// @Success 200 {object} dto.PortfolioProfitResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Failure 403 {object} errors.AppError
// @Failure 500 {object} errors.AppError
// @Router /api/actuary/{actId}/assets/profit [get]
func (h *PortfolioHandler) GetActuaryPortfolioProfit(c *gin.Context) {
	actID, err := strconv.ParseUint(c.Param("actId"), 10, 64)
	if err != nil {
		c.Error(pkgerrors.BadRequestErr("invalid actuary id"))
		return
	}

	assets, err := h.service.GetPortfolio(c.Request.Context(), uint(actID), model.OwnerTypeActuary)
	if err != nil {
		c.Error(err)
		return
	}

	var total float64
	for _, a := range assets {
		total += a.Profit
	}

	c.JSON(http.StatusOK, dto.PortfolioProfitResponse{TotalProfitRSD: total})
}
