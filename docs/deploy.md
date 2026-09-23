# 自定义部署指南

## 编译

```bash
git clone https://github.com/k05a4b/kzhikcn-api.git
cd kzhikcn-api
go mod download
go build -o kzhikcn
```

## 配置文件

运行服务前需准备配置文件。执行 `./kzhikcn serve` 时会自动创建默认配置文件，也可通过 `./kzhikcn gen-config` 命令生成。添加 `-d` 参数可生成完整的默认配置文件：

```bash
./kzhikcn gen-config -d
```

> [!note]
> 配置文件支持环境变量引用，格式为 `${环境变量名}`，例如 `${JWT_SECRET}` 表示读取环境变量 `JWT_SECRET` 的值。

> [!warning]
> 环境变量引用仅对字符串类型配置项生效。

各配置项的完整说明请参考[配置文件参考](config.md)。

## 启动服务

启动服务的基础命令如下：

```bash
./kzhikcn serve
```

默认使用 `./config.yml` 作为配置文件，监听 `0.0.0.0:5083` 端口。

### 指定配置文件

使用 `-c` 参数指定自定义配置文件路径：

```bash
./kzhikcn -c ./config.yml serve
```

### 指定监听地址

使用 `-a` 参数指定服务监听地址：

```bash
./kzhikcn serve -a 100.65.10.13:5083
```

### 组合使用

同时指定配置文件和监听地址：

```bash
./kzhikcn -c ./sys/config.yml serve -a 100.65.10.13:64435
```

## Docker 部署

镜像提供两个版本：

| 标签 | 说明 |
| :--- | :--- |
| `latest` / `<version>` | 精简镜像，仅包含运行时依赖 |
| `latest-extends` / `<version>-extends` | 扩展镜像，额外内置 `curl`、`jq`、`bash`、`python3`、`git`、`imagemagick` 等常用工具，便于 `events` 的 `command` 钩子执行脚本与网络请求 |

```bash
docker run -d --name kzhikcn \
  -p 5083:5083 \
  -e ADDRESS=0.0.0.0:5083 \
  -e JWT_SECRET=your-secret \
  -v ./sys:/app/sys \
  -v ./data:/app/data \
  kzhikcn-api:latest-extends
```

也可自行构建指定版本：

```bash
docker build --target production -t kzhikcn-api:latest .
docker build --target extends -t kzhikcn-api:latest-extends .
```

> [!warning]
> `events` 的 `command` 钩子以服务进程权限执行任意 shell 命令。外层 shell 为 `sh -c`，如需 bash 请在 `entry` 中显式书写 `bash -c '...'`。长任务需调大 `event_dispatcher.timeout`（默认 `5s`），否则会被超时终止。
