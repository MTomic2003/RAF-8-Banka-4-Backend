package service

import (
	"common/pkg/auth"
	"common/pkg/errors"
	"context"
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

	client := &model.Client{
		IdentityID:  identity.ID,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DateOfBirth: req.DateOfBirth,
		Gender:      req.Gender,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
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
