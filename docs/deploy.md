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
