# 分类相关 API

## 1. 获取分类列表（公共）

```
GET /api/v1/categories
```

**认证**：无需认证

**查询参数**：

| 参数名  | 类型   | 描述                                         | 默认值 |
|--------|--------|----------------------------------------------|--------|
| page   | 整数   | 页码                                         | 1      |
| limit  | 整数   | 每页数量，最大值为 200                       | 20     |
| expr   | 字符串 | 过滤表达式，支持 `id`、`category_name`、`description` 字段 | 无     |

**请求示例**：

```http
GET /api/v1/categories HTTP/1.1
Host: example.com
```

**响应示例**：

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "code": 200,
  "message": "",
  "data": [
    {
      "id": 1,
      "categoryName": "分类1",
      "description": "分类1的描述"
    },
    {
      "id": 2,
      "categoryName": "分类2",
      "description": "分类2的描述"
    }
  ],
  "meta": {
    "count": 2,
    "total": 2
  }
}
```

## 2. 获取分类下的文章（公共）

```
GET /api/v1/categories/{category}/articles
```

**认证**：无需认证

**路径参数**：
- `category` - 分类 ID 或分类名称

**查询参数**：

| 参数名  | 类型   | 描述                                         | 默认值 |
|--------|--------|----------------------------------------------|--------|
| page   | 整数   | 页码                                         | 1      |
| limit  | 整数   | 每页数量，最大值为 100                       | 20     |
| expr   | 字符串 | 过滤表达式，支持多个文章字段                 | 无     |

**请求示例**：

```http
GET /api/v1/categories/1/articles HTTP/1.1
Host: example.com
```

**响应示例**：

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "code": 200,
  "message": "",
  "data": {
    "id": 1,
    "categoryName": "分类1",
    "description": "分类1的描述",
    "articles": [
      {
        "id": "651227b9-ae18-41dc-b326-f21a8e331ce1",
        "createdAt": "2026-04-20T10:23:31.0541106+08:00",
        "updatedAt": "2026-04-20T10:23:31.0541106+08:00",
        "publishedAt": "2026-04-20T10:23:31.0541106+08:00",
        "customID": "article1",
        "title": "文章1",
        "views": 0,
        "likes": 0,
        "categoryID": 1,
        "tags": [
          {
            "id": 15,
            "tagName": "隐藏"
          },
          {
            "id": 20,
            "tagName": "备用"
          }
        ],
        "status": "published",
        "description": "文章1的描述",
        "coverImage": "https://example.com/example.jpg",
        "enableComment": true
      }
    ]
  },
  "meta": {
    "count": 1,
    "total": 1
  }
}
```

## 3. 管理后台 - 创建分类

```
POST /api/v1/admin/categories
```

**认证**：需要 JWT 认证

**请求体**：

```json
{
  "categoryName": "新分类",
  "description": "新分类的描述"
}
```

**请求示例**：

```http
POST /api/v1/admin/categories HTTP/1.1
Host: example.com
Authorization: Bearer <token>
Content-Type: application/json

{
  "categoryName": "技术",
  "description": "技术相关的文章"
}
```

**响应示例**：

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "code": 200,
  "message": "",
  "data": {
    "id": 3,
    "categoryName": "技术",
    "description": "技术相关的文章"
  }
}
```

**错误响应**：

- 400 Bad Request - 分类名称为空
  ```json
  {
    "success": false,
    "code": 400,
    "message": "分类名称不能为空",
    "data": null
  }
  ```

- 409 Conflict - 分类已存在
  ```json
  {
    "success": false,
    "code": 409,
    "message": "分类已存在",
    "data": null
  }
  ```

## 4. 管理后台 - 更新分类

```
PATCH /api/v1/admin/categories/{category}
```

**认证**：需要 JWT 认证

**路径参数**：
- `category` - 分类 ID 或分类名称

**请求体**：

```json
{
  "categoryName": "更新后的分类名称",
  "description": "更新后的分类描述"
}
```

**请求示例**：

```http
PATCH /api/v1/admin/categories/1 HTTP/1.1
Host: example.com
Authorization: Bearer <token>
Content-Type: application/json

{
  "categoryName": "技术分类",
  "description": "更新后的技术分类描述"
}
```

**响应示例**：

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "code": 200,
  "message": "",
  "data": null
}
```

**错误响应**：

- 404 Not Found - 分类不存在
  ```json
  {
    "success": false,
    "code": 404,
    "message": "分类不存在",
    "data": null
  }
  ```

## 5. 管理后台 - 删除分类

```
DELETE /api/v1/admin/categories/{category}
```

**认证**：需要 JWT 认证

**路径参数**：
- `category` - 分类 ID 或分类名称

**请求示例**：

```http
DELETE /api/v1/admin/categories/3 HTTP/1.1
Host: example.com
Authorization: Bearer <token>
```

**响应示例**：

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "code": 200,
  "message": "",
  "data": null
}
```

**错误响应**：

- 404 Not Found - 分类不存在
  ```json
  {
    "success": false,
    "code": 404,
    "message": "分类不存在",
    "data": null
  }
  ```