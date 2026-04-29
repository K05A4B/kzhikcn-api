# 快速入门

## 响应结构

一般情况下所有接口都是返回统一的响应结构，少部分接口在成功时可能不会返回统一的响应结构，但是出现错误时一般情况下都会返回统一的响应结构

标准的响应结构包含如下字段：

- `success`: 响应是否成功，布尔类型，表示请求是否成功
- `code`: HTTP状态码，HTTP的状态码和这个状态码对应
- `message`: 响应消息，字符串类型，这个字段会包含人类可读的信息
- `data`: 响应数据，任意类型，包含接口返回的具体数据内容
- `meta`: 元数据，不同接口返回的元数据不同，包含额外的信息，如分页信息
- `errorCode`: 错误码，字符串类型，机器可读的错误消息（可选值，当出现错误时返回）
- `traceId`: 跟踪ID，字符串类型，用于标识请求的唯一标识符（可选值，出现错误时返回）

成功响应示例：

```json
{
  "success": true,
  "code": 200,
  "message": "请求成功",
  "data": [
    {
    "id": 123,
    "name": "张三",
    "age": 30
    }
  ],
  "meta": {
    "total": 1,
    "count": 1,
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

某些接口支持查询条件表达式（查询参数 `expr`），用于筛选查询结果

### 查询条件表达式语法
逻辑运算符：
- `&` 与
- `|` 或
- `!` 非

比较运算符：
- `<` 小于
- `>` 大于
- `<=` 小于等于
- `>=` 大于等于
- `!=` 不等于
- `=` 等于
- `~` 包含

其中比较运算符左侧必须是字段，右侧必须是值

逻辑运算符两侧不能是字段或者值

为了保证安全，左侧字段一般来说采取白名单机制，**只有在白名单内的字段才能出现在表达式中**，具体哪些字段在白名单内请参考对应的接口文档

例子：

查询已发布的文章中点赞点赞量大于100的文章列表
```
status='published' & likes > 100 
```

查询标题以“技术”结尾的文章
```
title ~ "%技术"
```

### 查询条件表达式错误码

查询表达式解析出错后会返回http 400，且错误码为`system.expr.invalid`，意为查询条件表达式无效

## 权限认证

本项目的权限模型非常简单，使用jwt进行权限认证，后台不分普通用户/管理员用户，所有用户都是管理员用户

jwt token 可以通过权限认证相关接口认证后获取，详情请参考[权限认证接口](./api/auth.md)

### 接口限流

所有接口会依照配置文件中配置的内容进行api限流，但是限流机制不会影响**管理员**和**拥有高配额密钥**的用户

> !important
> 高配额密钥在配置文件的httprate.high_quota_keys中配置
> 高配额密钥**并没有**访问管理员接口的权限，只能绕过限流机制

### 公共接口与管理员接口

所有管理员接口都已`/admin/`开头

**例如：**
`/api/v1/articles` 和 `/api/v1/admin/articles`，前者为普通用户接口，后者为管理员接口（后者能查询`draft`状态的文章）

### 未授权响应示例

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

+ [权限认证](./api/auth.md)
+ [文章相关接口](./api/articles.md)
  + [获取文章列表](./api/articles.md#获取文章列表)
  + [获取单篇文章信息](./api/articles.md#获取单篇文章信息)
  + [增加文章浏览量](./api/articles.md#增加文章浏览量)
  + [增加文章点赞量](./api/articles.md#增加文章点赞量)
  + [获取渲染后的文章内容](./api/articles.md#获取渲染后的文章内容)
  + [获取文章资源](./api/articles.md#获取文章资源)
  + [获取文章列表（管理员）](./api/articles.md#获取文章列表（管理员）)
  + [创建文章](./api/articles.md#创建文章)
  + [更新文章信息](./api/articles.md#更新文章信息)
  + [获取文章内容](./api/articles.md#获取文章内容)
  + [更新文章内容](./api/articles.md#更新文章内容)
  + [批量删除文章](./api/articles.md#批量删除文章)
  + [获取已删除的文章列表](./api/articles.md#获取已删除的文章列表)
  + [恢复已删除的文章](./api/articles.md#恢复已删除的文章)
  + [获取文章资源列表](./api/articles.md#获取文章资源列表)
  + [上传资源](./api/articles.md#上传资源)
  + [删除文章资源](./api/articles.md#删除资源)

+ [文章分类相关接口](./api/categories.md)
+ [文章标签相关接口](./api/tags.md)
