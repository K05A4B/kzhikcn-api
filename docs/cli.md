# CLI 工具

`kzhikcn` 除启动服务外，还内置了一套命令行工具，用于生成配置、管理管理员、执行数据库迁移以及校验配置。

> [!NOTE]
> CLI 的帮助信息中程序名显示为 `kzhikcn-api-cli`，以下示例统一使用编译后的二进制名 `kzhikcn`。

## 基本用法

```bash
kzhikcn [全局选项] <命令> [命令选项]
```

查看帮助与版本：

```bash
kzhikcn --help
kzhikcn --version
kzhikcn admin --help
```

## 全局选项

| 选项 | 别名 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `--config` | `-c` | `config.yml` | 配置文件路径 |
| `--log-level` | | 配置值 | 覆盖日志级别，可选 `debug` / `info` / `warn` / `error`，仅本次运行生效 |
| `--quiet` | `-q` | `false` | 静默模式，等价于 `--log-level error` |
| `--help` | `-h` | | 显示帮助 |
| `--version` | `-v` | | 显示版本 |

> [!TIP]
> 全局选项需位于命令之前，命令选项需位于命令之后。
> `--format` 属于命令选项，仅在 `admin find`、`admin list`、`config show` 上提供，需放在对应子命令之后。

命令按类别分组：

| 分类 | 命令 | 说明 |
| :--- | :--- | :--- |
| 基础命令 | `gen-config` | 生成配置文件 |
| 基础命令 | `serve` | 启动 HTTP 服务 |
| 配置与数据 | `migrate` | 执行数据库迁移并退出 |
| 配置与数据 | `config` | 配置校验与查看 |
| 管理员 | `admin` | 管理员增删改查与 MFA |

---

## gen-config

生成配置文件。

| 选项 | 别名 | 说明 |
| :--- | :--- | :--- |
| `--default` | `-d` | 直接输出完整默认配置，不进行交互 |

不带 `-d` 时会依次提示输入 `WEBSITE_URL`、`WEBSITE_NAME`、`WEBSITE_DESCRIPTION`、`JWT_SECRET`；`JWT_SECRET` 留空会自动随机生成。生成时会覆盖目标文件（截断写入）。

```bash
# 生成完整默认配置
kzhikcn gen-config -d

# 交互式生成
kzhikcn -c ./sys/config.yml gen-config
```

---

## serve

启动 HTTP 服务。若配置文件不存在会自动写入默认配置。

| 选项 | 别名 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `--address` | `-a` | `0.0.0.0:5083` | 服务监听地址 |

```bash
kzhikcn serve
kzhikcn -c ./sys/config.yml serve -a 100.65.10.13:5083
```

启动流程：加载（或创建）配置 → 初始化数据库/缓存/存储 → 执行迁移 → 构建路由 → 监听端口。收到 `SIGINT` / `SIGTERM` 后会在 5 秒内优雅关闭。

---

## migrate

加载配置、连接数据库并执行迁移，完成后立即退出，不启动服务。

```bash
kzhikcn -c ./config.yml migrate
```

---

## config

配置相关的子命令。

### config check

校验配置文件，检查必填项与取值合法性，包括：

- `db.driver` / `db.dsn`
- `storage.provider` / `storage.articles.base_path`
- `cache.provider`（`local` 需 `cache.local.dir`，`redis` 需 `cache.redis.addr`）
- `auth.jwt.secret`
- 启用 RSS/Sitemap 时的 `machine_readable_resources.base_url`
- `log.log_level` 是否为合法级别

未被解析的环境变量占位符（如 `${JWT_SECRET}` 对应的环境变量未设置）也会被视为错误。

```bash
kzhikcn -c ./config.yml config check
```

校验通过：

```text
配置文件 ./config.yml 校验通过
```

校验失败时逐条输出问题并返回非零退出码：

```text
配置错误: auth.jwt.secret 存在未解析的环境变量占位符
配置文件 ./config.yml 校验未通过，共 1 项错误
```

### config show

输出当前生效的配置，敏感字段（`auth.jwt.secret`、`cache.redis.password`）以 `******` 脱敏。

| 选项 | 说明 |
| :--- | :--- |
| `--format` | `text`（默认，YAML）或 `json` |

```bash
kzhikcn -c ./config.yml config show
kzhikcn -c ./config.yml config show --format json
```

---

## admin

管理员相关命令。所有 `admin` 子命令都会在加载配置后自动执行数据库迁移，再执行具体操作。

> [!NOTE]
> `admin`、`migrate` 等命令只访问数据库，不会初始化缓存，因此在服务运行期间也可以安全执行，不会与 `serve` 争抢 Badger 缓存目录锁（Badger 为独占锁，缓存目录同时只能被一个进程打开）。

### admin add

添加管理员。

| 选项 | 别名 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `--name` | `-n` | 是 | 管理员名称 |
| `--password` | `-p` | 否 | 密码，省略时交互式输入 |
| `--email` | `-e` | 否 | 电子邮件 |

```bash
kzhikcn admin add -n alice -e alice@example.com
kzhikcn admin add -n alice -p secret123
```

> [!NOTE]
> 密码长度不得少于 6 位；用户名重复会直接报错。`-p` 省略时会隐藏输入并要求二次确认，此方式仅在交互式终端可用。

### admin modify

修改管理员信息。使用 `-i` 或 `-n` 指定被修改的管理员；若用 `-i` 指定后再传 `-n`，则表示修改用户名。

| 选项 | 别名 | 说明 |
| :--- | :--- | :--- |
| `--id` | `-i` | 通过 ID 选定管理员 |
| `--name` | `-n` | 通过用户名选定管理员 / 修改后的用户名 |
| `--email` | | 设置电子邮件 |
| `--avatar` | | 设置头像地址 |
| `--mfa` | | 设置 MFA 状态（`--mfa` 启用 / `--mfa=false` 禁用） |
| `--totp-secret` | | 设置 TOTP 密钥 |

```bash
# 按用户名定位并修改邮箱
kzhikcn admin modify -n alice --email alice@example.com

# 按 ID 定位并把用户名改为 bob
kzhikcn admin modify -i 2 -n bob

# 启用 MFA
kzhikcn admin modify -n alice --mfa
```

未提供任何要修改的字段时会报错。

### admin passwd

修改管理员密码。

| 选项 | 别名 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `--name` | `-n` | 是 | 管理员账户名 |
| `--password` | `-p` | 否 | 新密码，省略时交互式输入 |

```bash
kzhikcn admin passwd -n alice
kzhikcn admin passwd -n alice -p newsecret
```

### admin find

查询单个管理员信息。

| 选项 | 别名 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `--name` | `-n` | 是 | 管理员账户名 |
| `--format` | | 否 | `text` 或 `json` |

```bash
kzhikcn admin find -n alice
kzhikcn admin find -n alice --format json
```

JSON 输出示例：

```json
{
  "id": 2,
  "username": "alice",
  "email": "alice@example.com",
  "avatar": "",
  "enableMFA": false
}
```

### admin list

列出所有管理员。

| 选项 | 说明 |
| :--- | :--- |
| `--format` | `text`（默认，表格）或 `json` |

```bash
kzhikcn admin list
kzhikcn admin list --format json
```

文本输出示例：

```text
ID  用户名    MFA  电子邮件            头像
1   admin    否
2   alice    否    alice@example.com
```

### admin delete

删除管理员。需指定 `-i` 或 `-n`，删除前会请求确认。

| 选项 | 别名 | 说明 |
| :--- | :--- | :--- |
| `--id` | `-i` | 通过 ID 指定管理员 |
| `--name` | `-n` | 通过用户名指定管理员 |
| `--yes` | `-y` | 跳过确认，适用于非交互环境 |

```bash
kzhikcn admin delete -n alice
kzhikcn admin delete -i 2 --yes
```

### admin mfa

为管理员生成或重置 TOTP 密钥，输出 `otpauth` URL，可直接导入验证器应用。

| 选项 | 别名 | 说明 |
| :--- | :--- | :--- |
| `--id` | `-i` | 通过 ID 指定管理员 |
| `--name` | `-n` | 通过用户名指定管理员 |
| `--enable` | | 生成后立即启用 MFA |
| `--force` | | 覆盖已存在的 TOTP 密钥 |

```bash
# 生成密钥并启用
kzhikcn admin mfa -n alice --enable

# 已配置过 TOTP 时需强制覆盖
kzhikcn admin mfa -n alice --force
```

> [!WARNING]
> 管理员已存在 TOTP 密钥时，必须显式使用 `--force` 才会覆盖，避免误操作导致已绑定的验证器失效。

---

## 交互式输入

以下操作依赖交互式终端，非交互环境（如 CI、管道）需改用对应选项：

| 场景 | 交互行为 | 非交互替代 |
| :--- | :--- | :--- |
| `admin add` / `admin passwd` 未提供 `--password` | 隐藏读取密码并二次确认 | 显式传入 `--password` |
| `admin delete` 未提供 `--yes` | 请求 `y/N` 确认 | 显式传入 `--yes` |

## 退出码

| 退出码 | 说明 |
| :--- | :--- |
| `0` | 执行成功 |
| 非 `0` | 执行失败，输出错误信息；`--help` / `--version` 等由 CLI 框架返回其自身退出码 |

## 日志输出

日志级别优先取配置文件中的 `log.log_level`，可用全局 `--log-level` 或 `--quiet` 覆盖。

控制台输出的规则：

- `serve` 始终输出到控制台；
- 其它子命令在未启用文件日志（`log.enable: false`）时输出到控制台，启用文件日志后仅写入日志文件，避免与文件重复。

## 常见工作流

首次部署：

```bash
# 1. 生成配置并按需修改
kzhikcn gen-config -d

# 2. 校验配置
kzhikcn -c ./config.yml config check

# 3. 执行数据库迁移（serve 也会自动执行）
kzhikcn -c ./config.yml migrate

# 4. 创建管理员账号
kzhikcn -c ./config.yml admin add -n alice

# 5. 启动服务
kzhikcn -c ./config.yml serve
```

脚本中读取管理员列表：

```bash
kzhikcn -c ./config.yml admin list --format json
```
