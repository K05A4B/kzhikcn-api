# 文章相关 API

## 获取文章列表

查询状态为`published`的文章列表

```
GET /api/v1/articles
```

**认证**：无需认证

**查询参数**：

| 参数名  | 类型   | 描述                               | 默认值             |
| ------- | ------ | ---------------------------------- | ------------------ |
| page    | 整数   | 页码                               | 1                  |
| limit   | 整数   | 每页数量，最大值为 100             | 20                 |
| orderBy | 字符串 | 排序字段，格式为 `字段名:排序方向` | "publishedAt:desc" |
| expr    | 字符串 | 过滤表达式（见快速入门）           | 无                 |

**排序字段说明**：

- `publishedAt` - 按发布时间升序排序
- `createdAt` - 按创建时间升序排序
- `updatedAt` - 按更新时间升序排序
- `likes` - 按点赞数升序排序
- `views` - 按查看数升序排序
- `publishedAt:desc` - 按发布时间降序排序
- `createdAt:desc` - 按创建时间降序排序
- `updatedAt:desc` - 按更新时间降序排序
- `likes:desc` - 按点赞数降序排序
- `views:desc` - 按查看数降序排序

**过滤表达式允许字段**：

`id`, `title`, `views`, `likes`, `description`, `enable_comment`, `custom_id`, `created_at`, `update_at`, `published_at`

**请求示例**：

```http
GET /api/v1/articles?page=1&limit=1 HTTP/1.1
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
  ],
  "meta": {
    "count": 1,
    "total": 2
  }
}
```

**接口错误代码**

- `articles.find_failed` 查询文章信息失败

---

## 获取单篇文章信息

查询状态为`published`和`hidden`的某篇文章的信息，如果没有找到相应的文章，则响应404

```
GET /api/v1/articles/{article_id}
```

**认证**：无需认证

**路径参数**：

- `article_id` - 文章 ID 或自定义 ID

**请求示例**：

```http
GET /api/v1/articles/651227b9-ae18-41dc-b326-f21a8e331ce1 HTTP/1.1
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
  },
  "meta": {}
}
```

**接口错误代码**

- `articles.find_failed` 查询文章信息失败
- `articles.not_found` 没有找到文章

---

## 增加文章浏览量

增加某篇状态为`published`、`hidden`的文章的浏览量

```
POST /api/v1/articles/{article_id}/view
```

**认证**：无需认证

**路径参数**：

- `article_id` - 文章 ID 或自定义 ID

**请求示例**：

```http
POST /api/v1/articles/651227b9-ae18-41dc-b326-f21a8e331ce1/view HTTP/1.1
Content-Type: application/json

{}
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
    "views": 1 // 响应的是当前浏览量
  },
  "meta": {}
}
```

**接口错误代码**

- `articles.not_found` 没有找到文章
- `articles.find_failed` 查询文章信息失败
- `articles.views.update_failed` 更新浏览量失败

---

## 增加文章点赞量

增加某篇状态为`published`、`hidden`的文章的点赞量

```
POST /api/v1/articles/{article_id}/like
```

**认证**：无需认证

**路径参数**：

- `article_id` - 文章 ID 或自定义 ID

**请求示例**：

```http
POST /api/v1/articles/651227b9-ae18-41dc-b326-f21a8e331ce1/like HTTP/1.1
Content-Type: application/json

{}
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
    "likes": 1 // 响应的是当前点赞量
  },
  "meta": {}
}
```

**接口错误代码**

- `articles.not_found` 没有找到文章
- `articles.find_failed` 查询文章信息失败
- `articles.likes.update_failed` 更新点赞量失败

---

## 获取渲染后的文章内容

```
GET /api/v1/articles/{article_id}/content
```

获取渲染后的状态为`published`和`hidden`的文章内容，支持 JSON 和 HTML 格式响应

**认证**：无需认证

**路径参数**：

- `article_id` - 文章 ID 或自定义 ID

**请求头**：

- `Accept` - 响应格式，支持 `application/json`（默认）和 `text/html`

**请求示例**：

**JSON 格式请求**：

```http
GET /api/v1/articles/651227b9-ae18-41dc-b326-f21a8e331ce1/content HTTP/1.1
Accept: application/json
```

**HTML 格式请求**：

```http
GET /api/v1/articles/651227b9-ae18-41dc-b326-f21a8e331ce1/content HTTP/1.1
Accept: text/html
```

**响应示例**：

**JSON 格式响应**：

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "code": 200,
  "message": "",
  "data": "<h1>文章标题</h1><p>文章内容</p>",
  "meta": {}
}
```

**HTML 格式响应**：

```html
HTTP/1.1 200 OK Content-Type: text/html; charset=utf-8

<h1>文章标题</h1>
<p>文章内容</p>
```

**接口错误代码**

- `articles.not_found` 没有找到文章
- `articles.find_failed` 查询文章信息失败
- `articles.content.not_found` 文章正文文件不存在
- `articles.content.render_failed` 渲染文章正文失败

---

## 获取文章资源

```
GET /api/v1/articles/{article_id}/assets/{asset_id}
```

**认证**：无需认证

**路径参数**：

- `article_id` - 文章 ID 或自定义 ID
- `asset_id` - 资源 ID

**请求示例**：

```http
GET /api/v1/articles/651227b9-ae18-41dc-b326-f21a8e331ce1/assets/img.jpg HTTP/1.1
```

**响应**：

- 成功：返回资源文件（如图片、文件等）
- 失败：返回 JSON 格式错误信息

**接口错误代码**：

- `articles.not_found` 没有找到文章
- `articles.find_failed` 查询文章信息失败
- `articles.content.not_found` 文章正文文件不存在
- `articles.assets.check_asset_failed`检查资源状态失败
- `articles.assets.not_found`没有找到资源
- `articles.assets.load_failed`加载资源失败

---

## 获取文章列表（管理员）

获取文章列表，和公共的相比可以查询到任意状态的文章

```
GET /api/v1/admin/articles
```

**认证**：需要 JWT 认证

**查询参数**：

> !note
> 此接口支持查询条件表达式
> 在白名单中的字段包含：
> `id`, `status`, `title`, `views`, `likes`, `description`, `enable_comment`, `custom_id`, `created_at`,`update_at`, `published_at`, `status`

- 与公共接口相同，可额外使用 `expr` 参数使用[查询条件表达式](../introduct.md#查询条件表达式)
- `onlyDeleted`（bool）：仅展示已软删除文章

**过滤表达式允许字段**：

`id`, `title`, `views`, `likes`, `description`, `enable_comment`, `custom_id`, `created_at`, `update_at`, `published_at`, `status`
**请求示例**：

```http
GET /api/v1/admin/articles HTTP/1.1
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
    "data": [
        {
            "id": "4bfbb387-b3fb-4b0f-a4a5-8cb7c7d667ac",
            "createdAt": "2026-04-10T22:03:10.9258348+08:00",
            "updatedAt": "2026-04-10T22:03:10.9258348+08:00",
            "publishedAt": null,
            "customID": "draft_1",
            "title": "草稿文章1",
            "views": 0,
            "likes": 0,
            "categoryID": 10,
            "category": {
                "id": 10,
                "categoryName": "分类10",
                "description": "分类10的描述"
            },
            "tags": [
                {
                    "id": 9,
                    "tagName": "草稿"
                },
                {
                    "id": 10,
                    "tagName": "待编辑"
                }
            ],
            "status": "draft",
            "description": "这是草稿文章1",
            "coverImage": "",
            "enableComment": false
        }
    ],
    "meta": {
        "count": 1,
        "total": 15
    }
}
```

**接口错误代码**：

- `articles.find_failed` 查询文章信息失败

---

## 创建文章

创建文章

```
POST /api/v1/admin/articles
```

**认证**：需要 JWT 认证

**请求体参数**：

| 字段名        | 类型   | 必填   | 描述                               | 默认值 |
| ------------- | ------ | ------ | ---------------------------------- | ------ |
| title         | 字符串 | **是** | 文章标题                           | 无     |
| customID      | 字符串 | 否     | 自定义文章ID                       | 文章ID |
| description   | 字符串 | 否     | 文章描述                           | -      |
| coverImage    | 字符串 | 否     | 封面图片URL                        | -      |
| category      | 字符串 | 否     | 文章分类名字（必须存在对应的分类） | -      |
| tags          | 数组   | 否     | 文章标签列表                       | []     |
| enableComment | 布尔值 | 否     | 是否启用评论                       | false  |
| status        | 字符串 | 否     | 文章状态（published/draft/hidden） | draft  |

**请求示例**：

```http
POST /api/v1/admin/articles HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "新文章",
  "tags": ["技术", "教程"]
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
    "id": "651227b9-ae18-41dc-b326-f21a8e331ce1",
    "createdAt": "2026-04-20T10:23:31.0541106+08:00",
    "updatedAt": "2026-04-20T10:23:31.0541106+08:00",
    "publishedAt": "2026-04-20T10:23:31.0541106+08:00",
    "customID": "651227b9-ae18-41dc-b326-f21a8e331ce1",
    "title": "新文章",
    "views": 0,
    "likes": 0,
    "categoryID": null,
    "category": {
      "id": 0,
      "categoryName": "",
      "description": ""
    },
    "tags": [
      {
        "id": 1,
        "tagName": "技术"
      },
      {
        "id": 2,
        "tagName": "教程"
      }
    ],
    "status": "draft",
    "description": "",
    "coverImage": "",
    "enableComment": false
  },
  "meta": {}
}
```

**接口错误代码**：

- `articles.find_failed` 查询文章信息失败
- `articles.category_not_found`没有找到分类
- `articles.create_failed`创建文章失败

---

## 更新文章信息

```
PATCH /api/v1/admin/articles/{article_id}
```

**认证**：需要 JWT 认证

**路径参数**：

- `article_id` - 文章 ID 或自定义 ID

**请求体参数**：

> !note
>
> 空值表示不修改
>
> tags字段的空值为`null`，如果tags字段为`[]`则意为删除所有标签

| 字段名        | 类型   | 必填 | 描述                               |
| ------------- | ------ | ---- | ---------------------------------- |
| title         | 字符串 | 否   | 文章标题                           |
| customID      | 字符串 | 否   | 自定义文章ID                       |
| description   | 字符串 | 否   | 文章描述                           |
| coverImage    | 字符串 | 否   | 封面图片URL                        |
| category      | 字符串 | 否   | 文章分类名字（必须存在对应的分类） |
| tags          | 数组   | 否   | 文章标签列表                       |
| enableComment | 布尔值 | 否   | 是否启用评论                       |
| status        | 字符串 | 否   | 文章状态（published/draft/hidden） |

**请求示例**：

```http
PATCH /api/v1/admin/articles/651227b9-ae18-41dc-b326-f21a8e331ce1 HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "更新后的文章标题",
  "status": "published",
  "customID": "article1",
  "tags": ["技术", "新标签"]
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
    "id": "651227b9-ae18-41dc-b326-f21a8e331ce1",
    "createdAt": "2026-04-20T10:23:31.0541106+08:00",
    "updatedAt": "2026-04-20T10:23:31.0541106+08:00",
    "publishedAt": "2026-04-20T10:23:31.0541106+08:00",
    "customID": "article1",
    "title": "更新后的文章标题",
    "views": 0,
    "likes": 0,
    "categoryID": null,
    "category": {
      "id": 0,
      "categoryName": "",
      "description": ""
    },
    "tags": [
      {
        "id": 1,
        "tagName": "技术"
      },
      {
        "id": 3,
        "tagName": "新标签"
      }
    ],
    "status": "published",
    "description": "",
    "coverImage": "",
    "enableComment": false
  },
  "meta": {}
}
```

**接口错误代码**：

- `articles.find_failed` 查询文章信息失败
- `articles.not_found`没有找到文章
- `articles.category_not_found`没有找到分类
- `articles.update_failed`更新文章信息失败

---

## 获取文章内容

获取文章的原始内容（Markdown），支持 JSON 格式和markdown格式响应

```
GET /api/v1/admin/articles/{article_id}/raw-content
```

**认证**：需要 JWT 认证

**路径参数**：

- `article_id` - 文章 ID 或自定义 ID

**请求示例**：

**JSON 格式请求**：

```http
GET /api/v1/admin/articles/651227b9-ae18-41dc-b326-f21a8e331ce1/raw-content HTTP/1.1
Authorization: Bearer <token>
Accept: application/json
```

**Markdown 格式请求**：

```http
GET /api/v1/admin/articles/651227b9-ae18-41dc-b326-f21a8e331ce1/raw-content HTTP/1.1
Authorization: Bearer <token>
Accept: text/plain
```

**响应示例**：

**JSON 格式请求**：

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "code": 200,
  "message": "",
  "data": {
    "content": "# 文章标题\n\n文章内容"
  },
  "meta": {}
}
```

**Markdown 格式请求**：

```http
HTTP/1.1 200 OK
Content-Type: text/plain

# 文章标题

文章内容
```

**接口错误代码**：

- `articles.find_failed` 查询文章信息失败
- `articles.not_found`没有找到文章
- `articles.content.not_found`没有找到文章正文
- `articles.content.load_failed`加载文章正文失败

---

## 更新文章内容

```
PUT /api/v1/admin/articles/{article_id}/raw-content
```

**认证**：需要 JWT 认证

**路径参数**：

- `article_id` - 文章 ID 或自定义 ID

**请求体**：正文文本

**请求示例**：

```http
PUT /api/v1/admin/articles/651227b9-ae18-41dc-b326-f21a8e331ce1/raw-content HTTP/1.1
Authorization: Bearer <token>
Content-Type: text/plain

# 更新后的文章标题

更新后的文章内容
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

- `articles.find_failed` 查询文章信息失败
- `articles.not_found`没有找到文章
- `articles.content.write_failed`更新文章正文失败

---

## 批量删除文章

```
DELETE /api/v1/admin/articles/batch-delete
```

**认证**：需要 JWT 认证

**请求体**：

| 字段       | 类型     | 必填   | 描述                                           | 默认值 |
| ---------- | -------- | ------ | ---------------------------------------------- | ------ |
| ids        | []string | **是** | 要删除的文章ID（不能是自定义id）               | -      |
| hardDelete | bool     | 否     | 硬删除（直接删除文章和其附带的资源，无法恢复） | false  |

**请求示例**：

```http
DELETE /api/v1/admin/articles/batch-delete HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "ids": ["651227b9-ae18-41dc-b326-f21a8e331ce1"]
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

- `articles.delete_failed`删除文章失败
- `articles.clean_assets_failed`清除文章资源失败

---

## 获取已删除的文章列表

```
GET /api/v1/admin/articles/trash-bin
```

> !note
> 此接口支持查询条件表达式
> 在白名单中的字段包含：
> `id`, `status`, `title`, `views`, `likes`, `description`, `enable_comment`, `custom_id`, `created_at`,`update_at`, `published_at`, `status`

**认证**：需要 JWT 认证

**请求示例**：

```http
GET /api/v1/admin/articles/trash-bin HTTP/1.1
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
  "data": [
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
  ],
  "meta": {
    "count": 1,
    "total": 1
  }
}
```

**接口错误代码**：

- `articles.find_failed`查询文章失败

---

## 恢复已删除的文章

```
POST /api/v1/admin/articles/trash-bin/restore
```

**认证**：需要 JWT 认证

**请求体**：

| 字段 | 类型     | 必填   | 描述                             | 默认值 |
| ---- | -------- | ------ | -------------------------------- | ------ |
| ids  | []string | **是** | 要恢复的文章ID（不能是自定义id） | -      |

**请求示例**：

```http
POST /api/v1/admin/articles/trash-bin/restore HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "ids": ["651227b9-ae18-41dc-b326-f21a8e331ce1"]
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

- `articles.restore_failed`恢复文章失败

---

## 获取文章资源列表

```
GET /api/v1/admin/articles/{article_id}/assets
```

**认证**：需要 JWT 认证

**路径参数**：

- `article_id` - 文章 ID 或自定义 ID

**请求示例**：

```http
GET /api/v1/admin/articles/651227b9-ae18-41dc-b326-f21a8e331ce1/assets HTTP/1.1
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
  "data": [
    "image1.jpg",
    "1234.exe",
    "5678.pptx"
  ],
  "meta": {}
}

```

**接口错误代码**：

- `articles.assets.list_failed`列出资源列表失败

---

## 上传文章资源

```
POST /api/v1/admin/articles/{article_id}/assets
```

**认证**：需要 JWT 认证

**路径参数**：

- `article_id` - 文章 ID 或自定义 ID

**请求**：

- 内容类型：`multipart/form-data`
- 表单字段：`file` - 要上传的文件（filename字段的值就是asset id）

**请求示例**：

```http
POST /api/v1/admin/articles/651227b9-ae18-41dc-b326-f21a8e331ce1/assets HTTP/1.1
Host: example.com
Authorization: Bearer <token>
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW

------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="file"; filename="image.jpg"
Content-Type: image/jpeg

[文件内容]
------WebKitFormBoundary7MA4YWxkTrZu0gW--
```

**响应示例**：

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "code": 200,
  "message": "",
  "data": "image.jpg", // asset id
  "meta": {}
}
```

**接口错误代码**：

- `articles.assets.file_missing`上传的负载中没有找到文件
- `articles.assets.invalid_filename`资源ID（文件名）不合法
- `articles.assets.filename_is_required`资源ID（文件名）是必须提供的
- `articles.assets.load_failed`加载资源失败
- `articles.assets.upload_failed`上传资源失败

---

## 删除文章资源

```
DELETE /api/v1/admin/articles/{article_id}/assets/{asset_id}
```

**认证**：需要 JWT 认证

**路径参数**：

- `article_id` - 文章 ID 或自定义 ID
- `asset_id` - 资源 ID

**请求示例**：

```http
DELETE /api/v1/admin/articles/651227b9-ae18-41dc-b326-f21a8e331ce1/assets/image1.jpg HTTP/1.1
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
    "data": null,
    "meta": {}
}
```

**接口错误代码**：

- `articles.delete_assets_failed`删除文章资源失败
