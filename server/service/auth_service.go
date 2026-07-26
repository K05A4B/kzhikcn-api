package service

import (
	"context"
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/data/cache"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/common/secutils"
	"time"

	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
)

var (
	ErrAuthenticationFailed   = errors.New("authentication failed")
	ErrValidatePasswordFailed = errors.New("validate password failed")
	ErrGenerateTokenFailed    = errors.New("generate token failed")

	ErrChallengeMaxAttemptsExceeded = errors.New("challenge max attempts exceeded")
	ErrChallengeNotFound            = errors.New("challenge not found")
	ErrMFARequired                  = errors.New("mfa required")
	ErrChallengeExpired             = errors.New("challenge expired")

	challengeKey = cache.Keys("http", "auth", "mfa", "challenges")
)

type MFAChallenge struct {
	duration time.Duration

	ID          string
	Username    string
	UserID      uint
	MaxAttempts int
	Expire      time.Time
}

func (c *MFAChallenge) IsExpired() bool {
	return c.Expire.Before(time.Now())
}

func (c *MFAChallenge) Commit(ctx context.Context) error {
	return cache.SetJson(ctx, cache.Keys(challengeKey, c.ID), c, time.Until(c.Expire))
}

func (c *MFAChallenge) Delete(ctx context.Context) error {
	return cache.Delete(ctx, cache.Keys(challengeKey, c.ID))
}

type AuthService struct {
	ctx   *ServiceContext
	hooks *AuthHooks
}

func NewAuthService(ctx *ServiceContext, hooks *AuthHooks) *AuthService {
	if hooks == nil {
		hooks = &AuthHooks{}
	}
	return &AuthService{ctx: ctx, hooks: hooks}
}

func (a *AuthService) CreateChallenge(ctx context.Context, username string, userId uint) (*MFAChallenge, error) {
	challenge := &MFAChallenge{
		duration: 2 * time.Minute,

		ID:          utils.RandomString(24),
		Username:    username,
		UserID:      userId,
		MaxAttempts: 5,
		Expire:      time.Now().Add(2 * time.Minute),
	}

	return challenge, challenge.Commit(ctx)
}

func (a *AuthService) GetChallenge(ctx context.Context, challengeId string) (*MFAChallenge, error) {
	challenge := &MFAChallenge{}
	key := cache.Keys(challengeKey, challengeId)

	exists, err := cache.Exists(ctx, key)
	if !exists {
		return nil, ErrChallengeNotFound
	}

	err = cache.GetJson(ctx, key, challenge)
	return challenge, err
}

// 登录管理员账号
// 可能返回的错误：
//   - ErrValidatePasswordFailed: 密码验证时出错
//   - ErrGenerateTokenFailed: 生成token失败
//   - ErrMFARequired: 需要进行多因素认证
//   - ErrAuthenticationFailed: 认证失败
func (a *AuthService) AdminLogin(ctx context.Context, admin *data.Admin, password string) (token string, err error) {
	if admin == nil {
		return "", ErrAuthenticationFailed
	}

	ok, err := secutils.ComparePassword(admin.Password, password)
	if err != nil {
		a.hooks.triggerFailed(ctx, admin.Username, "validate password failed")
		return "", errors.Wrap(ErrValidatePasswordFailed, err.Error())
	}

	if !ok {
		a.hooks.triggerFailed(ctx, admin.Username, "password not match")
		return "", ErrAuthenticationFailed
	}

	if admin.EnableMFA {
		return "", ErrMFARequired
	}

	token, err = authtoken.IssueToken(admin.ID, admin.Username, true)
	if err != nil {
		a.hooks.triggerFailed(ctx, admin.Username, "generate token failed")
		return "", errors.Wrap(ErrGenerateTokenFailed, err.Error())
	}

	a.hooks.triggerSuccess(ctx, admin)

	return token, nil
}

func (a *AuthService) Logout(ctx context.Context, claims *authtoken.TokenClaims) error {
	a.hooks.triggerLogout(ctx, claims)

	return authtoken.RevokeToken(ctx, claims)
}

func (a *AuthService) TOTPValidate(ctx context.Context, secret string, challenge *MFAChallenge, code string) error {
	if challenge == nil {
		return errors.New("challenge cannot be nil")
	}

	if challenge.IsExpired() {
		return ErrChallengeExpired
	}

	if challenge.MaxAttempts <= 0 {
		return ErrChallengeMaxAttemptsExceeded
	}

	challenge.MaxAttempts--

	err := challenge.Commit(ctx)
	if err != nil {
		// 如果保存失败，则直接删除挑战
		challenge.Delete(ctx)
		return err
	}

	if !totp.Validate(code, secret) {
		return ErrAuthenticationFailed
	}

	return nil
}
