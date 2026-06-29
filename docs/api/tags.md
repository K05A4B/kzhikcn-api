# 标签相关 API

## 1. 获取标签列表（公共）

```
GET /api/v1/tags
```

**认证**：无需认证

**查询参数**：

| 参数名 | 类型   | 描述                                            | 默认值 |
| ------ | ------ | ----------------------------------------------- | ------ |
| page   | 整数   | 页码                                            | 1      |
| limit  | 整数   | 每页数量，最大值为 200                          | 50     |
| expr   | 字符串 | 过滤表达式，支持 `id`、`tag_name` 字段          | -      |

**请求示例**：

```http
GET /api/v1/tags HTTP/1.1
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
      "tagName": "技术"
    },
    {
      "id": 2,
      "tagName": "教程"
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
| `topics.tags.find_failed` | 查询标签信息失败 |

---

## 2. 获取标签下的文章（公共）

获取某个标签下所有 `published` 状态的文章。

```
GET /api/v1/tags/{tag}/articles
```

**认证**：无需认证

**路径参数**：

- `tag` - 标签 ID 或标签名称

**查询参数**：

| 参数名 | 类型   | 描述                                         | 默认值 |
| ------ | ------ | -------------------------------------------- | ------ |
| page   | 整数   | 页码                                         | 1      |
| limit  | 整数   | 每页数量，最大值为 100                       | 20     |
| expr   | 字符串 | 过滤表达式，支持文章字段（同文章列表白名单） | -      |

**请求示例**：

```http
GET /api/v1/tags/1/articles HTTP/1.1
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
    "tagName": "技术",
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
        "category": {
          "id": 1,
          "categoryName": "分类1",
          "description": "分类1的描述"
        },
        "tags": [
          { "id": 1, "tagName": "技术" },
          { "id": 2, "tagName": "教程" }
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
| `topics.tags.not_found_tag` | 没有找到标签信息 |
| `topics.tags.find_failed` | 查询标签信息失败 |

---

## 3. 管理后台 - 更新标签

```
PATCH /api/v1/tags/{tag}
```

**认证**：需要 JWT 认证

**路径参数**：

- `tag` - 标签 ID 或标签名称

**请求体**：

| 字段名   | 类型   | 必填 | 描述   |
| -------- | ------ | ---- | ------ |
| tagName  | 字符串 | 否   | 新名称 |

**请求示例**：

```http
PATCH /api/v1/tags/1 HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "tagName": "新技术"
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
| `topics.tags.not_found_tag` | 没有找到标签信息 |
| `topics.tags.update_failed` | 更新标签信息失败 |

---

## 4. 管理后台 - 批量删除标签

```
DELETE /api/v1/tags/batch-delete
```

**认证**：需要 JWT 认证

**请求体**：

| 字段 | 类型     | 必填   | 描述               | 默认值 |
| ---- | -------- | ------ | ------------------ | ------ |
| ids  | `[]uint` | **是** | 要删除的标签 ID 列表 | -      |

**请求示例**：

```http
DELETE /api/v1/tags/batch-delete HTTP/1.1
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
| `topics.tags.delete.ids_is_empty` | 请提供要删除的标签 ID |
| `topics.tags.delete_failed` | 删除标签失败 |