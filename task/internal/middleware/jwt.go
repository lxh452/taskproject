package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

// Token在Redis中的key前缀
const (
	TokenKeyPrefix = "auth:token:"
)

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey   string        `json:"secretKey"`   // 密钥
	ExpireTime  time.Duration `json:"expireTime"`  // 过期时间
	RefreshTime time.Duration `json:"refreshTime"` // 刷新时间
	Issuer      string        `json:"issuer"`      // 签发者
	Audience    string        `json:"audience"`    // 受众
}

// Claims JWT声明
type Claims struct {
	UserID     string `json:"userId"`
	Username   string `json:"username"`
	RealName   string `json:"realName"`
	Role       string `json:"role"`
	EmployeeID string `json:"employeeId"` // 员工ID（如果已加入公司）
	CompanyID  string `json:"companyId"`  // 公司ID（如果已加入公司）
	jwt.RegisteredClaims
}

// JWTMiddleware JWT中间件
type JWTMiddleware struct {
	config      JWTConfig
	redisClient *redis.Redis
}

// NewJWTMiddleware 创建JWT中间件
func NewJWTMiddleware(config JWTConfig) *JWTMiddleware {
	return &JWTMiddleware{
		config: config,
	}
}

// SetRedisClient 设置Redis客户端（用于Token校验）
func (j *JWTMiddleware) SetRedisClient(client *redis.Redis) {
	j.redisClient = client
}

// ValidateTokenWithRedis 验证Token是否在Redis中有效
func (j *JWTMiddleware) ValidateTokenWithRedis(tokenString string, userID string) error {
	if j.redisClient == nil {
		// 如果没有配置Redis，跳过Redis校验
		return nil
	}

	tokenKey := fmt.Sprintf("%s%s", TokenKeyPrefix, userID)
	storedToken, err := j.redisClient.Get(tokenKey)
	if err != nil {
		// Redis 获取失败时，只记录日志，不阻止请求（JWT本身已验证通过）
		logx.Infof("从Redis获取Token失败（可能是Redis重启）: %v, 跳过Redis校验", err)
		return nil
	}

	if storedToken == "" {
		// Token不在Redis中，可能是Redis重启导致数据丢失，允许通过（JWT本身已验证）
		logx.Infof("Token不在Redis中（可能是Redis重启），跳过Redis校验, userId=%s", userID)
		return nil
	}

	if storedToken != tokenString {
		// 只有当Redis中有token且不匹配时才拒绝（说明用户在其他设备登录了）
		return errors.New("token mismatch, may have logged in from another device")
	}

	return nil
}

// GenerateToken 生成JWT令牌
func (j *JWTMiddleware) GenerateToken(userID, username, realName, role string) (string, error) {
	return j.GenerateTokenWithEmployee(userID, username, realName, role, "", "")
}

// GenerateTokenWithEmployee 生成带员工信息的JWT令牌
func (j *JWTMiddleware) GenerateTokenWithEmployee(userID, username, realName, role, employeeID, companyID string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:     userID,
		Username:   username,
		RealName:   realName,
		Role:       role,
		EmployeeID: employeeID,
		CompanyID:  companyID,
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

	// 生成新令牌（包含员工信息）
	return j.GenerateTokenWithEmployee(claims.UserID, claims.Username, claims.RealName, claims.Role, claims.EmployeeID, claims.CompanyID)
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

		// 验证Token是否在Redis中有效（确保登录一致性）
		if err := j.ValidateTokenWithRedis(tokenString, claims.UserID); err != nil {
			logx.Errorf("Redis Token验证失败: %v, userId=%s", err, claims.UserID)
			http.Error(w, "Token invalid or expired, please login again", http.StatusUnauthorized)
			return
		}

		// 将用户信息添加到请求上下文
		ctx := context.WithValue(r.Context(), "userId", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "realName", claims.RealName)
		ctx = context.WithValue(ctx, "role", claims.Role)
		ctx = context.WithValue(ctx, "employeeId", claims.EmployeeID)
		ctx = context.WithValue(ctx, "companyId", claims.CompanyID)
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

// GetEmployeeID 从上下文中获取员工ID
func GetEmployeeID(ctx context.Context) (string, bool) {
	employeeID, ok := ctx.Value("employeeId").(string)
	return employeeID, ok
}

// GetCompanyID 从上下文中获取公司ID
func GetCompanyID(ctx context.Context) (string, bool) {
	companyID, ok := ctx.Value("companyId").(string)
	return companyID, ok
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
