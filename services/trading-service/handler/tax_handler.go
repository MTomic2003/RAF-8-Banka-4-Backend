package handler

import (
	"math"
	"net/http"
	"strconv"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/service"
	"github.com/gin-gonic/gin"
)

type TaxHandler struct {
	taxService *service.TaxService
	userClient client.UserServiceClient
}

func NewTaxHandler(taxService *service.TaxService, userClient client.UserServiceClient) *TaxHandler {
	return &TaxHandler{taxService: taxService, userClient: userClient}
}

// ListTaxUsers godoc
// @Summary List users with tax information
// @Description Returns a paginated list of clients and/or actuaries with their total tax owed in RSD. Filterable by user type, first name, and last name.
// @Tags tax
// @Produce json
// @Param userType query string false "Filter by user type (client, actuary)"
// @Param first_name query string false "Filter by first name"
// @Param last_name query string false "Filter by last name"
// @Param page query int false "Page number" minimum(1)
// @Param page_size query int false "Page size" minimum(1) maximum(100)
// @Success 200 {object} dto.ListTaxUsersResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Failure 403 {object} errors.AppError
// @Security BearerAuth
// @Router /api/tax/users [get]
func (h *TaxHandler) ListTaxUsers(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.ListTaxUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(errors.BadRequestErr(err.Error()))
		return
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}

	var entries []dto.UserTaxEntry

	if req.UserType == "" || req.UserType == "client" {
		clientsResp, err := h.userClient.GetAllClients(ctx, req.Page, req.PageSize, req.FirstName, req.LastName)
		if err != nil {
			_ = c.Error(errors.InternalErr(err))
			return
		}
		for _, cl := range clientsResp.Clients {
			tax, err := h.taxService.GetClientTotalTax(ctx, cl.Id)
			if err != nil {
				_ = c.Error(err)
				return
			}
			entries = append(entries, dto.UserTaxEntry{
				ID:         cl.Id,
				FirstName:  cl.FirstName,
				LastName:   cl.LastName,
				Email:      cl.Email,
				UserType:   "client",
				TaxOwedRSD: tax,
			})
		}
	}

	if req.UserType == "" || req.UserType == "actuary" {
		actuariesResp, err := h.userClient.GetAllActuaries(ctx, req.Page, req.PageSize, req.FirstName, req.LastName)
		if err != nil {
			_ = c.Error(errors.InternalErr(err))
			return
		}
		for _, act := range actuariesResp.Actuaries {
			tax, err := h.taxService.GetEmployeeTotalTax(ctx, uint(act.Id))
			if err != nil {
				_ = c.Error(err)
				return
			}
			entries = append(entries, dto.UserTaxEntry{
				ID:         act.Id,
				FirstName:  act.FirstName,
				LastName:   act.LastName,
				Email:      act.Email,
				UserType:   "actuary",
				TaxOwedRSD: tax,
			})
		}
	}

	if entries == nil {
		entries = []dto.UserTaxEntry{}
	}

	total := int64(len(entries))
	totalPages := int32(math.Ceil(float64(total) / float64(req.PageSize)))

	c.JSON(http.StatusOK, dto.ListTaxUsersResponse{
		Data:       entries,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	})
}

// CollectTaxes godoc
// @Summary Trigger tax collection
// @Description Runs the tax collection process for all users. Restricted to authorized personnel.
// @Tags tax
// @Produce json
// @Success 200 {object} dto.CollectTaxesResponse
// @Failure 401 {object} errors.AppError
// @Failure 403 {object} errors.AppError
// @Security BearerAuth
// @Router /api/tax/collect [post]
func (h *TaxHandler) CollectTaxes(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.taxService.CollectTaxes(ctx); err != nil {
		_ = c.Error(errors.InternalErr(err))
		return
	}

	c.JSON(http.StatusOK, dto.CollectTaxesResponse{Message: "Tax collection completed"})
}

// GetClientAccumulatedTax godoc
// @Summary Get accumulated tax for a client
// @Description Returns the total accumulated tax for a specific client across all their accounts in RSD.
// @Tags tax
// @Produce json
// @Param clientId path int true "Client ID"
// @Success 200 {object} dto.TaxInfoResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Failure 403 {object} errors.AppError
// @Security BearerAuth
// @Router /api/client/{clientId}/accumulated-tax [get]
func (h *TaxHandler) GetClientAccumulatedTax(c *gin.Context) {
	ctx := c.Request.Context()
	clientIDStr := c.Param("clientId")
	clientID, err := strconv.ParseUint(clientIDStr, 10, 64)
	if err != nil {
		_ = c.Error(errors.BadRequestErr("invalid clientId"))
		return
	}

	total, err := h.taxService.GetClientTotalTax(ctx, clientID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.TaxInfoResponse{
		TotalTax: total,
	})
}

// GetActuaryAccumulatedTax godoc
// @Summary Get accumulated tax for an actuary
// @Description Returns the total accumulated tax for a specific actuary (employee) across all accounts in RSD.
// @Tags tax
// @Produce json
// @Param actId path int true "Actuary ID"
// @Success 200 {object} dto.TaxInfoResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Failure 403 {object} errors.AppError
// @Security BearerAuth
// @Router /api/actuary/{actId}/accumulated-tax [get]
func (h *TaxHandler) GetActuaryAccumulatedTax(c *gin.Context) {
	ctx := c.Request.Context()
	actIDStr := c.Param("actId")
	actID, err := strconv.ParseUint(actIDStr, 10, 64)
	if err != nil {
		_ = c.Error(errors.BadRequestErr("invalid actId"))
		return
	}

	total, err := h.taxService.GetEmployeeTotalTax(ctx, uint(actID))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.TaxInfoResponse{
		TotalTax: total,
	})
}
