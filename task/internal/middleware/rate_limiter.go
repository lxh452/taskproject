package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// RateLimiterConfig 限流配置
type RateLimiterConfig struct {
	LoginLimit    int           `json:"loginLimit"`    // 登录接口限制: 次数/窗口
	LoginWindow   time.Duration `json:"loginWindow"`   // 登录限流窗口
	APILimit      int           `json:"apiLimit"`      // 普通API限制: 次数/窗口
	APIWindow     time.Duration `json:"apiWindow"`     // API限流窗口
	BurstSize     int           `json:"burstSize"`     // 突发容量
	BlockDuration time.Duration `json:"blockDuration"` // 封禁时长
}

// DefaultRateLimiterConfig 默认限流配置
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		LoginLimit:    10,               // 登录: 10次/分钟
		LoginWindow:   time.Minute,      // 1分钟窗口
		APILimit:      100,              // API: 100次/分钟
		APIWindow:     time.Minute,      // 1分钟窗口
		BurstSize:     20,               // 突发20次
		BlockDuration: 15 * time.Minute, // 封禁15分钟
	}
}

// RateLimiter 限流中间件
type RateLimiter struct {
	redisClient    *redis.Redis
	config         RateLimiterConfig
	securityLogger SecurityLogger
}

// SecurityLogger 安全日志接口
type SecurityLogger interface {
	LogRateLimitExceeded(ip, userID, path string, limit int)
}

// NewRateLimiter 创建限流中间件
func NewRateLimiter(redisClient *redis.Redis, config RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		config:      config,
	}
}

// SetSecurityLogger 设置安全日志记录器
func (r *RateLimiter) SetSecurityLogger(logger SecurityLogger) {
	r.securityLogger = logger
}

// RateLimitResponse 限流响应
type RateLimitResponse struct {
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	RetryAfter int    `json:"retryAfter"` // 重试等待秒数
}

// getClientIP 获取客户端真实IP
func getClientIP(r *http.Request) string {
	// 优先从X-Forwarded-For获取
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 其次从X-Real-IP获取
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// 最后从RemoteAddr获取
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// isLoginPath 判断是否是登录相关路径
func isLoginPath(path string) bool {
	loginPaths := []string{
		"/auth/login",
		"/auth/register",
		"/admin/login",
	}
	for _, p := range loginPaths {
		if strings.Contains(path, p) {
			return true
		}
	}
	return false
}

// checkRateLimit 检查限流 (滑动窗口算法)
func (r *RateLimiter) checkRateLimit(key string, limit int, window time.Duration) (bool, int, error) {
	if r.redisClient == nil {
		return true, 0, nil // 没有Redis则跳过限流
	}

	now := time.Now().UnixMilli()
	windowStart := now - window.Milliseconds()

	// 使用Redis ZSET实现滑动窗口
	// 1. 移除窗口外的记录
	_, err := r.redisClient.Zremrangebyscore(key, 0, windowStart)
	if err != nil {
		logx.Errorf("[RateLimiter] 清理过期记录失败: %v", err)
		return true, 0, nil // Redis错误时放行
	}

	// 2. 获取当前窗口内的请求数
	count, err := r.redisClient.Zcard(key)
	if err != nil {
		logx.Errorf("[RateLimiter] 获取请求计数失败: %v", err)
		return true, 0, nil
	}

	// 3. 检查是否超过限制
	if int(count) >= limit {
		// 计算需要等待的时间
		oldestScore, err := r.redisClient.ZrangeWithScores(key, 0, 0)
		if err == nil && len(oldestScore) > 0 {
			oldestTime := int64(oldestScore[0].Score)
			waitTime := int((oldestTime + window.Milliseconds() - now) / 1000)
			if waitTime < 1 {
				waitTime = 1
			}
			return false, waitTime, nil
		}
		return false, int(window.Seconds()), nil
	}

	// 4. 添加当前请求记录
	_, err = r.redisClient.Zadd(key, now, fmt.Sprintf("%d", now))
	if err != nil {
		logx.Errorf("[RateLimiter] 添加请求记录失败: %v", err)
	}

	// 5. 设置key过期时间
	r.redisClient.Expire(key, int(window.Seconds())+10)

	return true, 0, nil
}

// isBlocked 检查IP是否被封禁
func (r *RateLimiter) isBlocked(ip string) bool {
	if r.redisClient == nil {
		return false
	}

	key := fmt.Sprintf("rate:blocked:%s", ip)
	exists, _ := r.redisClient.Exists(key)
	return exists
}

// blockIP 封禁IP
func (r *RateLimiter) blockIP(ip string) {
	if r.redisClient == nil {
		return
	}

	key := fmt.Sprintf("rate:blocked:%s", ip)
	r.redisClient.Setex(key, "1", int(r.config.BlockDuration.Seconds()))
	logx.Infof("[RateLimiter] IP已被封禁: %s, 时长: %v", ip, r.config.BlockDuration)
}

// Handle 限流中间件处理函数
func (r *RateLimiter) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ip := getClientIP(req)
		path := req.URL.Path

		// 检查IP是否被封禁
		if r.isBlocked(ip) {
			r.writeRateLimitResponse(w, int(r.config.BlockDuration.Seconds()), "IP已被临时封禁，请稍后重试")
			return
		}

		var key string
		var limit int
		var window time.Duration

		if isLoginPath(path) {
			// 登录接口使用IP限流
			key = fmt.Sprintf("rate:login:%s", ip)
			limit = r.config.LoginLimit
			window = r.config.LoginWindow
		} else {
			// 普通API使用IP限流（如果有用户ID则使用用户ID）
			userID, ok := GetUserID(req.Context())
			if ok && userID != "" {
				key = fmt.Sprintf("rate:api:%s", userID)
			} else {
				key = fmt.Sprintf("rate:api:ip:%s", ip)
			}
			limit = r.config.APILimit
			window = r.config.APIWindow
		}

		allowed, retryAfter, err := r.checkRateLimit(key, limit, window)
		if err != nil {
			logx.Errorf("[RateLimiter] 限流检查错误: %v", err)
			next(w, req)
			return
		}

		if !allowed {
			// 记录限流日志
			userID, _ := GetUserID(req.Context())
			logx.Infof("[RateLimiter] 限流触发 - IP: %s, UserID: %s, Path: %s, Limit: %d", ip, userID, path, limit)

			// 记录安全日志
			if r.securityLogger != nil {
				r.securityLogger.LogRateLimitExceeded(ip, userID, path, limit)
			}

			// 如果是登录接口且频繁触发，考虑封禁IP
			if isLoginPath(path) {
				// 检查短时间内触发限流的次数
				blockKey := fmt.Sprintf("rate:block_count:%s", ip)
				count, _ := r.redisClient.Incr(blockKey)
				r.redisClient.Expire(blockKey, 300) // 5分钟内

				if count >= 3 {
					r.blockIP(ip)
				}
			}

			r.writeRateLimitResponse(w, retryAfter, "请求过于频繁，请稍后重试")
			return
		}

		next(w, req)
	}
}

// writeRateLimitResponse 写入限流响应
func (r *RateLimiter) writeRateLimitResponse(w http.ResponseWriter, retryAfter int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
	w.WriteHeader(http.StatusTooManyRequests)

	resp := RateLimitResponse{
		Code:       429,
		Msg:        msg,
		RetryAfter: retryAfter,
	}

	json.NewEncoder(w).Encode(resp)
}

// LoginRateLimiter 专门用于登录接口的限流中间件
func (r *RateLimiter) LoginRateLimiter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		ip := getClientIP(req)

		// 检查IP是否被封禁
		if r.isBlocked(ip) {
			r.writeRateLimitResponse(w, int(r.config.BlockDuration.Seconds()), "IP已被临时封禁，请稍后重试")
			return
		}

		key := fmt.Sprintf("rate:login:%s", ip)
		allowed, retryAfter, _ := r.checkRateLimit(key, r.config.LoginLimit, r.config.LoginWindow)

		if !allowed {
			logx.Infof("[RateLimiter] 登录限流触发 - IP: %s", ip)

			// 记录安全日志
			if r.securityLogger != nil {
				r.securityLogger.LogRateLimitExceeded(ip, "", req.URL.Path, r.config.LoginLimit)
			}

			r.writeRateLimitResponse(w, retryAfter, "登录请求过于频繁，请稍后重试")
			return
		}

		next(w, req)
	}
}
