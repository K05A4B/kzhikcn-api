package authtoken

import (
	"context"
	"crypto/sha256"
	"kzhikcn/pkg/config"

	"github.com/golang-jwt/jwt/v5"
)

func Parse(tokenStr string) (*TokenClaims, *jwt.Token, error) {
	conf := config.GetConf()
	jwtConf := conf.Auth.JWT

	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		key := sha256.Sum256([]byte(jwtConf.Secret))
		return key[:], nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, nil, err
	}

	return claims, token, err
}

type ctxKey struct{}

var ctxKeyInstance = ctxKey{}

func WithClaims(ctx context.Context, claims *TokenClaims) context.Context {
	return context.WithValue(ctx, ctxKeyInstance, claims)
}

func GetClaims(ctx context.Context) *TokenClaims {
	claims, ok := ctx.Value(ctxKeyInstance).(*TokenClaims)
	if !ok {
		return nil
	}

	return claims
}
