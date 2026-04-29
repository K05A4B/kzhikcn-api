# 当前用户（管理员） API

> 所有接口均需要 JWT 认证。

## 获取当前用户信息

```
GET /api/v1/users/me
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
    "email": "admin@example.com",
    "avatar": "https://example.com/avatar.png",
    "enableMFA": false
  },
  "meta": {}
}
```

**接口错误代码**：

- `users.admin.find_failed` 查询管理员失败

---

## 更新当前用户信息

```
PATCH /api/v1/users/me
```

**请求体**：

| 字段名   | 类型   | 必填 | 描述     |
| -------- | ------ | ---- | -------- |
| username | 字符串 | 否   | 用户名   |
| email    | 字符串 | 否   | 邮箱     |
| avatar   | 字符串 | 否   | 头像 URL |

**响应**：成功时返回统一响应结构，`data` 为空

**接口错误代码**：

- `users.admin.update_info_failed` 更新管理员信息失败

---

## 修改密码

```
PUT /api/v1/users/me/password
```

**请求体**：

| 字段名      | 类型   | 必填 | 描述   |
| ----------- | ------ | ---- | ------ |
| oldPassword | 字符串 | 是   | 原密码 |
| newPassword | 字符串 | 是   | 新密码 |

**响应**：成功时返回统一响应结构，`data` 为空

**接口错误代码**：

- `users.admin.not_found` 管理员不存在
- `users.admin.compare_password_failed` 校验密码失败
- `users.admin.validate_failed` 原密码错误
- `users.admin.change_password_failed` 更新密码失败

---

## 生成 TOTP Secret

```
POST /api/v1/users/me/mfa/totp-secret
```

**请求体**：

| 字段名   | 类型   | 必填 | 描述     |
| -------- | ------ | ---- | -------- |
| password | 字符串 | 是   | 当前密码 |

**响应示例**：

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "code": 200,
  "message": "",
  "data": {
    "secret": "BASE32SECRET",
    "accountName": "admin",
    "issuer": "kzhikcn-api",
    "url": "otpauth://totp/..."
  },
  "meta": {}
}
```

**接口错误代码**：

- `users.admin.not_found` 管理员不存在
- `users.admin.compare_password_failed` 校验密码失败
- `users.admin.totp.generate_failed` 生成 TOTP 失败
- `users.admin.totp.update_secret_failed` 更新 TOTP 密钥失败

---

## 启用 / 禁用 MFA

```
PUT /api/v1/users/me/mfa/enable
PUT /api/v1/users/me/mfa/disable
```

**请求体**：

| 字段名   | 类型   | 必填 | 描述        |
| -------- | ------ | ---- | ----------- |
| password | 字符串 | 是   | 当前密码    |
| otp      | 字符串 | 是   | TOTP 验证码 |

**响应**：成功时返回统一响应结构，`data` 为空

**接口错误代码**：

- `users.admin.not_found` 管理员不存在
- `users.admin.compare_password_failed` 校验密码失败
- `users.admin.mfa.invalid_otp` 无效的 OTP
- `users.admin.mfa.update_failed` 更新 MFA 失败
