# 标签相关 API

## 获取标签列表（公共）

```
GET /api/v1/tags
```

**认证**：无需认证

**查询参数**：

| 参数名 | 类型   | 描述                     | 默认值 |
| ------ | ------ | ------------------------ | ------ |
| page   | 整数   | 页码                     | 1      |
| limit  | 整数   | 每页数量，最大值为 200   | 50     |
| expr   | 字符串 | 过滤表达式（见快速入门） | 无     |

**过滤表达式允许字段**：

`id`, `tag_name`

**响应示例**：

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "code": 200,
  "message": "",
  "data": [
    { "id": 1, "tagName": "技术" },
    { "id": 2, "tagName": "教程" }
  ],
  "meta": {
    "count": 2,
    "total": 2
  }
}
```

**接口错误代码**：

- `topics.tags.find_failed` 查询标签失败

---

## 获取标签下的文章（公共）

```
GET /api/v1/tags/{tag}/articles
```

**认证**：无需认证

**路径参数**：

- `tag` - 标签 ID 或标签名称

**查询参数**：

| 参数名 | 类型   | 描述                     | 默认值 |
| ------ | ------ | ------------------------ | ------ |
| page   | 整数   | 页码                     | 1      |
| limit  | 整数   | 每页数量，最大值为 100   | 20     |
| expr   | 字符串 | 过滤表达式（见快速入门） | 无     |

**过滤表达式允许字段**：

`id`, `title`, `views`, `likes`, `description`, `enable_comment`, `custom_id`, `created_at`, `update_at`

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
        "title": "文章1",
        "views": 0,
        "likes": 0,
        "status": "published"
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

- `topics.tags.not_found_tag` 没有找到标签
- `topics.tags.find_failed` 查询标签失败

---

## 管理后台 - 更新标签

```
PATCH /api/v1/admin/tags/{tag}
```

**认证**：需要 JWT 认证

**路径参数**：

- `tag` - 标签 ID 或标签名称

**请求体**：

```json
{
  "tagName": "新标签名称"
}
```

**响应**：成功时返回统一响应结构，`data` 为空

**接口错误代码**：

- `topics.tags.not_found_tag` 没有找到标签
- `topics.tags.update_failed` 更新标签失败
- `topics.tags.find_failed` 查询标签失败

---

## 管理后台 - 批量删除标签

```
POST /api/v1/admin/tags/batch-delete
```

**认证**：需要 JWT 认证

**请求体**：

| 字段 | 类型   | 必填 | 描述                   |
| ---- | ------ | ---- | ---------------------- |
| ids  | []uint | 是   | 需要删除的标签 ID 列表 |

**响应**：成功时返回统一响应结构，`data` 为空

**接口错误代码**：

- `topics.tags.delete.ids_is_empty` ids 不能为空
- `topics.tags.delete_failed` 删除标签失败
