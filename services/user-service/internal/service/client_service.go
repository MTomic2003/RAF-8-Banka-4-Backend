package service

import (
	"common/pkg/auth"
	"common/pkg/errors"
	"context"
	crand "crypto/rand"
	"encoding/base32"
	"fmt"
	"net/url"
	"strings"
	"time"

	"user-service/internal/config"
	"user-service/internal/dto"
	"user-service/internal/model"
	"user-service/internal/repository"
)

type ClientService struct {
	clientRepo          repository.ClientRepository
	identityRepo        repository.IdentityRepository
	activationTokenRepo repository.ActivationTokenRepository
	emailService        Mailer
	cfg                 *config.Configuration
}

func NewClientService(
	clientRepo repository.ClientRepository,
	identityRepo repository.IdentityRepository,
	activationTokenRepo repository.ActivationTokenRepository,
	emailService Mailer,
	cfg *config.Configuration,
) *ClientService {
	return &ClientService{
		clientRepo:          clientRepo,
		identityRepo:        identityRepo,
		activationTokenRepo: activationTokenRepo,
		emailService:        emailService,
		cfg:                 cfg,
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
	if err := s.identityRepo.Create(ctx, identity); err != nil {
		return nil, errors.InternalErr(err)
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
	if err := s.clientRepo.Create(ctx, client); err != nil {
		return nil, errors.InternalErr(err)
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
	if err := s.activationTokenRepo.Create(ctx, activationToken); err != nil {
		return nil, errors.InternalErr(err)
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
func (s *ClientService) GetAllClients(ctx context.Context, query *dto.ListClientsQuery) ([]*model.Client, int64, error) {
	return s.clientRepo.FindAll(ctx, query)
}

func (s *ClientService) UpdateClient(ctx context.Context, id uint, req *dto.UpdateClientRequest) (*model.Client, error) {
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

	return client, nil
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
