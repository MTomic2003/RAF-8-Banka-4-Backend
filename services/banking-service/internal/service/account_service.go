package service

import (
	"banking-service/internal/client"
	"banking-service/internal/dto"
	"banking-service/internal/model"
	"banking-service/internal/repository"
	"common/pkg/errors"
	"context"
	"fmt"
	"math/rand"

	"gorm.io/gorm"
)

const (
	bankCode   = "444"
	branchCode = "0001"
)

var subtypeTypeCodes = map[model.Subtype]string{
	model.SubtypeStandard:   "11",
	model.SubtypeSavings:    "13",
	model.SubtypePension:    "14",
	model.SubtypeYouth:      "15",
	model.SubtypeStudent:    "16",
	model.SubtypeUnemployed: "17",
	model.SubtypeLLC:        "12",
	model.SubtypeJointStock: "12",
	model.SubtypeFoundation: "12",
}

type AccountService struct {
	repo       repository.AccountRepository
	userClient client.UserClient
	db         *gorm.DB
}

func NewAccountService(
	repo repository.AccountRepository,
	userClient client.UserClient,
	db *gorm.DB,
) *AccountService {
	return &AccountService{
		repo:       repo,
		userClient: userClient,
		db:         db,
	}
}

func (s *AccountService) getTypeCode(accountKind model.AccountKind, accountType model.AccountType, subtype model.Subtype) string {
	if accountKind == model.AccountKindForeign {
		if accountType == model.AccountTypeBusiness {
			return "22"
		}
		return "21"
	}
	if code, ok := subtypeTypeCodes[subtype]; ok {
		return code
	}
	return "11"
}

func (s *AccountService) generateAccountNumber(typeCode string) string {
	random := fmt.Sprintf("%09d", rand.Intn(1_000_000_000))
	return bankCode + branchCode + random + typeCode
}

func (s *AccountService) isValidAccountNumber(number string) bool {
	sum := 0
	for _, ch := range number {
		sum += int(ch - '0')
	}
	return sum%11 != 0
}

func (s *AccountService) generateValidAccountNumber(accountKind model.AccountKind, accountType model.AccountType, subtype model.Subtype) string {
	typeCode := s.getTypeCode(accountKind, accountType, subtype)
	for {
		number := s.generateAccountNumber(typeCode)
		if s.isValidAccountNumber(number) {
			return number
		}
	}
}

func (s *AccountService) findCurrencyByCode(ctx context.Context, code string) (*model.Currency, error) {
	var currency model.Currency
	result := s.db.WithContext(ctx).Where("code = ?", code).First(&currency)
	if result.Error != nil {
		return nil, errors.NotFoundErr("currency not found: " + code)
	}
	return &currency, nil
}

func (s *AccountService) Create(ctx context.Context, req dto.CreateAccountRequest) (*model.Account, error) {
	if _, err := s.userClient.GetClientByID(ctx, req.ClientID); err != nil {
		return nil, errors.NotFoundErr("client not found")
	}

	if _, err := s.userClient.GetEmployeeByID(ctx, req.EmployeeID); err != nil {
		return nil, errors.NotFoundErr("employee not found")
	}

	if req.AccountType == model.AccountTypeBusiness && req.CompanyID == nil {
		return nil, errors.BadRequestErr("business account requires a company")
	}
	if req.AccountType == model.AccountTypePersonal && req.CompanyID != nil {
		return nil, errors.BadRequestErr("personal account cannot have a company")
	}

	currencyCode := "RSD"
	if req.AccountKind == model.AccountKindForeign {
		if req.CurrencyCode == "" {
			return nil, errors.BadRequestErr("currency code is required for foreign accounts")
		}
		currencyCode = req.CurrencyCode
	}

	if req.AccountKind == model.AccountKindCurrent && req.Subtype == "" {
		return nil, errors.BadRequestErr("subtype is required for current accounts")
	}

	currency, err := s.findCurrencyByCode(ctx, currencyCode)
	if err != nil {
		return nil, err
	}

	dailyLimit := 250000.0
	monthlyLimit := 1000000.0
	if req.AccountKind == model.AccountKindForeign {
		dailyLimit = 5000.0
		monthlyLimit = 20000.0
	}

	account := &model.Account{
		AccountNumber:    s.generateValidAccountNumber(req.AccountKind, req.AccountType, req.Subtype),
		Name:             req.Name,
		ClientID:         req.ClientID,
		EmployeeID:       req.EmployeeID,
		CompanyID:        req.CompanyID,
		Balance:          req.InitialBalance,
		AvailableBalance: req.InitialBalance,
		ExpiresAt:        req.ExpiresAt,
		CurrencyID:       currency.CurrencyID,
		AccountType:      req.AccountType,
		AccountKind:      req.AccountKind,
		Subtype:          req.Subtype,
		DailyLimit:       dailyLimit,
		MonthlyLimit:     monthlyLimit,
	}

	if err := s.repo.Create(ctx, account); err != nil {
		return nil, errors.InternalErr(err)
	}

	return account, nil
}
