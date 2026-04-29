# 部署说明

## 本地运行

1. 生成配置：`go run . gen-config`
2. 启动服务：`go run . serve -a 0.0.0.0:5083`

首次启动会自动迁移数据库并创建默认管理员（`admin` / `admin`），请尽快修改密码。

## Docker 部署

项目包含 `Dockerfile`，镜像通过 `${ADDRESS}` 指定监听地址。生产环境建议挂载以下目录：

- `config.yml`：配置文件
- `./sys`：数据库与日志
- `./data/articles`：文章与资源文件

## TLS

在 `config.yml` 中设置 `cert_file` 与 `key_file` 即可启用 HTTPS。

## 生产建议

- 配置强随机 `JWT_SECRET`
- 设置 `http_rate` 限流规则
- 如需分布式缓存，启用 Redis 并配置 `cache.redis.*`
