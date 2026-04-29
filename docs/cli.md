# CLI 使用

可执行文件名取决于编译方式（如 `kzhikcn` 或 `kzhikcn.exe`）。

全局参数：

- `-c, --config`：配置文件路径（默认 `config.yml`）

## 生成配置文件

```
<binary> gen-config
```

根据提示生成 `config.yml`，会询问域名、JWT 密钥等信息；若已存在会覆盖写入。

不使用 `gen-config` 也可以直接启动服务：当 `config.yml` 缺失时，会从默认模板自动生成，但不会进行交互式提问。

## 启动服务

```
<binary> serve -a 0.0.0.0:5083
```

- `-a, --address`：监听地址

## 管理员命令

```
<binary> admin <subcommand>
```

### 添加管理员

```
<binary> admin add -n <name> -p <password> [-e <email>]
```

### 修改管理员

```
<binary> admin modify -i <id> [--mfa <true|false>] [--totp-secret <secret>]
<binary> admin modify -n <name> [--mfa <true|false>] [--totp-secret <secret>]
```

### 修改管理员密码

```
<binary> admin passwd -n <name> -p <new-password>
```

### 查询管理员

```
<binary> admin find -n <name>
```
