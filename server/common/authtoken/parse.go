package authtoken

import (
	"context"
	"crypto/sha256"
	"kzhikcn/pkg/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired     = jwt.ErrTokenExpired
	ErrTokenNotValidYet = jwt.ErrTokenNotValidYet
	ErrSignatureInvalid = jwt.ErrSignatureInvalid
)

func Parse(tokenStr string) (*TokenClaims, *jwt.Token, error) {
	conf := config.GetConf()
	jwtConf := conf.Auth.JWT

	claims := &TokenClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		key := sha256.Sum256([]byte(jwtConf.Secret))
		return key[:], nil
	})
	if err != nil {
		return nil, nil, err
	}

	return claims, token, err
}

func Validate(token *jwt.Token) error {
	claims := token.Claims
	exp, _ := claims.GetExpirationTime()
	nbf, _ := claims.GetNotBefore()
	now := time.Now()

	if exp.After(now) {
		return ErrTokenExpired
	}

	if nbf.Before(now) {
		return ErrTokenNotValidYet
	}

	if !token.Valid {
		return ErrSignatureInvalid
	}

	return nil
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
