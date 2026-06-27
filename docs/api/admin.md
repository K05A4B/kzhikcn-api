# 管理员管理 API

## 1. 获取当前管理员信息

```
GET /api/v1/users/me
```

**认证**：需要 JWT 认证

**请求示例**：

```http
GET /api/v1/users/me HTTP/1.1
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
  "data": {
    "id": 1,
    "username": "admin",
    "email": "",
    "avatar": "",
    "enableMFA": false
  },
  "meta": {}
}
```

**接口错误代码**：

| 错误码 | 错误描述 |
| :----- | :------- |
| `users.admin.find_failed` | 查询管理员失败 |

---

## 2. 更新管理员信息

```
PATCH /api/v1/users/me
```

**认证**：需要 JWT 认证

**请求体参数**：

| 字段名   | 类型   | 必填 | 描述     |
| -------- | ------ | ---- | -------- |
| username | 字符串 | 否   | 用户名   |
| email    | 字符串 | 否   | 邮箱地址 |
| avatar   | 字符串 | 否   | 头像 URL |

> [!note]
> 空字符串表示不修改该字段。

**请求示例**：

```http
PATCH /api/v1/users/me HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "email": "user@example.com",
  "avatar": "https://example.com/avatar.jpg"
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
| `users.admin.find_failed` | 查询管理员失败 |
| `users.admin.update_info_failed` | 更新管理员信息失败 |

---

## 3. 修改密码

```
PUT /api/v1/users/me/password
```

**认证**：需要 JWT 认证

**请求体参数**：

| 字段名      | 类型   | 必填 | 描述       |
| ----------- | ------ | ---- | ---------- |
| oldPassword | 字符串 | **是** | 旧密码     |
| newPassword | 字符串 | **是** | 新密码     |

**请求示例**：

```http
PUT /api/v1/users/me/password HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "oldPassword": "admin",
  "newPassword": "newpassword123"
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
| `users.admin.not_found` | 没有找到对应管理员 |
| `users.admin.find_failed` | 查询管理员失败 |
| `users.admin.compare_password_failed` | 校验密码时失败 |
| `users.admin.validate_failed` | 旧密码验证失败 |
| `users.admin.change_password_failed` | 更新密码失败 |

---

## 4. 生成 TOTP 密钥

为当前用户生成 TOTP 密钥并保存。生成后需调用启用 MFA 接口才能生效。

```
POST /api/v1/users/me/mfa/totp-secret
```

**认证**：需要 JWT 认证

**请求体参数**：

| 字段名   | 类型   | 必填 | 描述   |
| -------- | ------ | ---- | ------ |
| password | 字符串 | **是** | 密码确认 |

**请求示例**：

```http
POST /api/v1/users/me/mfa/totp-secret HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "password": "admin"
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
    "secret": "JBSWY3DPEHPK3PXP",
    "accountName": "admin",
    "issuer": "kzhikcn-api",
    "url": "otpauth://totp/kzhikcn-api:admin?secret=JBSWY3DPEHPK3PXP&issuer=kzhikcn-api"
  },
  "meta": {}
}
```

> [!note]
> `url` 字段可直接生成二维码供认证器应用（如 Google Authenticator、Authy）扫描。

**接口错误代码**：

| 错误码 | 错误描述 |
| :----- | :------- |
| `users.admin.not_found` | 没有找到对应管理员 |
| `users.admin.find_failed` | 查询管理员失败 |
| `users.admin.compare_password_failed` | 密码校验失败 |
| `users.admin.totp.generate_failed` | 生成 TOTP 失败 |
| `users.admin.totp.update_secret_failed` | 更新 TOTP 密钥失败 |

---

## 5. 启用/禁用 MFA

启用或禁用当前用户的多因素认证（需先[生成 TOTP 密钥](#4-生成-totp-密钥)）。

```
PUT /api/v1/users/me/mfa/{action}
```

**认证**：需要 JWT 认证

**路径参数**：

- `action` - 操作类型：`enable`（启用）或 `disable`（禁用）

**请求体参数**：

| 字段名   | 类型   | 必填 | 描述                         |
| -------- | ------ | ---- | ---------------------------- |
| password | 字符串 | **是** | 密码确认                     |
| otp      | 字符串 | **是** | TOTP 验证码（禁用时也需要）  |

> [!note]
> 启用 MFA 时需要提供正确的 TOTP 验证码来验证密钥配置无误。
> 禁用 MFA 时同样需要提供 TOTP 验证码以防止恶意关闭。

**请求示例**（启用）：

```http
PUT /api/v1/users/me/mfa/enable HTTP/1.1
Authorization: Bearer <token>
Content-Type: application/json

{
  "password": "admin",
  "otp": "750629"
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
| `users.admin.not_found` | 没有找到对应管理员 |
| `users.admin.find_failed` | 查询管理员失败 |
| `users.admin.compare_password_failed` | 密码校验失败 |
| `users.admin.mfa.invalid_otp` | 无效的 OTP 验证码 |
| `users.admin.mfa.update_failed` | 更新 MFA 设置失败 |
