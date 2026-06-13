package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	appErrors "github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/repository"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/banking-service/internal/service"
)

type handlerUserClient struct{}

func (handlerUserClient) GetClientByID(context.Context, uint) (*pb.GetClientByIdResponse, error) {
	return &pb.GetClientByIdResponse{Id: 9, FullName: "Client", Email: "client@example.com"}, nil
}

func (handlerUserClient) GetEmployeeByID(context.Context, uint) (*pb.GetEmployeeByIdResponse, error) {
	return &pb.GetEmployeeByIdResponse{Id: 1, FullName: "Employee One", Email: "employee@example.com"}, nil
}

func setupBankingHandlerDB(t *testing.T, models ...interface{}) *gorm.DB {
	t.Helper()

	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func requestJSON(router http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, _ := json.Marshal(body)
		reader = bytes.NewReader(payload)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestPayeeHandlerLifecycle(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	db := setupBankingHandlerDB(t, &model.Payee{})
	payeeHandler := NewPayeeHandler(service.NewPayeeService(repository.NewPayeeRepository(db)))
	clientID := uint(42)

	router := gin.New()
	router.Use(appErrors.ErrorHandler())
	router.Use(func(c *gin.Context) {
		auth.SetAuth(c, &auth.AuthContext{IdentityID: 100, IdentityType: auth.IdentityClient, ClientID: &clientID})
		c.Next()
	})
	router.GET("/payees", payeeHandler.GetAll)
	router.POST("/payees", payeeHandler.Create)
	router.PATCH("/payees/:id", payeeHandler.Update)
	router.DELETE("/payees/:id", payeeHandler.Delete)

	rec := requestJSON(router, http.MethodPost, "/payees", gin.H{"name": "Supplier", "account_number": "444000100000000001"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d body=%s", rec.Code, rec.Body.String())
	}
	var created struct {
		PayeeID uint `json:"payee_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created payee: %v", err)
	}
	if created.PayeeID == 0 {
		t.Fatal("expected created payee id")
	}

	rec = requestJSON(router, http.MethodGet, "/payees", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get all status = %d body=%s", rec.Code, rec.Body.String())
	}
	var list []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode payee list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len(list) = %d, want 1", len(list))
	}

	rec = requestJSON(router, http.MethodPatch, "/payees/1", gin.H{"name": "Updated Supplier"})
	if rec.Code != http.StatusOK {
		t.Fatalf("update status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = requestJSON(router, http.MethodDelete, "/payees/1", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = requestJSON(router, http.MethodPatch, "/payees/bad", gin.H{"name": "Bad"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad id status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestExchangeHandlerRatesAndCalculate(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	db := setupBankingHandlerDB(t, &model.ExchangeRate{})
	now := time.Now().UTC()
	if err := db.Create(&model.ExchangeRate{
		CurrencyCode:         model.EUR,
		BaseCurrency:         model.RSD,
		BuyRate:              117,
		MiddleRate:           118,
		SellRate:             119,
		ProviderUpdatedAt:    now,
		ProviderNextUpdateAt: now.Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("create exchange rate: %v", err)
	}

	exchangeHandler := NewExchangeHandler(service.NewExchangeService(repository.NewExchangeRateRepository(db), nil))
	router := gin.New()
	router.Use(appErrors.ErrorHandler())
	router.GET("/rates", exchangeHandler.GetRates)
	router.GET("/calculate", exchangeHandler.Calculate)

	rec := requestJSON(router, http.MethodGet, "/rates", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("rates status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = requestJSON(router, http.MethodGet, "/calculate?amount=119&from_currency=RSD&to_currency=EUR", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("calculate status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = requestJSON(router, http.MethodGet, "/calculate?amount=-1&from_currency=RSD&to_currency=EUR", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("negative amount status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = requestJSON(router, http.MethodGet, "/calculate?amount=10&from_currency=BAD&to_currency=EUR", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad currency status = %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCompanyHandlerListsAndCreates(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	db := setupBankingHandlerDB(t, &model.Company{}, &model.WorkCode{})
	workCode := model.WorkCode{Code: "62.01", Description: "Software"}
	if err := db.Create(&workCode).Error; err != nil {
		t.Fatalf("create work code: %v", err)
	}
	companyHandler := NewCompanyHandler(service.NewCompanyService(repository.NewCompanyRepository(db), handlerUserClient{}, db, nil))
	router := gin.New()
	router.Use(appErrors.ErrorHandler())
	router.GET("/companies", companyHandler.GetCompanies)
	router.GET("/work-codes", companyHandler.GetWorkCodes)
	router.POST("/companies", companyHandler.Create)

	rec := requestJSON(router, http.MethodGet, "/work-codes", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("work codes status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = requestJSON(router, http.MethodPost, "/companies", gin.H{
		"name":                "Acme",
		"registration_number": "12345678",
		"tax_number":          "123456789",
		"work_code_id":        workCode.WorkCodeID,
		"address":             "Main",
		"owner_id":            9,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create company status = %d body=%s", rec.Code, rec.Body.String())
	}

	rec = requestJSON(router, http.MethodGet, "/companies", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("companies status = %d body=%s", rec.Code, rec.Body.String())
	}
	var companies []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &companies); err != nil {
		t.Fatalf("decode companies: %v", err)
	}
	if len(companies) != 1 {
		t.Fatalf("len(companies) = %d, want 1", len(companies))
	}

	rec = requestJSON(router, http.MethodPost, "/companies", gin.H{"name": "missing fields"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid company status = %d body=%s", rec.Code, rec.Body.String())
	}
}
