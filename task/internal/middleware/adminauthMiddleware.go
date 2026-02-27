// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package middleware

import (
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"net/http"
)

// Admin Token在Redis中的key前缀
const (
	AdminTokenKeyPrefix = "admin:token:"
)

type AdminAuthMiddleware struct {
	jwtMiddleware *JWTMiddleware
	redisClient   *redis.Redis
}

func NewAdminAuthMiddleware(jwtMiddleware *JWTMiddleware, redisClient *redis.Redis) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{
		jwtMiddleware: jwtMiddleware,
		redisClient:   redisClient,
	}
}

// ValidateAdminTokenWithRedis 验证管理员Token是否在Redis中有效
func (m *AdminAuthMiddleware) ValidateAdminTokenWithRedis(tokenString string, adminID string) error {
	if m.redisClient == nil {
		return nil
	}

	tokenKey := fmt.Sprintf("%s%s", AdminTokenKeyPrefix, adminID)
	storedToken, err := m.redisClient.Get(tokenKey)
	if err != nil {
		logx.Infof("从Redis获取Admin Token失败: %v, 跳过Redis校验", err)
		return nil
	}

	if storedToken == "" {
		logx.Infof("Admin Token不在Redis中，跳过Redis校验, adminId=%s", adminID)
		return nil
	}

	if storedToken != tokenString {
		return fmt.Errorf("admin token mismatch, may have logged in from another device")
	}

	return nil
}

// Handle 管理员认证中间件处理函数
func (m *AdminAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 提取令牌
		tokenString, err := m.jwtMiddleware.ExtractTokenFromHeader(r)
		if err != nil {
			logx.Errorf("提取JWT令牌失败: %v", err)
			writeJSONError(w, http.StatusUnauthorized, "未授权访问")
			return
		}

		// 验证令牌
		claims, err := m.jwtMiddleware.ParseToken(tokenString)
		if err != nil {
			logx.Errorf("JWT令牌验证失败: %v", err)
			writeJSONError(w, http.StatusUnauthorized, "令牌无效或已过期")
			return
		}

		// 检查是否是管理员角色
		if claims.Role != "admin" && claims.Role != "super_admin" {
			logx.Errorf("非管理员角色尝试访问管理端: role=%s, userId=%s", claims.Role, claims.UserID)
			writeJSONError(w, http.StatusForbidden, "权限不足，仅管理员可访问")
			return
		}

		// 验证Token是否在Redis中有效
		if err := m.ValidateAdminTokenWithRedis(tokenString, claims.UserID); err != nil {
			logx.Errorf("Redis Admin Token验证失败: %v, adminId=%s", err, claims.UserID)
			writeJSONError(w, http.StatusUnauthorized, "令牌无效或已过期，请重新登录")
			return
		}

		// 将管理员信息添加到请求上下文（使用类型安全的 key）
		ctx := context.WithValue(r.Context(), CtxKeyUserID, claims.UserID)
		ctx = context.WithValue(ctx, CtxKeyAdminID, claims.UserID)
		ctx = context.WithValue(ctx, CtxKeyUsername, claims.Username)
		ctx = context.WithValue(ctx, CtxKeyRealName, claims.RealName)
		ctx = context.WithValue(ctx, CtxKeyRole, claims.Role)
		ctx = context.WithValue(ctx, CtxKeyClaims, claims)

		// 继续处理请求
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
