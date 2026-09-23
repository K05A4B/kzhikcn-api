package service

import (
	"context"
	"testing"

	"kzhikcn/pkg/data"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_NotifyLoginSuccess(t *testing.T) {
	var got *data.Admin

	hooks := &AuthHooks{
		OnLoginSuccess: []func(context.Context, *data.Admin) error{
			func(_ context.Context, admin *data.Admin) error {
				got = admin
				return nil
			},
		},
	}

	svc := NewAuthService(nil, hooks)
	admin := &data.Admin{ID: 1, Username: "admin"}

	svc.NotifyLoginSuccess(context.Background(), admin)

	require.NotNil(t, got)
	assert.Equal(t, uint(1), got.ID)
	assert.Equal(t, "admin", got.Username)
}

func TestAuthService_NotifyLoginFailed(t *testing.T) {
	var gotUsername, gotReason string

	hooks := &AuthHooks{
		OnLoginFailed: []func(context.Context, string, string) error{
			func(_ context.Context, username, reason string) error {
				gotUsername, gotReason = username, reason
				return nil
			},
		},
	}

	svc := NewAuthService(nil, hooks)

	svc.NotifyLoginFailed(context.Background(), "admin", "invalid totp code")

	assert.Equal(t, "admin", gotUsername)
	assert.Equal(t, "invalid totp code", gotReason)
}
