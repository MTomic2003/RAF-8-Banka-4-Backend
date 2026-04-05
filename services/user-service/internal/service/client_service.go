package service

import (
	"context"
	crand "crypto/rand"
	"encoding/base32"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/auth"
	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/errors"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/config"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/model"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/repository"
)

type ClientService struct {
	clientRepo          repository.ClientRepository
	identityRepo        repository.IdentityRepository
	activationTokenRepo repository.ActivationTokenRepository
	emailService        Mailer
	cfg                 *config.Configuration
	txManager           repository.TransactionManager
}

func NewClientService(
	clientRepo repository.ClientRepository,
	identityRepo repository.IdentityRepository,
	activationTokenRepo repository.ActivationTokenRepository,
	emailService Mailer,
	cfg *config.Configuration,
	txManager repository.TransactionManager,
) *ClientService {
	return &ClientService{
		clientRepo:          clientRepo,
		identityRepo:        identityRepo,
		activationTokenRepo: activationTokenRepo,
		emailService:        emailService,
		cfg:                 cfg,
		txManager:           txManager,
	}
}

func (s *ClientService) Register(ctx context.Context, req *dto.CreateClientRequest) (*model.Client, error) {
	emailExists, err := s.identityRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	if emailExists {
		return nil, errors.ConflictErr("email already in use")
	}

	usernameExists, err := s.identityRepo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	if usernameExists {
		return nil, errors.ConflictErr("username already in use")
	}

	identity := &model.Identity{
		Email:    req.Email,
		Username: req.Username,
		Type:     auth.IdentityClient,
		Active:   false,
	}
	mobileSecret, err := generateMobileVerificationSecret()
	if err != nil {
		return nil, errors.InternalErr(err)
	}

	client := &model.Client{
		IdentityID:               identity.ID,
		FirstName:                req.FirstName,
		LastName:                 req.LastName,
		MobileVerificationSecret: mobileSecret,
		DateOfBirth:              req.DateOfBirth,
		Gender:                   req.Gender,
		PhoneNumber:              req.PhoneNumber,
		Address:                  req.Address,
	}
	tokenStr, err := generateSecureToken(16)
	if err != nil {
		return nil, errors.InternalErr(err)
	}

	activationToken := &model.ActivationToken{
		IdentityID: identity.ID,
		Token:      tokenStr,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	if err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.identityRepo.Create(txCtx, identity); err != nil {
			return errors.InternalErr(err)
		}

		client.IdentityID = identity.ID
		if err := s.clientRepo.Create(txCtx, client); err != nil {
			return errors.InternalErr(err)
		}

		activationToken.IdentityID = identity.ID
		if err := s.activationTokenRepo.Create(txCtx, activationToken); err != nil {
			return errors.InternalErr(err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	activationBase := strings.TrimRight(s.cfg.URLs.FrontendBaseURL, "/")
	link := fmt.Sprintf("%s/activate?token=%s", activationBase, url.QueryEscape(tokenStr))

	if err := s.emailService.Send(
		identity.Email,
		"Welcome!",
		fmt.Sprintf("Kliknite ovde da postavite lozinku: %s", link),
	); err != nil {
		return nil, errors.ServiceUnavailableErr(err)
	}

	client.Identity = *identity
	return client, nil
}

func (s *ClientService) GetAllClients(ctx context.Context, query *dto.ListClientsQuery) (*dto.ListClientsResponse, error) {
	clients, total, err := s.clientRepo.FindAll(ctx, query)
	if err != nil {
		return nil, errors.InternalErr(err)
	}

	return dto.ToClientResponseList(clients, total, query.Page, query.PageSize), nil
}

func (s *ClientService) UpdateClient(ctx context.Context, id uint, req *dto.UpdateClientRequest) (*dto.ClientResponse, error) {
	client, err := s.clientRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.InternalErr(err)
	}
	if client == nil {
		return nil, errors.NotFoundErr("client not found")
	}

	if req.FirstName != nil {
		client.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		client.LastName = *req.LastName
	}
	if req.Gender != nil {
		client.Gender = *req.Gender
	}
	if req.DateOfBirth != nil {
		client.DateOfBirth = *req.DateOfBirth
	}
	if req.PhoneNumber != nil {
		client.PhoneNumber = *req.PhoneNumber
	}
	if req.Address != nil {
		client.Address = *req.Address
	}

	if err := s.clientRepo.Update(ctx, client); err != nil {
		return nil, errors.InternalErr(err)
	}

	return dto.ToClientResponse(client), nil
}

func (s *ClientService) GetMobileVerificationSecret(ctx context.Context, clientID uint) (string, error) {
	client, err := s.clientRepo.FindByID(ctx, clientID)
	if err != nil {
		return "", errors.InternalErr(err)
	}
	if client == nil || client.MobileVerificationSecret == "" {
		return "", errors.NotFoundErr("mobile verification secret not found")
	}

	return client.MobileVerificationSecret, nil
}

func generateMobileVerificationSecret() (string, error) {
	secret := make([]byte, 20)
	if _, err := crand.Read(secret); err != nil {
		return "", err
	}

	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}
