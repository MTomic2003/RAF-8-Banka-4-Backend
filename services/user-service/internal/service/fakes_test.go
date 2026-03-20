package service

import (
	"common/pkg/auth"
	"common/pkg/permission"
	"context"
	"user-service/internal/config"
	"user-service/internal/dto"
	"user-service/internal/model"

	"golang.org/x/crypto/bcrypt"
)

type fakeIdentityRepo struct {
	byID       *model.Identity
	byEmail    *model.Identity
	byUsername *model.Identity

	emailExists    bool
	usernameExists bool

	findErr   error
	createErr error
	updateErr error
	existsErr error

	createdIdentity *model.Identity
	updatedIdentity *model.Identity
}

func (f *fakeIdentityRepo) Create(_ context.Context, identity *model.Identity) error {
	f.createdIdentity = identity
	identity.ID = 1
	return f.createErr
}

func (f *fakeIdentityRepo) FindByID(_ context.Context, _ uint) (*model.Identity, error) {
	return f.byID, f.findErr
}

func (f *fakeIdentityRepo) FindByEmail(_ context.Context, _ string) (*model.Identity, error) {
	return f.byEmail, f.findErr
}

func (f *fakeIdentityRepo) FindByUsername(_ context.Context, _ string) (*model.Identity, error) {
	return f.byUsername, f.findErr
}

func (f *fakeIdentityRepo) Update(_ context.Context, identity *model.Identity) error {
	f.updatedIdentity = identity
	return f.updateErr
}

func (f *fakeIdentityRepo) EmailExists(_ context.Context, _ string) (bool, error) {
	return f.emailExists, f.existsErr
}

func (f *fakeIdentityRepo) UsernameExists(_ context.Context, _ string) (bool, error) {
	return f.usernameExists, f.existsErr
}

type fakeEmployeeRepo struct {
	byID         *model.Employee
	byIdentityID *model.Employee
	allEmps      []model.Employee
	allTotal     int64

	findErr   error
	createErr error
	updateErr error
	getAllErr error

	createdEmployee *model.Employee
	updatedEmployee *model.Employee
}

func (f *fakeEmployeeRepo) FindByID(_ context.Context, _ uint) (*model.Employee, error) {
	return f.byID, f.findErr
}

func (f *fakeEmployeeRepo) FindByIdentityID(_ context.Context, _ uint) (*model.Employee, error) {
	return f.byIdentityID, f.findErr
}

func (f *fakeEmployeeRepo) Create(_ context.Context, emp *model.Employee) error {
	f.createdEmployee = emp
	return f.createErr
}

func (f *fakeEmployeeRepo) Update(_ context.Context, emp *model.Employee) error {
	f.updatedEmployee = emp
	return f.updateErr
}

func (f *fakeEmployeeRepo) GetAll(_ context.Context, _, _, _, _ string, _, _ int) ([]model.Employee, int64, error) {
	return f.allEmps, f.allTotal, f.getAllErr
}

type fakeClientRepo struct {
	byID         *model.Client
	byIdentityID *model.Client
	allClients   []model.Client
	allTotal     int64

	findErr   error
	createErr error
	updateErr error
	getAllErr  error

	createdClient *model.Client
	updatedClient *model.Client
}

func (f *fakeClientRepo) Create(_ context.Context, client *model.Client) error {
	f.createdClient = client
	return f.createErr
}

func (f *fakeClientRepo) FindByIdentityID(_ context.Context, _ uint) (*model.Client, error) {
	return f.byIdentityID, f.findErr
}

func (f *fakeClientRepo) FindByID(_ context.Context, id uint) (*model.Client, error) {
	return f.byID, f.findErr
}

func (f *fakeClientRepo) FindAll(_ context.Context, _ *dto.ListClientsQuery) ([]model.Client, int64, error) {
	return f.allClients, f.allTotal, f.getAllErr
}

func (f *fakeClientRepo) Update(_ context.Context, client *model.Client) error {
	f.updatedClient = client
	return f.updateErr
}

type fakeActivationTokenRepo struct {
	token     *model.ActivationToken
	findErr   error
	createErr error
	deleteErr error
}

func (f *fakeActivationTokenRepo) Create(_ context.Context, _ *model.ActivationToken) error {
	return f.createErr
}

func (f *fakeActivationTokenRepo) FindByToken(_ context.Context, _ string) (*model.ActivationToken, error) {
	return f.token, f.findErr
}

func (f *fakeActivationTokenRepo) Delete(_ context.Context, _ *model.ActivationToken) error {
	return f.deleteErr
}

type fakeResetTokenRepo struct {
	token     *model.ResetToken
	findErr   error
	createErr error
	deleteErr error
}

func (f *fakeResetTokenRepo) Create(_ context.Context, _ *model.ResetToken) error {
	return f.createErr
}

func (f *fakeResetTokenRepo) FindByToken(_ context.Context, _ string) (*model.ResetToken, error) {
	return f.token, f.findErr
}

func (f *fakeResetTokenRepo) DeleteByIdentityID(_ context.Context, _ uint) error {
	return f.deleteErr
}

type fakeRefreshTokenRepo struct {
	token     *model.RefreshToken
	findErr   error
	createErr error
	deleteErr error
}

func (f *fakeRefreshTokenRepo) Create(_ context.Context, _ *model.RefreshToken) error {
	return f.createErr
}

func (f *fakeRefreshTokenRepo) FindByToken(_ context.Context, _ string) (*model.RefreshToken, error) {
	return f.token, f.findErr
}

func (f *fakeRefreshTokenRepo) DeleteByIdentityID(_ context.Context, _ uint) error {
	return f.deleteErr
}

type fakePositionRepo struct {
	exists    bool
	existsErr error
}

func (f *fakePositionRepo) Exists(_ context.Context, _ uint) (bool, error) {
	return f.exists, f.existsErr
}

// --- Mailer Fake ---

type fakeMailer struct {
	sendErr error
	sent    bool
}

func (f *fakeMailer) Send(_, _, _ string) error {
	f.sent = true
	return f.sendErr
}

func testConfig() *config.Configuration {
	return &config.Configuration{
		JWTSecret:     "test-secret",
		JWTExpiry:     15,
		RefreshExpiry: 10080,
		URLs: config.URLConfig{
			FrontendBaseURL: "http://localhost:5173",
		},
	}
}

func hashedPassword(plain string) string {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	return string(h)
}

func activeIdentity() *model.Identity {
	return &model.Identity{
		ID:           1,
		Email:        "john@example.com",
		Username:     "johndoe",
		PasswordHash: hashedPassword("Password12"),
		Type:         auth.IdentityEmployee,
		Active:       true,
	}
}

func activeEmployee() *model.Employee {
	return &model.Employee{
		EmployeeID: 1,
		IdentityID: 1,
		FirstName:  "John",
		LastName:   "Doe",
		Identity:   *activeIdentity(),
		Permissions: []model.EmployeePermission{
			{EmployeeID: 1, Permission: permission.EmployeeView},
		},
	}
}

func activeClientIdentity() *model.Identity {
	return &model.Identity{
		ID:           2,
		Email:        "client@example.com",
		Username:     "clientuser",
		PasswordHash: hashedPassword("Password12"),
		Type:         auth.IdentityClient,
		Active:       true,
	}
}

func activeClient() *model.Client {
	return &model.Client{
		ClientID:   1,
		IdentityID: 2,
		FirstName:  "Jane",
		LastName:   "Client",
		Identity:   *activeClientIdentity(),
	}
}

func withAuth(ctx context.Context, identityID uint, identityType auth.IdentityType) context.Context {
	return auth.SetAuthOnContext(ctx, &auth.AuthContext{
		IdentityID:   identityID,
		IdentityType: identityType,
		Permissions:  []permission.Permission{permission.EmployeeView},
	})
}
