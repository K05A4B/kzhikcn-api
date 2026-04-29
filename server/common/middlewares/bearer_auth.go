package middlewares

import (
	"kzhikcn/pkg/utils"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/common/hdl"
	"net/http"
)

var ErrUnauthorized = hdl.DefineError(401, "未授权", "system.unauthorized")
var ErrParseTokenFailed = hdl.DefineError(500, "解析Token失败", "system.parse_token_failed")
var ErrValidateTokenIDFailed = hdl.DefineError(500, "检查Token ID失败", "system.validate_token_id_failed")

func BearerParse(h http.Handler) http.Handler {
	return hdl.Middleware(func(w http.ResponseWriter, r *http.Request, meta map[string]any) error {
		tokenStr := authtoken.GetTokenString(r)
		if utils.IsEmptyString(tokenStr) {
			h.ServeHTTP(w, r)
			return nil
		}

		claims, token, err := authtoken.Parse(tokenStr)
		if err != nil {
			h.ServeHTTP(w, r)
			return nil
		}

		if token == nil || !token.Valid {
			h.ServeHTTP(w, r)
			return nil
		}

		revoked, err := authtoken.IsRevoked(r.Context(), claims)
		if err != nil {
			return ErrValidateTokenIDFailed
		}

		if revoked {
			h.ServeHTTP(w, r)
			return nil
		}

		r = r.WithContext(authtoken.WithClaims(r.Context(), claims))
		h.ServeHTTP(w, r)

		return nil
	})
}

func BearerAuth(h http.Handler) http.Handler {
	return hdl.Middleware(func(w http.ResponseWriter, r *http.Request, meta map[string]any) error {
		claims := authtoken.GetClaims(r.Context())
		if claims == nil {
			return ErrUnauthorized
		}

		h.ServeHTTP(w, r)
		return nil
	})
}
