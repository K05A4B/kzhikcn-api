package auth

import "kzhikcn/server/common/hdl"

var (
	ErrAuthenticationFailed   = hdl.DefineError(401, "认证失败", "auth.authentication_failed")
	ErrValidatePasswordFailed = hdl.DefineError(500, "校验密码时出错", "auth.validate_password_failed")
	ErrFindAdminFailed        = hdl.DefineError(500, "查找管理员时出错", "auth.find_admin_failed")

	ErrRevokeTokenFailed   = hdl.DefineError(500, "注销token时出错", "auth.token.revoke_failed")
	ErrGenerateTokenFailed = hdl.DefineError(500, "生成token时出错", "auth.token.generate_failed")

	ErrCreateChallengeFailed = hdl.DefineError(500, "创建MFA挑战时出错", "auth.mfa.create_challenge_failed")
	ErrCleanChallengeFailed  = hdl.DefineError(500, "清除MFA挑战时出错", "auth.mfa.clean_challenge_failed")
	ErrInvalidChallenge      = hdl.DefineError(400, "无效的MFA挑战", "auth.mfa.invalid_challenge")
	ErrGetChallengeFailed    = hdl.DefineError(500, "获取MFA挑战时出错", "auth.mfa.get_challenge_failed")
)
