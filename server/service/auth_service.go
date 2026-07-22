package service

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
