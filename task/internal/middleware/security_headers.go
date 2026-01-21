package middleware

import (
	"net/http"
)

// SecurityHeadersMiddleware 安全响应头中间件
// 添加各种安全相关的HTTP响应头，防止XSS、点击劫持等攻击
type SecurityHeadersMiddleware struct {
	// CSPDirectives Content-Security-Policy指令
	CSPDirectives string
}

// NewSecurityHeadersMiddleware 创建安全响应头中间件
func NewSecurityHeadersMiddleware() *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{
		// 默认CSP策略：允许同源资源，内联样式（Vue需要），以及常见CDN
		CSPDirectives: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' https:; frame-ancestors 'none'",
	}
}

// Handle 处理请求，添加安全响应头
func (m *SecurityHeadersMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// X-Content-Type-Options: 防止MIME类型嗅探
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options: 防止点击劫持
		w.Header().Set("X-Frame-Options", "DENY")

		// X-XSS-Protection: 启用浏览器XSS过滤器（虽然现代浏览器已弃用，但仍有兼容价值）
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy: 控制Referer头的发送
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content-Security-Policy: 内容安全策略
		if m.CSPDirectives != "" {
			w.Header().Set("Content-Security-Policy", m.CSPDirectives)
		}

		// Permissions-Policy: 限制浏览器功能（原Feature-Policy）
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Strict-Transport-Security: 强制HTTPS（仅在生产环境启用）
		// 注意：这个头只在HTTPS连接时有效
		// w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next(w, r)
	}
}
