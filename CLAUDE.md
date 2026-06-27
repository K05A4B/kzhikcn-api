# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build -o .tmp/kzhikcn.exe .

# Run with hot-reload (uses .air.toml)
air

# Run server directly (auto-creates default admin: admin/admin)
go run . -c config.yml serve

# Generate default config
go run . -c config.yml gen-config -d

# Run tests
go test ./...

# Run single package tests
go test ./pkg/queryfilter/...
go test ./pkg/config/...

# Run single test
go test -run TestName ./pkg/config/

# Docker build
docker build -t kzhikcn-api .
```

## Architecture Overview

这是一个**Go 无头 CMS 后端**，使用 `go-chi` 做路由、`GORM` 做 ORM、`JWT + TOTP` 做认证、`Badger/Redis` 做缓存。

### 分层结构

```
app.go ──────────────────────────────── 入口
│
├─ internal/cli/ ─────────────────────  CLI (urfave/cli/v2)
│   ├─ serve/     serve 命令（启动 HTTP）
│   ├─ admin/     管理员管理（增删改查）
│   └─ gen_config 生成默认配置
│
├─ server/ ─────────────────────────── HTTP 层
│   ├─ server.go       启动 http.ListenAndServe[TLS]
│   ├─ router.go       根路由（traceID → recover → API）
│   ├─ api/
│   │   ├─ router.go         /api/ 挂载点（BearerParse → HttpRate → CacheControl）
│   │   ├─ api_v1_router.go  /api/v1/ 路由定义
│   │   ├─ articles/    文章 CRUD + 资源上传下载
│   │   ├─ auth/        JWT 登录 / TOTP MFA / 登出
│   │   ├─ admin/       管理员信息、修改密码、MFA 管理
│   │   └─ topics/      分类与标签管理
│   ├─ common/
│   │   ├─ hdl/             泛型 Handler 框架：自动解析 payload → 校验 → 处理 → JSON 响应
│   │   ├─ middlewares/     BearerAuth, BearerParse, HttpRate, Recover, TraceID, Cache, DotDotSlash
│   │   ├─ authtoken/      JWT 签发、解析、吊销
│   │   ├─ articlemd/      goldmark 渲染 + Markdown AST 转换
│   │   ├─ httputil/       分页、排序、表达式查询、缓存控制工具
│   │   └─ machineresouces/ RSS + Sitemap 生成
│   └─ service/articles.go  业务逻辑层（目前仅占位）
│
├─ pkg/ ────────────────────────────── 共享库
│   ├─ config/         YAML 配置加载（支持环境变量 ${VAR}）
│   ├─ data/           GORM 模型 + 数据访问
│   │   ├─ articles.go     Article 模型 + CRUD + 软删除/恢复
│   │   ├─ categories.go   Category 模型
│   │   ├─ tags.go         Tag 模型
│   │   ├─ admins.go       Admin 模型（bcrypt + TOTP）
│   │   ├─ friend_links.go 友情链接
│   │   ├─ cache/          Cache 接口 + Badger/Redis 实现
│   │   └─ util.go         QueryModifier 函数式查询构建器
│   ├─ assets/         文章内容 + 附件存储抽象（本地文件系统）
│   ├─ queryfilter/    自定义查询表达式 DSL → SQL（支持 AND/OR/NOT/LIKE）
│   ├─ log/            zerolog 封装
│   └─ traceid/        请求追踪 ID
│
└─ internal/appinfo/ ── 从 app_info.json 嵌入版本信息
```

### 关键模式

**Handler 框架** (`server/common/hdl/`):
- 所有 API handler 通过 `hdl.New[T]()` 或 `hdl.NewSimpleHandler()` 创建
- 自动解析 JSON 请求体到泛型 T，串行执行 Validate → Handler
- 响应固定格式 `{success, code, message, data, meta, errorCode, traceId}`
- `hdl.Error()` 和 `hdl.DefineError()` 创建结构化错误，含 HTTP status、内部 error、errorCode
- `hdl.WriteRaw()` / `hdl.WriteRawData()` 可接管响应，直接写原始数据

**QueryModifier 模式** (`pkg/data/util.go`):
- 所有数据查询接受 `func(tx *gorm.DB) *gorm.DB` 函数式修饰符
- 通过 `ApplyQueryModifier` 链式调用，类似中间件
- 分页、排序、WHERE 条件都以此方式注入

**自定义查询 DSL** (`pkg/queryfilter/`):
- 通过 `?expr=` 参数支持表达式查询，如 `title~go&status=published|views>100`
- 有白名单机制控制可查询字段
- 解析为 AST → 转 SQL WHERE 子句

**配置钩子** (`pkg/config/config.go`):
- `config.HookLoaded()` 注册回调，配置加载后自动执行
- 用于解耦初始化顺序（如 assets 模块加载后初始化文章仓库）

### 数据模型

- **Article**: UUID PK, 软删除, 状态 `published|draft|hidden`, belongsTo Category, many2many Tags, 自定义 ID (CustomID)
- **Category**: AutoIncrement PK, hasMany Articles, 唯一名称
- **Tag**: AutoIncrement PK, many2many Articles via `article_tags`
- **Admin**: bcrypt 密码, TOTP 双因素, 支持多管理员
- **FriendLink / FriendLinkAudit**: 友情链接 + 审核流程 (pending/approved/rejected)

### API 路由分类

| 路由 | 认证 | 说明 |
|------|------|------|
| `/api/v1/ping` | 否 | 健康检查 |
| `/api/v1/auth/login` | 否 | 登录（支持 MFA 两阶段） |
| `/api/v1/auth/mfa/totp` | 否 | TOTP 验证 |
| `/api/v1/articles` | 公共读 | 文章列表（未认证仅 published） |
| `/api/v1/articles/{id}` | 公共读 | 文章详情 |
| `/api/v1/articles/{id}/view\|/like` | 否 | 浏览/点赞计数 |
| `/api/v1/categories\|/tags` | 否 | 分类/标签列表 |
| `/api/v1/auth/logout` | 是 | 吊销 Token |
| `/api/v1/articles` (POST/DELETE/PATCH) | 是 | 文章管理 |
| `/api/v1/articles/{id}/raw-content` | 是 | 文章内容读写 |
| `/api/v1/articles/{id}/assets` | 是 | 附件管理 |
| `/sitemap.xml\|/rss.xml` | 否 | SEO 资源 |
| `/api/v1/tags\|/categories` (管理) | 是 | 分类/标签 CRUD |

### 缓存

- 接口定义：`Set/Get/Delete/Exists`，支持 TTL
- 实现：本地 Badger（嵌入式 LSM-tree）或 Redis
- 用途：JWT 吊销列表、MFA challenge、可扩展业务缓存
- 缓存键格式：`cache.Keys("http", "auth", ...)` → `"http:auth:..."`

### 认证流程

1. `POST /auth/login` → 验证密码 → 返回 token 或 `needMFA`
2. MFA 时：生成 2 分钟有效期的 challenge → `POST /auth/mfa/totp` 验证 → 返回 token
3. Token 放在 `Authorization: Bearer <token>` 头
4. `Bearparse` 中间件（放行而非拦截）将 claims 放入 context
5. `BearerAuth` 中间件（拦截未认证）检查 claims 是否存在
6. 登出时将 token ID 加入缓存吊销列表

### 文章存储

- 每篇文章在磁盘上以 `{base_path}/{article_id}/` 目录存储
- `index.md` — 原始 Markdown 内容
- `assets/` — 上传的附件文件
- 内容读写使用带锁的 Reader/Writer（防止并发写）
- 支持 goldmark 渲染（表格、GFM、任务列表、懒加载图片）
