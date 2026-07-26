package service

import (
	"context"
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/log"
	"kzhikcn/server/common/authtoken"
)

type AuthHooks struct {
	OnLoginSuccess []func(ctx context.Context, admin *data.Admin) error
	OnLoginFailed  []func(ctx context.Context, username, reason string) error
	OnLogout       []func(ctx context.Context, claims *authtoken.TokenClaims) error
}

func (a *AuthHooks) triggerFailed(ctx context.Context, username, reason string) error {
	for _, fn := range a.OnLoginFailed {
		if err := fn(ctx, username, reason); err != nil {
			log.Errorf("login failed: %s, reason: %s, err: %v", username, reason, err)
			return err
		}
	}

	return nil
}

func (a *AuthHooks) triggerSuccess(ctx context.Context, admin *data.Admin) error {
	for _, fn := range a.OnLoginSuccess {
		if err := fn(ctx, admin); err != nil {
			log.Errorf("login success: %s, err: %v", admin.Username, err)
			return err
		}
	}

	return nil
}

func (a *AuthHooks) triggerLogout(ctx context.Context, claims *authtoken.TokenClaims) error {
	for _, fn := range a.OnLogout {
		if err := fn(ctx, claims); err != nil {
			log.Errorf("logout: %s, err: %v", claims.Subject, err)
			return err
		}
	}

	return nil
}
