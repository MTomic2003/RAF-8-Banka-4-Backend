package dto_test

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
	govalidator "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/permission"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/dto"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/user-service/internal/validator"
)

func init() {
	validator.RegisterValidators()
}

func validate(s interface{}) error {
	v, ok := binding.Validator.Engine().(*govalidator.Validate)
	if !ok {
		panic("failed to get validator engine")
	}

	return v.Struct(s)
}

func TestCreateEmployeeRequest_Validation(t *testing.T) {
	t.Parallel()

	valid := dto.CreateEmployeeRequest{
		FirstName:  "Jane",
		LastName:   "Doe",
		Email:      "jane@example.com",
		Username:   "janedoe",
		PositionID: 1,
	}

	tests := []struct {
		name    string
		modify  func(r dto.CreateEmployeeRequest) dto.CreateEmployeeRequest
		wantErr bool
	}{
		{
			name:   "valid request",
			modify: func(r dto.CreateEmployeeRequest) dto.CreateEmployeeRequest { return r },
		},
		{
			name: "valid with permissions",
			modify: func(r dto.CreateEmployeeRequest) dto.CreateEmployeeRequest {
				r.Permissions = []permission.Permission{permission.EmployeeView, permission.EmployeeCreate}
				return r
			},
		},
		{
			name: "missing first name",
			modify: func(r dto.CreateEmployeeRequest) dto.CreateEmployeeRequest {
				r.FirstName = ""
				return r
			},
			wantErr: true,
		},
		{
			name: "missing last name",
			modify: func(r dto.CreateEmployeeRequest) dto.CreateEmployeeRequest {
				r.LastName = ""
				return r
			},
			wantErr: true,
		},
		{
			name: "missing email",
			modify: func(r dto.CreateEmployeeRequest) dto.CreateEmployeeRequest {
				r.Email = ""
				return r
			},
			wantErr: true,
		},
		{
			name: "invalid email format",
			modify: func(r dto.CreateEmployeeRequest) dto.CreateEmployeeRequest {
				r.Email = "not-an-email"
				return r
			},
			wantErr: true,
		},
		{
			name: "missing username",
			modify: func(r dto.CreateEmployeeRequest) dto.CreateEmployeeRequest {
				r.Username = ""
				return r
			},
			wantErr: true,
		},
		{
			name: "missing position id",
			modify: func(r dto.CreateEmployeeRequest) dto.CreateEmployeeRequest {
				r.PositionID = 0
				return r
			},
			wantErr: true,
		},
		{
			name: "first name too long",
			modify: func(r dto.CreateEmployeeRequest) dto.CreateEmployeeRequest {
				r.FirstName = "abcdefghijklmnopqrstu" // 21 chars
				return r
			},
			wantErr: true,
		},
		{
			name: "last name too long",
			modify: func(r dto.CreateEmployeeRequest) dto.CreateEmployeeRequest {
				r.LastName = string(make([]byte, 101))
				return r
			},
			wantErr: true,
		},
		{
			name: "invalid permission",
			modify: func(r dto.CreateEmployeeRequest) dto.CreateEmployeeRequest {
				r.Permissions = []permission.Permission{"invalid.perm"}
				return r
			},
			wantErr: true,
		},
		{
			name: "duplicate permissions",
			modify: func(r dto.CreateEmployeeRequest) dto.CreateEmployeeRequest {
				r.Permissions = []permission.Permission{permission.EmployeeView, permission.EmployeeView}
				return r
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := tt.modify(valid)
			err := validate(req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoginRequest_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     dto.LoginRequest
		wantErr bool
	}{
		{
			name: "valid",
			req:  dto.LoginRequest{Email: "user@example.com", Password: "secret"},
		},
		{
			name:    "missing email",
			req:     dto.LoginRequest{Password: "secret"},
			wantErr: true,
		},
		{
			name:    "invalid email",
			req:     dto.LoginRequest{Email: "bad", Password: "secret"},
			wantErr: true,
		},
		{
			name:    "missing password",
			req:     dto.LoginRequest{Email: "user@example.com"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validate(tt.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestActivateEmployeeRequest_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     dto.ActivateAccountRequest
		wantErr bool
	}{
		{
			name: "valid",
			req:  dto.ActivateAccountRequest{Token: "abc123", Password: "ValidPass12"},
		},
		{
			name:    "missing token",
			req:     dto.ActivateAccountRequest{Password: "ValidPass12"},
			wantErr: true,
		},
		{
			name:    "missing password",
			req:     dto.ActivateAccountRequest{Token: "abc123"},
			wantErr: true,
		},
		{
			name:    "weak password",
			req:     dto.ActivateAccountRequest{Token: "abc123", Password: "weak"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validate(tt.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestChangePasswordRequest_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     dto.ChangePasswordRequest
		wantErr bool
	}{
		{
			name: "valid",
			req:  dto.ChangePasswordRequest{OldPassword: "OldPass12", NewPassword: "NewPass99"},
		},
		{
			name:    "missing old password",
			req:     dto.ChangePasswordRequest{NewPassword: "NewPass99"},
			wantErr: true,
		},
		{
			name:    "missing new password",
			req:     dto.ChangePasswordRequest{OldPassword: "OldPass12"},
			wantErr: true,
		},
		{
			name:    "new password too weak",
			req:     dto.ChangePasswordRequest{OldPassword: "OldPass12", NewPassword: "weak"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validate(tt.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestResetPasswordRequest_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     dto.ResetPasswordRequest
		wantErr bool
	}{
		{
			name: "valid",
			req:  dto.ResetPasswordRequest{Token: "tok", NewPassword: "NewPass12"},
		},
		{
			name:    "missing token",
			req:     dto.ResetPasswordRequest{NewPassword: "NewPass12"},
			wantErr: true,
		},
		{
			name:    "missing password",
			req:     dto.ResetPasswordRequest{Token: "tok"},
			wantErr: true,
		},
		{
			name:    "weak password",
			req:     dto.ResetPasswordRequest{Token: "tok", NewPassword: "weak"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validate(tt.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRefreshRequest_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     dto.RefreshRequest
		wantErr bool
	}{
		{
			name: "valid",
			req:  dto.RefreshRequest{RefreshToken: "some-token"},
		},
		{
			name:    "missing refresh token",
			req:     dto.RefreshRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validate(tt.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestForgotPasswordRequest_Validation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		req     dto.ForgotPasswordRequest
		wantErr bool
	}{
		{
			name: "valid",
			req:  dto.ForgotPasswordRequest{Email: "user@example.com"},
		},
		{
			name:    "missing email",
			req:     dto.ForgotPasswordRequest{},
			wantErr: true,
		},
		{
			name:    "invalid email",
			req:     dto.ForgotPasswordRequest{Email: "not-email"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validate(tt.req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
