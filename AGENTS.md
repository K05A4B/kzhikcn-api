# AGENTS.md

提供当前项目的关键架构和规范信息，用于 AI 在代码库中工作时保持上下文一致。

## Commands

```bash
# Build
go build -o .tmp/kzhikcn.exe .

# Run with hot-reload
air

# Run server directly (auto-creates default admin: admin/admin)
go run . -c config.yml serve

# Generate default config
go run . -c config.yml gen-config -d

# Run all tests
go test ./...

# Run single package tests
go test ./pkg/queryfilter/...
go test ./pkg/config/...
go test ./server/service/...

# Run single test
go test -run TestName ./server/service/

# Docker build
docker build -t kzhikcn-api .
```

## Architecture Overview

**Go 无头 CMS**：`go-chi` 路由，`GORM` ORM，`JWT + TOTP` 认证，`Badger/Redis` 缓存。

### 分层结构

```
main.go ──────────────────────────────── 入口
│
├─ internal/cli/ ─────────────────────  CLI (urfave/cli/v2)
│   ├─ serve/     serve 命令（启动 HTTP）
│   ├─ admin/     管理员管理（增删改查）
│   └─ gen_config 生成默认配置
│
├─ pkg/ ────────────────────────────── 共享库
│   ├─ hdl/            泛型 Handler 框架：自动解析 payload → 校验 → 处理 → JSON 响应
│   ├─ config/         YAML 配置加载（支持环境变量 ${VAR}、自定义 Duration/Size/Rate 类型）
│   ├─ data/           GORM 模型 + 纯 CRUD 数据访问（不含业务逻辑）
│   │   ├─ articles.go     Article 模型
│   │   ├─ categories.go   Category 模型
│   │   ├─ tags.go         Tag 模型 + GetTags
│   │   ├─ admins.go       Admin 模型（bcrypt + TOTP）
│   │   ├─ friend_links.go 友情链接（即将移除）
│   │   ├─ cache/          Cache 接口 + Badger/Redis 实现
│   │   └─ util.go         QueryModifier 函数式查询构建器
│   ├─ assets/         文章内容 + 附件存储抽象（本地文件系统）
│   │   ├─ article_repo.go   Init() + 全局 ArticlesRepo
│   │   └─ article/          Repository 接口 + LocalRepository 实现
│   ├─ queryfilter/    自定义查询表达式 DSL → SQL（支持 AND/OR/NOT/LIKE）
│   ├─ log/            zerolog 封装
│   └─ traceid/        请求追踪 ID
│
├─ server/ ─────────────────────────── HTTP + 业务逻辑
│   ├─ app/                App 生命周期编排
│   │   ├─ app.go          状态机：New → Bootstrap → Initialize → Migrate → Ready → Serve → Shutdown
│   │   └─ context.go      AppContext（嵌入 context.Context + ArticleService）
│   ├─ router/             路由树构建 + 全局中间件链
│   │   ├─ router.go        根路由（CORS → RealIP → TraceID → Recover → DotDotSlash → API）
│   │   └─ api_v1_router.go  /api/v1/ 所有路由定义
│   ├─ api/                Handler 层（仅做 HTTP 解析 + 调用 Service，不含业务逻辑）
│   │   ├─ ping.go          健康检查
│   │   ├─ machine_resources.go  Sitemap + RSS
│   │   ├─ articles/        文章 CRUD + 内容 + 资源
│   │   ├─ auth/            JWT 登录 / TOTP MFA / 登出
│   │   ├─ admin/           管理员信息、密码、MFA 管理
│   │   └─ topics/          分类与标签
│   ├─ common/
│   │   ├─ middlewares/     BearerAuth, BearerParse, CORS, HttpRate, Recover, TraceID, Cache, DotDotSlash
│   │   ├─ authtoken/      JWT 签发、解析、吊销（含缓存吊销列表）
│   │   ├─ articlemd/      goldmark 渲染 + Markdown AST 转换
│   │   ├─ httputil/       分页、排序、表达式查询、缓存控制工具
│   │   └─ machineresouces/ RSS + Sitemap 生成
│   └─ service/            Service 层（业务规则 + 跨模块协调）
│       ├─ service.go          ServiceContext + ArticleHooks + AuthHooks 定义
│       ├─ article_service.go  Article CRUD、内容管理、资产管理、计数
│       ├─ article_service_test.go  23 个测试用例（SQLite :memory:）
│       ├─ auth_service.go     骨架
│       ├─ admin_service.go    骨架
│       └─ topic_service.go    骨架
│
└─ internal/appinfo/ ── 从 app_info.json 嵌入版本信息
```

### 调用链

```
HTTP → middleware 链 → chi 路由 → hdl.New(handlerFactory(appCtx))
                                  │
                                  ├─ handler 解析请求参数
                                  ├─ handler 调用 appCtx.ArticleService.SomeMethod()
                                  │   └─ service 执行业务规则 + 触发 hooks
                                  │       └─ data 层纯 CRUD
                                  ├─ handler 映射错误码
                                  └─ handler 返回 JSON 响应
```

### 关键模式

**Handler 工厂函数** (`server/api/*/`):

Handler 定义为包级工厂函数，接收 `*app.AppContext` 返回 `hdl.Handler[T]`。路由注册时调用工厂函数得到 Handler，再传入 `hdl.New()` 适配为 `http.Handler`。

```go
// 定义
func CreateArticle(appCtx *app.AppContext) hdl.Handler[data.EditableArticle] {
    return hdl.NewHandler(
        func(r *http.Request, resp *hdl.Response, payload data.EditableArticle) error {
            article, err := appCtx.ArticleService.Create(r.Context(), payload)
            if err == data.ErrCategoryNotFound {
                return ErrCategoryNotFound
            }
            if err != nil {
                return ErrCreateArticleFailed.Wrap(err)
            }
            resp.Data = article
            return nil
        },
        hdl.MissingFields(func(payload data.EditableArticle) []string { ... }),
    )
}

// 注册
r.Post("/articles", hdl.New(articles.CreateArticle(appCtx)))
```

**`hdl` 框架** (`pkg/hdl/`):
- 大部分 API handler 通过 hdl 包创建
- 自动解析 JSON 请求体到泛型 T，串行执行 Validate → Handler
- 响应固定格式 `{success, code, message, data, meta, errorCode, traceId}`
- `hdl.WriteRaw()` / `hdl.WriteRawData()` 可接管响应，直接写原始数据流

**错误处理**：
- 业务错误使用 `hdl.DefineError(statusCode, message, errorCode)` 在包级别预定义
- Handler 内部通过预定义的 `ErrXxx.Wrap(err)` 包装下层错误
- Handler **不允许**直接返回 `data.*`、`gorm.*` 或其他原始 error，必须包装为 `*hdl.HandlerError`
- 无下层错误时直接返回预定义 error：`return ErrArticleNotFound`

**QueryModifier 模式** (`pkg/data/util.go`):
- 所有数据查询接受 `func(tx *gorm.DB) *gorm.DB` 函数式修饰符
- 通过 `ApplyQueryModifier` 链式调用
- `data.Pagination(page, limit)`、`data.LimitQueryModifier(n)` 等内置辅助

**Service 生命周期钩子** (`server/service/service.go`):

- `ArticleHooks` 提供 8 个钩子点：BeforeCreate、AfterCreate、BeforePublish、AfterPublish、BeforeDelete、AfterDelete、BeforeUpdateInfo、AfterUpdateInfo
- `AuthHooks` 提供 4 个钩子点：LoginSuccess、LoginFailed、MFARequired、Logout
- 插件通过注册钩子函数拦截业务操作

**自定义查询 DSL** (`pkg/queryfilter/`):
- 通过 `?expr=` 参数支持表达式查询，如 `title~go&status=published|views>100`
- 白名单机制控制可查询字段
- 解析为 AST → 转 SQL WHERE 子句

**App 生命周期** (`server/app/app.go`):
- 状态机：`New → Bootstrapped → Initialized → Migrated → Ready → Serving → Stopped`
- 每个阶段有对应的 Hook 方法（`HookAfterConfig`、`HookAfterDB` 等）

### 数据模型

- **Article**: UUID PK, 软删除, 状态 `published|draft|hidden`, belongsTo Category, many2many Tags
- **Category**: AutoIncrement PK, hasMany Articles, 唯一名称
- **Tag**: AutoIncrement PK, many2many Articles via `article_tags`
- **Admin**: bcrypt 密码, TOTP 双因素
- ~~**FriendLink / FriendLinkAudit**: 友情链接 + 审核流程（即将移除）~~

### API 路由

| 路由 | 认证 | 说明 |
|------|------|------|
| `/api/v1/ping` | 否 | 健康检查 |
| `/api/v1/auth/login` | 否 | 登录（支持 MFA 两阶段） |
| `/api/v1/auth/mfa/totp` | 否 | TOTP 验证 |
| `/api/v1/auth/logout` | 是 | 吊销 Token |
| `/api/v1/articles` | 公共读 | 文章列表（未认证仅 published） |
| `/api/v1/articles/{article_id}` | 公共读 | 文章详情（未认证仅 published/hidden） |
| `/api/v1/articles/{article_id}/view\|/like` | 否 | 浏览/点赞计数 |
| `/api/v1/articles/{article_id}/content` | 否 | 渲染后 Markdown 内容 |
| `/api/v1/articles/{article_id}/raw-content` | 是 | 原始 Markdown 内容 |
| `/api/v1/articles/{article_id}/assets` | 是 | 附件管理 |
| `/api/v1/categories\|/tags` | 否 | 分类/标签列表 |
| `/api/v1/categories/{category}/articles` | 否 | 按分类查文章 |
| `/api/v1/tags/{tag}/articles` | 否 | 按标签查文章 |
| `/sitemap.xml\|/rss.xml` | 否 | SEO 资源 |

### Service 层测试

- 使用 SQLite `:memory:` 数据库 + `TestMain` 统一初始化
- 每个测试调用 `cleanDB(t)` 清空四张表（article_tags, articles, tags, categories）
- 通过 `mockRepo`（实现 `article.Repository` 接口）mock 文件操作
- `config.LoadConfig` 首次初始化 config 防止 `log.GetLogger` 空指针
- 示例：`go test -v -run TestCreate_ValidArticle ./server/service/`

### 注意事项（必须遵守）

1. 注释需要避免使用无意义字符，例如`------` 或 `=====` 等