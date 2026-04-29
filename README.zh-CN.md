# kzhikcn-api

kzhikcn-api 是一个基于 Go 的轻量级无头 CMS 后端，围绕 Markdown 内容、清晰的管理流程以及可预测的部署方式构建。

## 核心特点

- 文章生命周期：draft/hidden/published，支持软删除与恢复
- Markdown 正文与资源文件管理
- 分类与标签体系，支持表达式筛选
- JWT 鉴权 + 可选 TOTP 多因素认证
- API 限流与高配额密钥绕过
- RSS 与 Sitemap 生成
- 本地存储与缓存（可选 Redis）
- 0 依赖启动（默认本地 SQLite + 本地缓存）

## 环境要求

- Go 1.25+（toolchain 1.26.2）
- SQLite3（默认）或 MySQL
- 可选：Redis（缓存）

## 快速开始（本地）

1. 生成配置：`go run . gen-config`
2. 启动服务：`go run . serve -a 0.0.0.0:5083`
3. 首次启动会自动迁移数据库并创建默认管理员账号（`admin` / `admin`），请尽快修改密码。

也可以直接运行 `serve`。如果没有 `config.yml`，会基于默认模板自动生成，但不会交互式询问域名 / JWT 密钥等信息。

使用 `-c` 指定配置文件路径：`go run . -c ./config.yml serve`。

## 配置说明

- 默认配置文件：`config.yml`（不存在时会自动生成）。
- 配置中支持环境变量占位符，例如 `${WEBSITE_DESCRIPTION}` 表示读取环境变量，详见 `docs/configuration.md`。
- 需启用 HTTPS 时请设置 `cert_file` 与 `key_file`。

## CLI

管理员相关命令见 `docs/cli.md`。

## API

基础路径：`/api/v1`

更多细节见 `docs/introduct.md` 与 `docs/api/*`。

## 文档

- `docs/introduct.md`（API 综述）
- `docs/configuration.md`
- `docs/cli.md`
- `docs/deployment.md`

## 开发调试

可选使用 Air 配合 `.air.toml` 实现热更新。

## Docker

Docker 镜像通过 `${ADDRESS}` 监听地址（例如 `0.0.0.0:5083`）。生产环境建议挂载配置与数据目录。
