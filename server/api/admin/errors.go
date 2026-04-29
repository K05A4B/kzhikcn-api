package admin

import "kzhikcn/server/common/hdl"

var (
	ErrAdminNotFound              = hdl.DefineError(404, "没有找到对应管理员", "users.admin.not_found")
	ErrAdminComparePasswordFailed = hdl.DefineError(400, "校验密码时失败", "users.admin.compare_password_failed")
	ErrAdminValidateFailed        = hdl.DefineError(401, "验证失败", "users.admin.validate_failed")
	ErrFindAdminFailed            = hdl.DefineError(500, "查询管理员失败", "users.admin.find_failed")
	ErrChangePasswordFailed       = hdl.DefineError(500, "更新密码失败", "users.admin.change_password_failed")
	ErrUpdateAdminInfoFailed      = hdl.DefineError(500, "更新管理员信息失败", "users.admin.update_info_failed")

	ErrAdminGenerateTOTPFailed     = hdl.DefineError(500, "生成 TOTP 失败", "users.admin.totp.generate_failed")
	ErrAdminUpdateTOTPSecretFailed = hdl.DefineError(500, "更新 TOTP 密钥失败", "users.admin.totp.update_secret_failed")

	ErrAdminInvalidOTP      = hdl.DefineError(400, "无效的 OTP 验证码", "users.admin.mfa.invalid_otp")
	ErrAdminUpdateMFAFailed = hdl.DefineError(500, "更新 MFA 设置失败", "users.admin.mfa.update_failed")
)
