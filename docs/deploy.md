# 自定义部署指南

## 配置文件

运行服务前需准备配置文件。执行 `./kzhikcn serve` 时会自动创建默认配置文件，也可通过 `./kzhikcn gen-config` 命令生成。添加 `-d` 参数可生成完整的默认配置文件：

```bash
./kzhikcn gen-config -d
```

> [!note]
> 配置文件支持环境变量引用，格式为 `${环境变量名}`，例如 `${JWT_SECRET}` 表示读取环境变量 `JWT_SECRET` 的值。

> [!warning]
> 环境变量引用仅对字符串类型配置项生效。

默认配置文件内容如下：

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
    dir: ./sys/cache       # 本地缓存持久化目录

  # Redis 缓存配置（启用时需将 provider 设为 redis）
  redis:
    host: localhost        # Redis 服务地址
    port: 6379            # Redis 服务端口
    prefix: "kzhikcn:"    # 缓存键前缀，避免与其他应用冲突
    # timeout: 10s        # 连接超时时间
    # password: ""        # Redis 密码（如有）
    # db: 0               # 数据库编号


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
      # - { loc: /, last_modify: "{time}", change_freq: weekly, priority: 0.5 }
      # - { loc: /tags, change_freq: monthly }
      # - { loc: /categories, change_freq: monthly }

# 鉴权配置
auth:
  jwt:
    type: hs256                     # JWT 签名算法（仅支持 hs256）
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

## 启动服务

启动服务的基础命令如下：

```bash
./kzhikcn serve
```

默认使用 `./config.yml` 作为配置文件，监听 `0.0.0.0:5803` 端口。

### 指定配置文件

使用 `-c` 参数指定自定义配置文件路径：

```bash
./kzhikcn -c ./config.yml serve
```

### 指定监听地址

使用 `-a` 参数指定服务监听地址：

```bash
./kzhikcn serve -a 100.65.10.13:5803
```

### 组合使用

同时指定配置文件和监听地址：

```bash
./kzhikcn -c ./sys/config.yml serve -a 100.65.10.13:64435
```