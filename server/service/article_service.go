package service

type ArticleService struct {
	ctx   *ServiceContext
	hooks *ArticleHooks
}

func NewArticleService(ctx *ServiceContext, hooks *ArticleHooks) *ArticleService {
	if hooks == nil {
		hooks = &ArticleHooks{}
	}
	return &ArticleService{ctx: ctx, hooks: hooks}
}
