package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey     string        `json:"secretKey"`     // 密钥
	ExpireTime    time.Duration `json:"expireTime"`    // 过期时间
	RefreshTime   time.Duration `json:"refreshTime"`   // 刷新时间
	Issuer        string        `json:"issuer"`        // 签发者
	Audience      string        `json:"audience"`      // 受众
}

// Claims JWT声明
type Claims struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	RealName string `json:"realName"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTMiddleware JWT中间件
type JWTMiddleware struct {
	config JWTConfig
}

// NewJWTMiddleware 创建JWT中间件
func NewJWTMiddleware(config JWTConfig) *JWTMiddleware {
	return &JWTMiddleware{
		config: config,
	}
}

// GenerateToken 生成JWT令牌
func (j *JWTMiddleware) GenerateToken(userID, username, realName, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RealName: realName,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.Issuer,
			Audience:  []string{j.config.Audience},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.ExpireTime)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.SecretKey))
}

// ParseToken 解析JWT令牌
func (j *JWTMiddleware) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken 刷新JWT令牌
func (j *JWTMiddleware) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查是否在刷新时间范围内
	now := time.Now()
	if now.Sub(claims.IssuedAt.Time) < j.config.RefreshTime {
		return "", errors.New("token not ready for refresh")
	}

	// 生成新令牌
	return j.GenerateToken(claims.UserID, claims.Username, claims.RealName, claims.Role)
}

// ValidateToken 验证JWT令牌
func (j *JWTMiddleware) ValidateToken(tokenString string) error {
	_, err := j.ParseToken(tokenString)
	return err
}

// ExtractTokenFromHeader 从请求头中提取令牌
func (j *JWTMiddleware) ExtractTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header not found")
	}

	// 检查Bearer前缀
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", errors.New("token not found")
	}

	return token, nil
}

// Handle JWT中间件处理函数
func (j *JWTMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 提取令牌
		tokenString, err := j.ExtractTokenFromHeader(r)
		if err != nil {
			logx.Errorf("提取JWT令牌失败: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 验证令牌
		claims, err := j.ParseToken(tokenString)
		if err != nil {
			logx.Errorf("JWT令牌验证失败: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 将用户信息添加到请求上下文
		ctx := context.WithValue(r.Context(), "userId", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "realName", claims.RealName)
		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = context.WithValue(ctx, "claims", claims)

		// 继续处理请求
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserID 从上下文中获取用户ID
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("userId").(string)
	return userID, ok
}

// GetUsername 从上下文中获取用户名
func GetUsername(ctx context.Context) (string, bool) {
	username, ok := ctx.Value("username").(string)
	return username, ok
}

// GetRealName 从上下文中获取真实姓名
func GetRealName(ctx context.Context) (string, bool) {
	realName, ok := ctx.Value("realName").(string)
	return realName, ok
}

// GetRole 从上下文中获取角色
func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value("role").(string)
	return role, ok
}

// GetClaims 从上下文中获取声明
func GetClaims(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value("claims").(*Claims)
	return claims, ok
}

// RequireRole 要求特定角色的中间件
func (j *JWTMiddleware) RequireRole(requiredRole string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 先执行JWT验证
			j.Handle(next)(w, r)
			
			// 检查角色
			role, ok := GetRole(r.Context())
			if !ok || role != requiredRole {
				logx.Errorf("用户角色不匹配: 需要 %s, 实际 %s", requiredRole, role)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
	}
}

// RequireAnyRole 要求任意一个角色的中间件
func (j *JWTMiddleware) RequireAnyRole(requiredRoles ...string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 先执行JWT验证
			j.Handle(next)(w, r)
			
			// 检查角色
			role, ok := GetRole(r.Context())
			if !ok {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			hasRole := false
			for _, requiredRole := range requiredRoles {
				if role == requiredRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				logx.Errorf("用户角色不匹配: 需要 %v 中的任意一个, 实际 %s", requiredRoles, role)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}
	}
}
