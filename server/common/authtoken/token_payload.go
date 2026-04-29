package authtoken

import "github.com/golang-jwt/jwt/v5"

type TokenClaims struct {
	jwt.RegisteredClaims
	AdminId uint `json:"adminId"`
	IsAdmin bool `json:"is_admin"`
}
