package middleware

import "strings"

// PublicPaths 不需要 JWT / CSRF / Authz 校验的公开路径（单一来源）
var PublicPaths = []string{
	"/api/v1/auth/login",
	"/api/v1/auth/register",
	"/api/v1/auth/logout",
	"/api/v1/auth/send-code",
	"/api/v1/auth/reset-password",
	"/api/v1/admin/login",
	"/api/v1/company/invite/parse",
}

// AuthzExtraPaths 权限中间件额外放行的路径（已登录但不需要员工角色）
var AuthzExtraPaths = []string{
	"/api/v1/company/create",
	"/api/v1/employee/join",
	"/api/v1/company/list",
	"/api/v1/department/list",
	"/api/v1/position/list",
}

// IsPublicPath 判断路径是否为公开路径
func IsPublicPath(path string) bool {
	for _, p := range PublicPaths {
		if path == p {
			return true
		}
	}
	// 静态文件
	if strings.HasPrefix(path, "/static/") {
		return true
	}
	// admin 路径中的公开接口已在 PublicPaths 中列出，其余 admin 路径不自动放行
	return false
}

// IsAuthzExemptPath 判断路径是否豁免权限校验（公开路径 + 额外放行路径）
func IsAuthzExemptPath(path string) bool {
	if IsPublicPath(path) {
		return true
	}
	for _, p := range AuthzExtraPaths {
		if path == p {
			return true
		}
	}
	// Admin 路径需要权限校验，不豁免
	return false
}
