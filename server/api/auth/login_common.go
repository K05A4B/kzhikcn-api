package auth

import (
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/data/cache"
	"kzhikcn/server/common/authtoken"
	"net/http"
	"time"
)

type challenge struct {
	Username    string
	UserId      uint
	MaxAttempts int
	Expire      time.Time
}

var mfaChallengesKey = cache.Keys("http", "auth", "mfaChallenges")

func permitLogin(r *http.Request, admin *data.Admin) (string, error) {
	return authtoken.IssueToken(admin.ID, admin.Username)
}
