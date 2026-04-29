# 配置说明

默认配置文件为 `config.yml`，不存在时会自动生成。配置使用 YAML 格式，支持通过 `${ENV_VAR}` 引用环境变量。

默认模板即可 0 依赖启动：缓存使用 `local`，数据库使用 `sqlite3`，不需要提前配置 Redis 或 MySQL。

## 环境变量占位符

在配置中使用 `${WEBSITE_DESCRIPTION}` 这类语法，会读取同名环境变量的值。例如：

```yaml
machine_readable_resources:
	rss:
		description: ${WEBSITE_DESCRIPTION}
```

## 常用环境变量

- `WEBSITE_URL`：站点域名（用于 RSS / Sitemap）
- `WEBSITE_NAME`：站点名称
- `WEBSITE_DESCRIPTION`：站点描述
- `JWT_SECRET`：JWT 密钥
- `HTTP_RATE_API_KEY_1`：高配额密钥（可选）

## 主要配置项

### TLS

- `cert_file`：TLS 证书路径
- `key_file`：TLS 私钥路径

### 存储（storage）

- `storage.provider`：当前仅支持 `local`
- `storage.articles.base_path`：文章与资源文件存储路径

### 缓存（cache）

- `cache.provider`：`local` 或 `redis`
- `cache.local.dir`：本地缓存持久化路径
- `cache.redis.*`：Redis 连接信息与前缀

### 数据库（db）

- `db.driver`：`sqlite3` 或 `mysql`
- `db.dsn`：数据库连接字符串

默认值使用 SQLite：`file:./sys/database.db?_foreign_keys=on`。

### 鉴权（auth）

- `auth.jwt.secret`：JWT 密钥（建议配置为强随机）
- `auth.jwt.expiry`：Token 有效期（例如 `72h`）

### 接口限流（http_rate）

- `http_rate.limit_per_ip`：单 IP 限流规则
- `http_rate.black_list`：IP 黑名单（支持 CIDR）
- `http_rate.high_quota_keys`：高配额密钥列表（可绕过限流）

### 机器可读资源（machine_readable_resources）

- `base_url`：站点基础 URL
- `url_templates`：文章、分类、标签链接模板
- `rss`：RSS 开关与标题/描述
- `sitemap`：Sitemap 开关与扩展项

### 日志（log）

- `log.enable`：是否输出文件日志
- `log.log_level`：日志等级
- `log.lumberjack`：日志滚动参数

## 示例配置

项目自带模板位于 `pkg/assets/files/config.yml`，可作为参考。
