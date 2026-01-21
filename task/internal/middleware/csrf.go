package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// CSRFConfig CSRF配置
type CSRFConfig struct {
	TokenExpiry time.Duration // Token过期时间，默认2小时
	CookieName  string        // Cookie名称
	HeaderName  string        // Header名称
	Enabled     bool          // 是否启用
}

// DefaultCSRFConfig 默认CSRF配置
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenExpiry: 2 * time.Hour,
		CookieName:  "csrf_token",
		HeaderName:  "X-CSRF-Token",
		Enabled:     true,
	}
}

// CSRFMiddleware CSRF防护中间件
type CSRFMiddleware struct {
	redisClient    *redis.Redis
	config         CSRFConfig
	securityLogger SecurityLogger
}

// NewCSRFMiddleware 创建CSRF中间件
func NewCSRFMiddleware(redisClient *redis.Redis, config CSRFConfig) *CSRFMiddleware {
	return &CSRFMiddleware{
		redisClient: redisClient,
		config:      config,
	}
}

// SetSecurityLogger 设置安全日志记录器
func (c *CSRFMiddleware) SetSecurityLogger(logger SecurityLogger) {
	c.securityLogger = logger
}

// GenerateToken 生成CSRF Token
func (c *CSRFMiddleware) GenerateToken(userID string) (string, error) {
	token := uuid.New().String()
	key := c.getRedisKey(userID)

	err := c.redisClient.Setex(key, token, int(c.config.TokenExpiry.Seconds()))
	if err != nil {
		logx.Errorf("[CSRF] 生成Token失败: %v", err)
		return "", err
	}

	logx.Infof("[CSRF] 为用户 %s 生成Token", userID)
	return token, nil
}

// ValidateToken 验证CSRF Token
func (c *CSRFMiddleware) ValidateToken(userID, token string) bool {
	if token == "" {
		return false
	}

	key := c.getRedisKey(userID)
	storedToken, err := c.redisClient.Get(key)
	if err != nil {
		logx.Errorf("[CSRF] 获取Token失败: %v", err)
		return false
	}

	return storedToken == token
}

// RefreshToken 刷新Token（延长过期时间）
func (c *CSRFMiddleware) RefreshToken(userID string) error {
	key := c.getRedisKey(userID)
	return c.redisClient.Expire(key, int(c.config.TokenExpiry.Seconds()))
}

// RevokeToken 撤销Token
func (c *CSRFMiddleware) RevokeToken(userID string) error {
	key := c.getRedisKey(userID)
	_, err := c.redisClient.Del(key)
	return err
}

func (c *CSRFMiddleware) getRedisKey(userID string) string {
	return "csrf:" + userID
}

// isExemptPath 检查路径是否豁免CSRF验证
func isExemptPath(path string) bool {
	exemptPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/logout",
		"/api/v1/auth/send-code",
		"/api/v1/auth/reset-password",
		"/api/v1/admin/login",
		"/api/v1/company/invite/parse",
	}
	for _, p := range exemptPaths {
		if path == p {
			return true
		}
	}
	return false
}

// isExemptMethod 检查HTTP方法是否豁免CSRF验证
func isExemptMethod(method string) bool {
	exemptMethods := []string{"GET", "HEAD", "OPTIONS"}
	for _, m := range exemptMethods {
		if strings.EqualFold(method, m) {
			return true
		}
	}
	return false
}

// Handle CSRF中间件处理函数
func (c *CSRFMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 如果未启用，直接放行
		if !c.config.Enabled {
			next(w, r)
			return
		}

		path := r.URL.Path
		method := r.Method

		// 豁免的方法和路径
		if isExemptMethod(method) || isExemptPath(path) {
			next(w, r)
			return
		}

		// 获取用户ID
		userID, ok := GetUserID(r.Context())
		if !ok || userID == "" {
			// 未登录用户不需要CSRF验证（会被JWT中间件拦截）
			next(w, r)
			return
		}

		// 从Header获取Token
		token := r.Header.Get(c.config.HeaderName)
		if token == "" {
			// 也尝试从Cookie获取
			cookie, err := r.Cookie(c.config.CookieName)
			if err == nil {
				token = cookie.Value
			}
		}

		// 验证Token
		if !c.ValidateToken(userID, token) {
			ip := getClientIP(r)
			logx.Infof("[CSRF] Token验证失败 - IP: %s, UserID: %s, Path: %s", ip, userID, path)

			// 记录安全日志
			if c.securityLogger != nil {
				c.securityLogger.LogRateLimitExceeded(ip, userID, path, 0)
			}

			c.writeCSRFError(w)
			return
		}

		// 刷新Token过期时间
		c.RefreshToken(userID)

		next(w, r)
	}
}

// writeCSRFError 写入CSRF错误响应
func (c *CSRFMiddleware) writeCSRFError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)

	resp := map[string]interface{}{
		"code": 403,
		"msg":  "请求验证失败，请刷新页面重试",
	}
	json.NewEncoder(w).Encode(resp)
}
