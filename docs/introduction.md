# 快速入门

## 部署

下面仅展示快速部署，如需自定义部署请参考[自定义部署](deploy.md)

### Docker 部署

```bash
docker run -d --name kzhikcn-api \
  -p 5083:5083 \
  -e JWT_SECRET=your_jwt_secret \
  -e WEBSITE_NAME=your_website_name \
  -e WEBSITE_DESCRIPTION=your_website_description \
  -e WEBSITE_URL=https://example.com \
  -e ADDRESS=:5083 \
  -v ./data:/app/data \
  -v ./sys:/app/sys \
  ghcr.io/k05a4b/kzhikcn-api:latest
```

**docker-compose.yml**

```yaml
services:
  kzhikcn-api:
    image: ghcr.io/k05a4b/kzhikcn-api:latest
    container_name: kzhikcn-api
    ports:
      - 5083:5083
    environment:
      - JWT_SECRET=your_jwt_secret                # JWT 密钥
      - WEBSITE_NAME=your_website_name             # 站点名称（RSS/Sitemap 相关）
      - WEBSITE_DESCRIPTION=your_website_description # 站点描述（RSS/Sitemap 相关）
      - WEBSITE_URL=https://example.com            # 网站 URL（RSS/Sitemap 相关）
      - ADDRESS=:5083                              # 监听地址
    volumes:
      - ./data:/app/data
      - ./sys:/app/sys
    restart: always
```

## 响应结构

一般情况下所有接口都返回统一的 JSON 响应结构，少部分接口在成功时可能返回原始数据（如 HTML、文件流等），但出现错误时始终返回统一的 JSON 响应结构。

标准的响应结构包含如下字段：

- `success`: 布尔值，表示请求是否成功
- `code`: HTTP 状态码
- `message`: 字符串，人类可读的信息
- `data`: 任意类型，接口返回的具体数据内容
- `meta`: 元数据对象，包含额外信息（如分页信息 `count`、`total`）
- `errorCode`: 字符串，机器可读的错误码（仅在错误时出现）
- `traceId`: 字符串，请求唯一标识（仅在错误时出现）

成功响应示例：

```json
{
  "success": true,
  "code": 200,
  "message": "",
  "data": [
    {
      "id": "651227b9-ae18-41dc-b326-f21a8e331ce1",
      "title": "文章标题"
    }
  ],
  "meta": {
    "total": 1,
    "count": 1
  }
}
```

失败响应示例：

```json
{
  "success": false,
  "code": 400,
  "message": "请求体格式错误",
  "data": null,
  "meta": {},
  "errorCode": "system.parse_payload_error",
  "traceId": "1776735700-d34baebd5f69a469"
}
```

## 查询条件表达式

某些接口支持查询条件表达式（通过查询参数 `expr`），用于筛选查询结果。

### 语法

逻辑运算符：
- `&` 与（AND）
- `|` 或（OR）
- `!` 非（NOT）

比较运算符：
- `<` 小于
- `>` 大于
- `<=` 小于等于
- `>=` 大于等于
- `!=` 不等于
- `=` 等于
- `~` 包含（LIKE）

比较运算符左侧必须是字段名，右侧必须是值。

为了安全，可查询字段采用白名单机制，**只有在白名单内的字段才能出现在表达式中**。各接口的白名单字段请参考对应文档。

示例：

查询已发布文章中点赞量大于 100 的文章：

```
status='published' & likes > 100
```

查询标题以"技术"结尾的文章：

```
title ~ "%技术"
```

### 错误码

查询表达式解析失败时返回 HTTP 400，错误码为 `system.expr.invalid`。

## 权限认证

本项目权限模型简单，使用 JWT 进行认证。系统不分普通用户/管理员，所有用户都是管理员。

JWT Token 可通过登录接口获取，详情请参考[权限认证接口](./api/auth.md)。

### 接口限流

所有接口会根据配置文件中的配置进行限流。限流机制不会影响**持有有效 JWT Token** 和**持有高配额密钥**的请求。

> [!important]
> 高配额密钥在配置文件的 `http_rate.high_quota_keys` 中配置。
> 高配额密钥**没有**访问管理员接口的权限，仅能绕过限流。

### 认证响应示例

```http
HTTP/1.1 401 Unauthorized
Content-Type: application/json

{
  "success": false,
  "code": 401,
  "message": "未授权",
  "data": null,
  "meta": {},
  "errorCode": "system.unauthorized",
  "traceId": "1777373118-8c780fdaafcee2c"
}
```

## 接口列表

- 通用
  - `GET /api/v1/ping` - 健康检查（返回服务版本和时间戳）
  - `GET /sitemap.xml` - Sitemap（需在配置中启用）
  - `GET /rss.xml` - RSS 订阅（需在配置中启用）

- [权限认证](./api/auth.md)
  - [登录流程](./api/auth.md#登录流程)
  - [登录接口](./api/auth.md#登录接口)
  - [MFA 验证接口（TOTP）](./api/auth.md#mfa验证接口totp)
  - [登出接口](./api/auth.md#登出接口)

- [管理员管理](./api/admin.md)
  - [获取当前管理员信息](./api/admin.md#1-获取当前管理员信息)
  - [更新管理员信息](./api/admin.md#2-更新管理员信息)
  - [修改密码](./api/admin.md#3-修改密码)
  - [生成 TOTP 密钥](./api/admin.md#4-生成-totp-密钥)
  - [启用/禁用 MFA](./api/admin.md#5-启用禁用-mfa)

- [文章相关接口](./api/articles.md)
  - [获取文章列表](./api/articles.md#获取文章列表)
  - [获取单篇文章信息](./api/articles.md#获取单篇文章信息)
  - [增加文章浏览量](./api/articles.md#增加文章浏览量)
  - [增加文章点赞量](./api/articles.md#增加文章点赞量)
  - [获取渲染后的文章内容](./api/articles.md#获取渲染后的文章内容)
  - [获取文章资源](./api/articles.md#获取文章资源)
  - [创建文章](./api/articles.md#创建文章)
  - [更新文章信息](./api/articles.md#更新文章信息)
  - [获取文章原始内容](./api/articles.md#获取文章原始内容)
  - [更新文章原始内容](./api/articles.md#更新文章原始内容)
  - [批量删除文章](./api/articles.md#批量删除文章)
  - [获取已删除的文章列表](./api/articles.md#获取已删除的文章列表)
  - [恢复已删除的文章](./api/articles.md#恢复已删除的文章)
  - [获取文章资源列表](./api/articles.md#获取文章资源列表)
  - [上传资源](./api/articles.md#上传资源)
  - [删除文章资源](./api/articles.md#删除资源)

- [分类相关接口](./api/categories.md)
  - [获取分类列表](./api/categories.md#1-获取分类列表公共)
  - [获取分类下的文章](./api/categories.md#2-获取分类下的文章公共)
  - [创建分类](./api/categories.md#3-管理后台---创建分类)
  - [更新分类](./api/categories.md#4-管理后台---更新分类)
  - [删除分类](./api/categories.md#5-管理后台---删除分类)

- [标签相关接口](./api/tags.md)
  - [获取标签列表](./api/tags.md#1-获取标签列表公共)
  - [获取标签下的文章](./api/tags.md#2-获取标签下的文章公共)
  - [更新标签](./api/tags.md#3-管理后台---更新标签)
  - [批量删除标签](./api/tags.md#4-管理后台---批量删除标签)
