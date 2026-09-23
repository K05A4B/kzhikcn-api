# 配置文件参考

服务的运行行为通过 YAML 配置文件控制。本文档列出全部配置项，其中事件相关配置见[事件配置](#事件配置)。

## 概述

- 默认读取工作目录下的 `config.yml`，可通过全局选项 `-c` 指定其它路径。
- 配置文件不存在时，`serve` 会自动生成默认配置；也可执行 `./kzhikcn gen-config -d` 生成完整默认配置。
- 执行 `./kzhikcn -c ./config.yml config check` 可校验配置，`config show` 可查看生效配置（敏感字段已脱敏）。

### 环境变量引用

配置项支持环境变量引用，格式为 `${环境变量名}`，例如 `${JWT_SECRET}`。

> [!NOTE]
> 环境变量引用对**所有字符串类型配置项**生效，包括列表中的字符串项和 `events[].entry`。

> [!WARNING]
> 若引用的环境变量未设置，占位符会**原样保留**，`config check` 会将其判定为错误。

### 值格式

| 类型 | 说明 | 示例 |
| :--- | :--- | :--- |
| 时长 (duration) | Go 时长字符串 | `72h`、`5s`、`600ms` |
| 大小 (size) | 数字 + 单位（`b` / `kb` / `mb` / `gb`） | `64mb`、`8mb` |
| 限流速率 | `次数/单位`，单位可取 `ms` / `s` / `m` / `h`（默认 `s`） | `100/s`、`6000/m` |

## 顶层配置项

| 字段 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `cert_file` | 字符串 | 空 | TLS 证书路径，与 `key_file` 同时设置时启用 HTTPS |
| `key_file` | 字符串 | 空 | TLS 私钥路径 |
| `storage` | 对象 | - | 存储配置 |
| `cache` | 对象 | - | 缓存配置 |
| `event_dispatcher` | 对象 | - | 事件分发器配置 |
| `events` | 列表 | 空 | 事件订阅列表 |
| `machine_readable_resources` | 对象 | - | 机器可读资源（RSS / Sitemap）配置 |
| `auth` | 对象 | - | 鉴权配置 |
| `http_rate` | 对象 | - | 接口限流配置 |
| `db` | 对象 | - | 数据库配置 |
| `log` | 对象 | - | 日志配置 |
| `cors` | 对象 | - | 跨域配置 |

## storage

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `storage.provider` | 字符串 | 存储提供商，当前仅支持 `local` |
| `storage.articles.base_path` | 字符串 | 文章内容与附件的存储根目录 |

```yaml
storage:
  provider: local
  articles:
    base_path: ./data/articles
```

## cache

| 字段 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `cache.provider` | 字符串 | - | 缓存类型：`local`（内置缓存）或 `redis` |
| `cache.local.dir` | 字符串 | - | 本地缓存（Badger）持久化目录 |
| `cache.local.value_log_file_size` | 大小 | `64mb` | 单个值日志文件大小 |
| `cache.local.mem_table_size` | 大小 | `8mb` | 内存表大小 |
| `cache.redis.addr` | 字符串 | - | Redis 地址，格式 `host:port` |
| `cache.redis.prefix` | 字符串 | - | 缓存键前缀，避免与其他应用冲突 |
| `cache.redis.username` | 字符串 | 空 | Redis 用户名（Redis 6+ ACL） |
| `cache.redis.password` | 字符串 | 空 | Redis 密码 |
| `cache.redis.db` | 整数 | `0` | 数据库编号 |
| `cache.redis.timeout` | 时长 | - | 连接超时 |

> [!NOTE]
> 使用 `local` 缓存时目录为**独占锁**，同一目录同时只能被一个进程打开。只访问数据库的 CLI 命令（如 `admin`、`migrate`）不会初始化缓存，因此可在服务运行期间安全执行。

## 事件配置

事件系统允许在特定生命周期节点触发外部动作。每个事件可通过 `webhook` 发起 HTTP 请求，或通过 `command` 执行本地命令。默认为**异步执行**。

### event_dispatcher

事件分发器的超时与异步执行池配置。

| 字段 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `timeout` | 时长 | `10s` | 单次事件分发超时，同时作用于 webhook 的 HTTP 请求与 command 的执行时长；未配置或非正值时使用 `10s` |
| `workers` | 整数 | `4` | 异步执行池的 worker 数量；非正值时使用默认值。**仅启动时生效** |
| `queue_size` | 整数 | `256` | 异步执行池的队列长度；非正值时使用默认值。**仅启动时生效** |

> [!NOTE]
> `workers` 与 `queue_size` 仅在服务启动时读取。热重载配置时若发生变化，会记录警告提示需重启服务。

### events

`events` 为事件订阅列表，每项配置一个事件：

| 字段 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | 字符串 | 否 | 事件名称，仅用于标识与日志 |
| `on` | 字符串 | 是 | 触发时机，取值见[触发时机](#触发时机) |
| `type` | 字符串 | 是 | 执行类型：`webhook` 或 `command` |
| `entry` | 字符串 | 是 | 执行入口：webhook 为 URL，command 为命令行 |
| `async` | 布尔 | 否 | 是否异步执行，默认 `true`；设为 `false` 强制同步。`app.*` 事件始终同步，此项对其无效 |
| `timeout` | 时长 | 否 | 覆盖 `event_dispatcher.timeout`，对 webhook 与 command 均生效 |
| `headers` | 对象 | 否 | webhook 请求附带的 HTTP 头，仅对 `type: webhook` 生效；值支持 `${环境变量}` 引用 |

### 触发时机

`on` 使用「命名空间.动作」格式，当前支持的取值及 `data` 载荷如下：

| `on` | 触发点 | `data` 载荷 |
| :--- | :--- | :--- |
| `app.after_db` | 数据库初始化完成 | `null` |
| `app.after_cache` | 缓存初始化完成 | `null` |
| `app.after_storage` | 存储初始化完成 | `null` |
| `app.migration` | 数据库迁移完成 | `null` |
| `app.before_serve` | 开始监听端口前 | `null` |
| `app.shutdown` | 服务关闭时 | `null` |
| `article.creating` | 创建文章前 | 待创建的文章字段 |
| `article.created` | 创建文章后 | 文章对象 |
| `article.updating` | 更新文章信息前 | 文章对象 |
| `article.updated` | 更新文章信息后 | 文章对象 |
| `article.publishing` | 文章发布前 | 文章对象 |
| `article.published` | 文章发布后 | 文章对象 |
| `article.deleting` | 删除文章前 | `{ "articleID": "...", "isHard": true }` |
| `article.deleted` | 删除文章后 | `{ "articleID": "...", "isHard": true }` |
| `auth.login_success` | 登录成功（含 MFA 校验通过并签发 token） | 管理员对象（不含密码与 TOTP 密钥） |
| `auth.login_failed` | 登录失败（用户名不存在、密码错误、MFA 验证码错误等） | `{ "username": "...", "reason": "..." }` |
| `auth.logout` | 登出 | JWT claims（含 `adminId`、`is_admin`） |

> [!NOTE]
> `article.publishing` / `article.published` 仅在通过「更新文章信息」把状态变更为 `published` 时触发。若在「创建文章」时直接把 `status` 设为 `published`，只会触发 `article.created`，**不会**触发发布相关事件。

> [!NOTE]
> `auth.login_success` 在登录**完整成功**后触发：未启用 MFA 时在签发 token 后触发，启用 MFA 时在 TOTP 校验通过并签发 token 后触发。`auth.login_failed` 在用户名不存在、密码错误、MFA 验证码错误、签发 token 失败等场景触发。

### webhook

向 `entry` 指定的 URL 发起 `POST` 请求：

- 默认请求头为 `Content-Type: application/json` 与 `User-Agent: kzhikcn-api/<version>`，均可通过 `headers` 覆盖。
- `headers` 中的键值会写入请求头；其中 `Host` 会生效。
- 请求体为统一的事件信封：

  ```json
  {
    "name": "Article Created Hook",
    "on": "article.created",
    "data": { }
  }
  ```

- 响应状态码为 `2xx` 视为成功，其它状态码视为失败。
- HTTP 请求受 `event_dispatcher.timeout` 约束，事件级 `timeout` 可覆盖。

### command

通过系统 shell 执行 `entry`：

- Windows 使用 `cmd /c <entry>`，Linux / macOS 使用 `sh -c <entry>`。
- 事件信封（与 webhook 相同的 JSON）写入命令的 **标准输入**。
- 同时注入环境变量 `EVENT_NAME`、`EVENT_ON`。
- 命令退出码非 `0` 视为失败。
- 命令执行受 `event_dispatcher.timeout` 约束，事件级 `timeout` 可覆盖。
- `headers` 对 command 无效（命令无 HTTP 头概念）。

### 执行语义

- **默认异步**：事件在业务请求返回后由后台执行池处理，不增加接口延迟；`async: false` 可让单个事件同步执行（请求会等待其完成）。
- **`app.*` 强制同步**：生命周期事件（`app.after_db`、`app.shutdown` 等）时机敏感，始终同步执行，配置 `async: true` 也不会生效。
- **有界执行池**：异步事件投递到固定大小的队列（默认 4 个 worker、256 队列长度，可通过 `event_dispatcher.workers` / `event_dispatcher.queue_size` 调整）。队列满时事件被丢弃并记录告警，**不会阻塞**请求路径。
- **载荷快照**：异步事件在派发瞬间对 `data` 做序列化快照，请求后续对对象的修改不会影响已派发的事件。
- **关闭排空**：服务关闭时会等待在途异步事件完成，等待时间受关闭超时约束；超时未完成的事件可能丢失。
- **best-effort**：任一事件失败仅记录日志，**不会阻断**触发它的业务请求（例如登录、创建文章仍会正常返回成功）。
- **顺序**：同步事件在同一 `on` 下按配置顺序依次执行；异步事件的完成顺序不保证。
- `type` 不在 `webhook` / `command` 中时，会记录日志并跳过该事件。

### 示例

```yaml
event_dispatcher:
  timeout: 5s
  workers: 4
  queue_size: 256

events:
  # webhook：文章创建后回调（默认异步），自定义超时与请求头
  - name: Article Created
    on: article.created
    type: webhook
    entry: http://127.0.0.1:8080/hooks/article-created
    timeout: 30s
    headers:
      Authorization: "Bearer ${HOOK_TOKEN}"
      User-Agent: my-hook-client/1.0

  # command：登录成功后执行脚本，改为同步执行
  - name: Login Success
    on: auth.login_success
    type: command
    entry: python /opt/hooks/on_login.py
    async: false
```

`on_login.py` 可从标准输入读取事件信封，或直接读取环境变量 `EVENT_NAME` / `EVENT_ON`。

> [!WARNING]
> `entry` 来自受信任的配置文件。`command` 类型会以服务进程权限执行任意 shell 命令，请确保配置文件权限受到保护。

> [!NOTE]
> 事件载荷中的敏感字段（如管理员的密码哈希、TOTP 密钥）不会序列化输出。

> [!IMPORTANT]
> 旧版顶层配置项 `event_timeout` 已移除，请改用 `event_dispatcher.timeout`。旧配置中的 `event_timeout` 会被静默忽略并退回默认超时。

## machine_readable_resources

用于生成 RSS 与 Sitemap。

| 字段 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `base_url` | 字符串 | - | 站点基础 URL，会拼接在路径模板前 |
| `url_templates.article` | 字符串 | - | 文章路径模板 |
| `rss.enable` | 布尔 | `false` | 是否启用 RSS |
| `rss.title` | 字符串 | - | RSS 标题 |
| `rss.description` | 字符串 | - | RSS 描述 |
| `rss.max_articles` | 整数 | `0` | RSS 最大文章数 |
| `sitemap.enable` | 布尔 | `false` | 是否启用 Sitemap |
| `sitemap.extends` | 列表 | 空 | 额外的 Sitemap 条目 |

`url_templates.article` 支持占位符 `{{article.id}}`（文章 UUID）与 `{{article.custom_id}}`（文章自定义 ID）：

```yaml
machine_readable_resources:
  base_url: https://example.com
  url_templates:
    article: /articles/{{article.custom_id}}
```

> [!NOTE]
> 目前 RSS 与 Sitemap 仅使用 `url_templates.article`，分类与标签路径模板为预留配置。

### sitemap.extends

每个条目描述一个额外的 Sitemap URL：

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `loc` | 字符串 | 路径，会自动拼接 `base_url` |
| `last_mod` | 字符串 | 最后修改时间，支持 `{time}` 占位符 |
| `change_freq` | 字符串 | 更新频率，如 `daily`、`weekly`、`monthly` |
| `priority` | 整数 | 优先级 |

```yaml
  sitemap:
    enable: true
    extends:
      - { loc: / }
      - { loc: /, last_mod: "{time}", change_freq: weekly }
      - { loc: /tags, change_freq: monthly }
```

## auth

| 字段 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `auth.jwt.secret` | 字符串 | - | JWT 密钥，建议通过环境变量配置 |
| `auth.jwt.expiry` | 时长 | `72h` | Token 有效期 |

```yaml
auth:
  jwt:
    expiry: 72h
    secret: ${JWT_SECRET}
```

## http_rate

接口限流配置。JWT 认证通过或携带高配额密钥的请求不受限流限制。

| 字段 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `http_rate.limit_per_ip` | 限流速率 | - | 单 IP 限流速率 |
| `http_rate.black_list` | 列表 | 空 | IP 黑名单，支持单个 IP 与 CIDR，按顺序匹配 |
| `http_rate.high_quota_keys` | 列表 | 空 | 高配额密钥，请求头 `X-High-Quota-Key` 匹配时绕过限流 |

```yaml
http_rate:
  limit_per_ip: 100/s
  black_list:
    - 192.168.1.1/32
    - 10.1.1.0/24
  high_quota_keys:
    - ${HTTP_RATE_API_KEY_1}
```

## cors

| 字段 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `cors.enable` | 布尔 | `false` | 是否启用 CORS |
| `cors.allowed_origins` | 列表 | 空 | 允许的来源，支持具体域名或 `*` |
| `cors.allowed_methods` | 列表 | 空 | 允许的 HTTP 方法，留空时为 `GET, POST, PUT, DELETE, PATCH, OPTIONS` |
| `cors.allowed_headers` | 列表 | 空 | 允许的请求头，留空时为 `Content-Type, Authorization, X-Requested-With` |
| `cors.exposed_headers` | 列表 | 空 | 允许前端访问的响应头 |
| `cors.max_age` | 时长 | - | 预检请求结果缓存时间 |
| `cors.allow_credentials` | 布尔 | `false` | 是否允许携带凭证；为 `true` 时 `allowed_origins` 不能为 `*` |

## db

| 字段 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `db.driver` | 字符串 | - | 数据库驱动，目前支持 `sqlite3`（MySQL 未测试） |
| `db.dsn` | 字符串 | - | 数据库连接地址 |

```yaml
db:
  driver: sqlite3
  dsn: file:./sys/database.db?_foreign_keys=on
```

## log

| 字段 | 类型 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `log.enable` | 布尔 | `false` | 是否启用文件日志 |
| `log.log_level` | 字符串 | - | 日志级别：`debug` / `info` / `warn` / `error` |
| `log.lumberjack.filename` | 字符串 | - | 日志文件路径 |
| `log.lumberjack.maxsize` | 整数 | - | 单个日志文件最大大小（MB） |
| `log.lumberjack.maxage` | 整数 | - | 日志保留天数 |
| `log.lumberjack.maxbackups` | 整数 | - | 最大备份文件数 |
| `log.lumberjack.compress` | 布尔 | - | 是否压缩旧日志 |

```yaml
log:
  enable: true
  log_level: info
  lumberjack:
    filename: ./sys/logs/latest.log
    maxsize: 10
    maxage: 30
    maxbackups: 3
    compress: true
```

## 完整默认配置

以下为 `gen-config -d` 生成的完整默认配置：

```yaml
# TLS 证书配置
# cert_file: /sys/tls/cert.pem
# key_file: /sys/tls/key.pem

# 存储配置
storage:
  provider: local          # 存储提供商，当前仅支持 local（本地存储）
  articles:
    base_path: ./data/articles  # 文章数据存储基础路径

# 缓存配置
cache:
  provider: local          # 缓存类型：local（内置缓存）或 redis
  local:
    dir: ./sys/cache          # 本地缓存持久化目录
    value_log_file_size: 64mb # 单个值日志文件大小 (默认 64mb)
    mem_table_size: 8mb       # 内存表大小 (默认 8mb)

  # Redis 缓存配置（启用时需将 provider 设为 redis）
  redis:
    addr: localhost:6379   # Redis 服务地址 (host:port)
    prefix: "kzhikcn:"     # 缓存键前缀，避免与其他应用冲突
    # username: ""         # Redis 用户名（Redis 6+ ACL）
    # timeout: 10s         # 连接超时时间
    # password: ""         # Redis 密码（如有）
    # db: 0                # 数据库编号

# 事件分发配置
event_dispatcher:
  timeout: 5s      # 单次事件分发超时时间
  workers: 4       # 异步事件 worker 数量（仅启动时生效）
  queue_size: 256  # 异步事件队列长度（仅启动时生效）

# 事件订阅列表
events:
  # 事件默认异步执行；async: false 可改为同步（app.* 事件始终同步）
  # timeout 覆盖 event_dispatcher.timeout；headers 仅对 webhook 生效
  # - name: Article Created
  #   on: article.created
  #   type: webhook
  #   entry: http://your-webhook-url.lab/your/webhook
  #   async: true
  #   timeout: 30s
  #   headers:
  #     Authorization: "Bearer ${WEBHOOK_TOKEN}"

  # - name: Article Updated
  #   on: article.updated
  #   type: command
  #   entry: python /your/python/script.py
  #   async: false


# 机器可读资源配置
machine_readable_resources:
  base_url: ${WEBSITE_URL}        # 网站基础 URL

  url_templates:
    article: /articles/{{article.id}}      # 文章路径模板（{{article.custom_id}} 会被替换为文章自定义id）
    category: /categories/{{category.name}} # 分类路径模板（{{category.id}} 会被替换为分类id）
    tag: /tags/{{tag.name}}                # 标签路径模板（{{tag.id}} 会被替换为标签id）

  rss:
    enable: true                    # 是否启用 RSS 订阅
    title: ${WEBSITE_NAME}          # RSS 标题
    description: ${WEBSITE_DESCRIPTION}    # RSS 描述
    max_articles: 10                # RSS 最大文章数量

  sitemap:
    enable: true                    # 是否启用 Sitemap
    extends:                        # 扩展 Sitemap 条目
      - { loc: / }
      # - { loc: /, last_mod: "{time}", change_freq: weekly }
      # - { loc: /tags, change_freq: monthly }
      # - { loc: /categories, change_freq: monthly }

# 鉴权配置
auth:
  jwt:
    expiry: 72h                     # Token 有效期（72 小时）
    secret: ${JWT_SECRET}           # JWT 密钥（建议通过环境变量配置）

# 接口限流配置
# 注意：JWT 认证通过或携带高配额密钥的请求不受限流限制
http_rate:
  limit_per_ip: 100/s               # 单 IP 限流速率（每秒 100 次）
  # limit_per_ip: 6000/m            # 每分钟 6000 次
  # limit_per_ip: 360000/h          # 每小时 360000 次

  # IP 黑名单（支持 CIDR 格式，按顺序匹配）
  black_list: [
    # 192.168.1.1/32,    # 封禁（192.168.1.1）单个 IP
    # 192.168.1.2,       # 封禁（192.168.1.2）单个 IP
    # 10.1.1.0/24,       # 封禁（10.1.1.0/24）整个网段
    # 0.0.0.0/0,         # 封禁（0.0.0.0/0）所有 IPv4 地址
    # '::/0',            # 封禁（::/0）所有 IPv6 地址
    # 2000::/3           # 封禁（2000::/3） IPv6 全球单播地址
  ]

  # 高配额密钥（请求头 X-High-Quota-Key 匹配时绕过限流）
  high_quota_keys: [
    # "${HTTP_RATE_API_KEY_1}"
  ]

cors:
  enable: false

  # Access-Control-Allow-Origin
  # 指定哪些域可以访问资源。值可以是具体的域名（如 https://example.com）或 *（允许所有域）
  allowed_origins: [
    # "https://*.example.com",
    # "http://*.example.com",
    # "http://*.cn",
    # "http://foo.bar.example"
  ]

  # Access-Control-Allow-Methods
  # 指定允许的 HTTP 方法，留空时默认为 GET, POST, PUT, DELETE, PATCH, OPTIONS
  allowed_methods: [
    # "GET",
    # "POST",
    # "PUT",
    # "DELETE",
    # "PATCH",
    # "OPTIONS"
  ]

  # Access-Control-Allow-Headers
  # 指定实际请求中允许携带的自定义头部字段（如 X-Requested-With, Content-Type）
  allowed_headers: [
    # "Accept",
    # "Authorization",
    # "Content-Type",
    # "X-CSRF-Token"
  ]

  # Access-Control-Expose-Headers
  # 默认情况下，浏览器只允许前端访问少数基础响应头（如 Cache-Control）。此头部用于列出允许前端 JavaScript 访问的额外响应头（如 Authorization）。
  exposed_headers: [
    # "Authorization"
  ]

  # Access-Control-Max-Age
  # 指定预检请求（Preflight，即 OPTIONS 请求）的结果可以被缓存多长时间，以减少重复的预检请求。
  max_age: 600s

  # Access-Control-Allow-Credentials
  # 指定是否允许浏览器携带凭证信息（如 Cookies、HTTP 认证信息）。值为 true 时，Allow-Origin 不能设为 *，必须指定具体域名。
  allow_credentials: false

# 数据库配置
db:
  # 数据库驱动目前支持 sqlite3（mysql未测试）
  driver: sqlite3
  # 数据库连接地址
  dsn: file:./sys/database.db?_foreign_keys=on

# 日志配置
log:
  enable: true                      # 是否启用文件日志
  log_level: info                   # 日志级别

  # 日志轮换配置（基于 lumberjack）
  lumberjack:
    filename: ./sys/logs/latest.log # 日志文件路径
    maxsize: 10                     # 单个日志文件最大大小（MB）
    maxage: 30                      # 日志保留天数
    maxbackups: 3                   # 最大备份文件数
    compress: true                  # 是否压缩旧日志
```
