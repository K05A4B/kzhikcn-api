package authtoken

import (
	"net/http"
	"strings"
)

func GetTokenString(r *http.Request) string {
	authFields := strings.Fields(r.Header.Get("Authorization"))
	if len(authFields) < 2 {
		return ""
	}

	if strings.ToLower(authFields[0]) != "bearer" {
		return ""
	}

	return authFields[1]
}
