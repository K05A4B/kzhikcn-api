# 权限认证 API

## 登录

```
POST /api/v1/auth/login
```

**认证**：无需认证

**请求体**：

| 字段名   | 类型   | 必填 | 描述   |
| -------- | ------ | ---- | ------ |
| username | 字符串 | 是   | 用户名 |
| password | 字符串 | 是   | 密码   |

**响应说明**：

- `status`：`authorized` 或 `needMFA`
- `token`：当 `status=authorized` 时返回
- `challengeId`：当 `status=needMFA` 时返回，有效期 120 秒，最多尝试 5 次

**响应示例**：

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "code": 200,
  "message": "认证成功（建议启用MFA）",
  "data": {
    "status": "authorized",
    "token": "<jwt>"
  },
  "meta": {}
}
```

**接口错误代码**：

- `auth.authentication_failed` 认证失败（用户名或密码错误）
- `auth.validate_password_failed` 校验密码异常
- `auth.find_admin_failed` 查询管理员失败
- `auth.token.generate_failed` 生成 Token 失败
- `auth.mfa.create_challenge_failed` 创建 MFA 挑战失败

---

## 验证 TOTP

```
POST /api/v1/auth/mfa/totp
```

**认证**：无需认证

**请求体**：

| 字段名      | 类型   | 必填 | 描述         |
| ----------- | ------ | ---- | ------------ |
| challengeId | 字符串 | 是   | 挑战 ID      |
| otp         | 字符串 | 是   | 一次性验证码 |

**响应示例**：

```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true,
  "code": 200,
  "message": "",
  "data": {
    "token": "<jwt>"
  },
  "meta": {}
}
```

**接口错误代码**：

- `auth.mfa.invalid_challenge` 无效挑战
- `auth.mfa.get_challenge_failed` 获取挑战失败
- `auth.mfa.clean_challenge_failed` 清理挑战失败
- `auth.authentication_failed` 验证失败
- `auth.token.generate_failed` 生成 Token 失败
- `auth.find_admin_failed` 查询管理员失败

---

## 登出

```
POST /api/v1/auth/logout
```

**认证**：需要 JWT 认证

**请求体**：无

**响应**：成功时返回统一响应结构，`data` 为空

**接口错误代码**：

- `auth.token.revoke_failed` 撤销 Token 失败
