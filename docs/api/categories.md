# 分类相关 API

## 1. 获取分类列表（公共）

```
GET /api/v1/categories
```

**认证**：无需认证

**查询参数**：

| 参数名 | 类型   | 描述                                            | 默认值 |
| ------ | ------ | ----------------------------------------------- | ------ |
| page   | 整数   | 页码                                            | 1      |
| limit  | 整数   | 每页数量，最大值为 200                          | 20     |
| expr   | 字符串 | 过滤表达式，支持 `id`、`category_name`、`description` 字段 | -      |

**请求示例**：

```http
GET /api/v1/categories HTTP/1.1
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

**接口错误代码**：

| 错误码 | 错误描述 |
| :----- | :------- |
| `topics.categories.find_failed` | 查询分类信息失败 |

---

## 2. 获取分类下的文章

未认证时只能查询 `published` 状态的文章，认证后可查询所有状态的。

```
GET /api/v1/categories/{category}/articles
```

**认证**：可选（携带 JWT Token 可查询所有状态，否则仅 `published`）

**路径参数**：

- `category` - 分类 ID 或分类名称

**查询参数**：

| 参数名 | 类型   | 描述                                         | 默认值 |
| ------ | ------ | -------------------------------------------- | ------ |
| page   | 整数   | 页码                                         | 1      |
| limit  | 整数   | 每页数量，最大值为 100                       | 20     |
| expr   | 字符串 | 过滤表达式，支持文章字段（认证后方可使用 `status` 字段） | -      |

> [!note]
> 此接口支持查询条件表达式（`expr` 参数）。
> 白名单字段：`id`, `title`, `views`, `likes`, `description`, `enable_comment`, `custom_id`, `created_at`, `update_at`
> 认证后额外允许：`status`

**请求示例**：

```http
GET /api/v1/categories/1/articles HTTP/1.1
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
          { "id": 15, "tagName": "隐藏" },
          { "id": 20, "tagName": "备用" }
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

**接口错误代码**：

| 错误码 | 错误描述 |
| :----- | :------- |
| `topics.categories.not_found_category` | 没有找到分类信息 |
| `topics.categories.find_failed` | 查询分类信息失败 |

---

## 3. 管理后台 - 创建分类

```
POST /api/v1/categories
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
POST /api/v1/categories HTTP/1.1
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
  },
  "meta": {}
}
```

**接口错误代码**：

| 错误码 | 错误描述 |
| :----- | :------- |
| `topics.categories.create.category_name_is_required` | 分类名称为必填项 |
| `topics.categories.is_exist` | 分类已存在 |
| `topics.categories.create_failed` | 创建分类失败 |

---

## 4. 管理后台 - 更新分类

```
PATCH /api/v1/categories/{category}
```

**认证**：需要 JWT 认证

**路径参数**：

- `category` - 分类 ID 或分类名称

**请求体**：

| 字段名        | 类型   | 必填 | 描述     |
| ------------- | ------ | ---- | -------- |
| categoryName  | 字符串 | 否   | 分类名称 |
| description   | 字符串 | 否   | 分类描述 |

**请求示例**：

```http
PATCH /api/v1/categories/1 HTTP/1.1
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
  "data": null,
  "meta": {}
}
```

**接口错误代码**：

| 错误码 | 错误描述 |
| :----- | :------- |
| `topics.categories.not_found_category` | 没有找到分类信息 |
| `topics.categories.update_failed` | 更新分类信息失败 |

---

## 5. 管理后台 - 删除分类

批量删除分类。

```
DELETE /api/v1/categories/batch-delete
```

**认证**：需要 JWT 认证

**请求体**：

| 字段 | 类型     | 必填   | 描述               | 默认值 |
| ---- | -------- | ------ | ------------------ | ------ |
| ids  | `[]uint` | **是** | 要删除的分类 ID 列表 | -      |

**请求示例**：

```http
DELETE /api/v1/categories/batch-delete HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "ids": [3, 4]
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
  "data": null,
  "meta": {}
}
```

**接口错误代码**：

| 错误码 | 错误描述 |
| :----- | :------- |
| `topics.categories.delete.ids_is_empty` | 请提供要删除的分类 ID |
| `topics.categories.delete_failed` | 删除分类失败 |
