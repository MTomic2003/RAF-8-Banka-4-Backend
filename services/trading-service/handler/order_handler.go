package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/service"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// GetOrders godoc
// @Summary Get all orders
// @Description Returns a paginated and filtered list of orders. Clients see only their own orders, employees see all.
// @Tags orders
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Filter by status (PENDING, APPROVED, DECLINED)"
// @Param direction query string false "Filter by direction (BUY, SELL)"
// @Param is_done query bool false "Filter by completion status"
// @Success 200 {object} dto.ListOrdersResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Router /api/orders [get]
func (h *OrderHandler) GetOrders(c *gin.Context) {
	var query dto.ListOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}

	orders, total, err := h.service.GetOrders(c.Request.Context(), query)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.ListOrdersResponse{
		Data:     dto.ToOrderSummaryResponseList(orders),
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	})
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Creates a buy or sell order for a listing
// @Tags orders
// @Accept json
// @Produce json
// @Param request body dto.CreateOrderRequest true "Order details"
// @Success 201 {object} dto.OrderResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Router /api/orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToOrderResponse(*order))
}

// ApproveOrder godoc
// @Summary Approve a pending order
// @Description Supervisor approves a pending order
// @Tags orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} dto.OrderResponse
// @Failure 400 {object} errors.AppError
// @Failure 404 {object} errors.AppError
// @Router /api/orders/{id}/approve [patch]
func (h *OrderHandler) ApproveOrder(c *gin.Context) {
	orderID, err := parseOrderID(c)
	if err != nil {
		c.Error(err)
		return
	}

	order, err := h.service.ApproveOrder(c.Request.Context(), orderID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.ToOrderResponse(*order))
}

// DeclineOrder godoc
// @Summary Decline a pending order
// @Description Supervisor declines a pending order
// @Tags orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} dto.OrderResponse
// @Failure 400 {object} errors.AppError
// @Failure 404 {object} errors.AppError
// @Router /api/orders/{id}/decline [patch]
func (h *OrderHandler) DeclineOrder(c *gin.Context) {
	orderID, err := parseOrderID(c)
	if err != nil {
		c.Error(err)
		return
	}

	order, err := h.service.DeclineOrder(c.Request.Context(), orderID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.ToOrderResponse(*order))
}

// CancelOrder godoc
// @Summary Cancel an order
// @Description Cancel a pending or approved order that hasn't been fully executed
// @Tags orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} dto.OrderResponse
// @Failure 400 {object} errors.AppError
// @Failure 404 {object} errors.AppError
// @Router /api/orders/{id}/cancel [patch]
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID, err := parseOrderID(c)
	if err != nil {
		c.Error(err)
		return
	}

	order, err := h.service.CancelOrder(c.Request.Context(), orderID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.ToOrderResponse(*order))
}

func parseOrderID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, errors.BadRequestErr("invalid order id")
	}

	return uint(id), nil
}
