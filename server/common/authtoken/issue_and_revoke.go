package authtoken

import (
	"context"
	"crypto/sha256"
	"kzhikcn/internal/appinfo"
	"kzhikcn/pkg/config"
	"kzhikcn/pkg/data/cache"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func RevokeToken(ctx context.Context, t *TokenClaims) error {
	return cache.SetString(ctx, cache.Keys("http", "auth", "revokedTokens", t.ID), "1", time.Until(t.ExpiresAt.Time))
}

func IsRevoked(ctx context.Context, t *TokenClaims) (revoked bool, err error) {
	return cache.Exists(ctx, cache.Keys("http", "auth", "revokedTokens", t.ID))
}

func IssueToken(adminId uint, username string) (string, error) {
	now := time.Now()
	conf := config.GetConf()

	jwtConf := conf.Auth.JWT

	claims := &TokenClaims{
		IsAdmin: true,
		AdminId: adminId,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    appinfo.CurrentInfo.Name,
			Subject:   username,
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtConf.Expiry.Duration())),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	key := sha256.Sum256([]byte(jwtConf.Secret))
	return token.SignedString(key[:])
}
