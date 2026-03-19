package handler

import (
	"banking-service/internal/dto"
	"banking-service/internal/service"
	"common/pkg/errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PayeeHandler struct {
	service *service.PayeeService
}

func NewPayeeHandler(service *service.PayeeService) *PayeeHandler {
	return &PayeeHandler{service: service}
}

func (h *PayeeHandler) GetAll(c *gin.Context) {
	payees, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response := make([]dto.PayeeResponse, len(payees))
	for i, p := range payees {
		response[i] = dto.ToPayeeResponse(&p)
	}

	c.JSON(http.StatusOK, response)
}

func (h *PayeeHandler) Create(c *gin.Context) {
	var req dto.CreatePayeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}

	payee, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, dto.ToPayeeResponse(payee))
}

func (h *PayeeHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errors.BadRequestErr("invalid payee id"))
		return
	}

	var req dto.UpdatePayeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.BadRequestErr(err.Error()))
		return
	}

	payee, err := h.service.Update(c.Request.Context(), uint(id), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.ToPayeeResponse(payee))
}

func (h *PayeeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errors.BadRequestErr("invalid payee id"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}