package middleware

import (
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// NewCorsMiddleware returns a global CORS middleware.
// origins: allow list, pass empty or nil to allow all origins.
func NewCorsMiddleware(origins []string) func(http.HandlerFunc) http.HandlerFunc {
	allowAll := len(origins) == 0
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			reqOrigin := r.Header.Get("Origin")

			// 处理预检请求（OPTIONS）- 必须在所有其他处理之前
			if r.Method == http.MethodOptions {
				logx.Infof("收到 OPTIONS 预检请求: %s, Origin: %s", r.URL.Path, reqOrigin)
				// 设置CORS头
				if allowAll {
					// 允许所有源，但注意不能同时使用 "*" 和 credentials
					if reqOrigin != "" {
						h.Set("Access-Control-Allow-Origin", reqOrigin)
						h.Set("Access-Control-Allow-Credentials", "true")
					} else {
						h.Set("Access-Control-Allow-Origin", "*")
					}
				} else {
					// 检查请求源是否在白名单中
					allowed := false
					for _, o := range origins {
						if o == reqOrigin {
							allowed = true
							break
						}
					}
					if allowed && reqOrigin != "" {
						h.Set("Access-Control-Allow-Origin", reqOrigin)
						h.Set("Access-Control-Allow-Credentials", "true")
					} else if reqOrigin == "" {
						// 如果没有Origin头，可能是同源请求，允许
						h.Set("Access-Control-Allow-Origin", "*")
					} else {
						// 不在白名单中的源，开发环境允许通过（生产环境可以改为拒绝）
						logx.Infof("CORS OPTIONS: 请求源 %s 不在白名单中，但允许通过（开发模式）", reqOrigin)
						h.Set("Access-Control-Allow-Origin", reqOrigin)
						h.Set("Access-Control-Allow-Credentials", "true")
					}
				}
				h.Set("Vary", "Origin")
				h.Set("Access-Control-Allow-Headers", strings.Join([]string{
					"Authorization",
					"Content-Type",
					"Accept",
					"X-Requested-With",
				}, ", "))
				h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
				h.Set("Access-Control-Max-Age", "86400")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// 处理实际请求
			if allowAll {
				if reqOrigin != "" {
					h.Set("Access-Control-Allow-Origin", reqOrigin)
					h.Set("Access-Control-Allow-Credentials", "true")
				} else {
					h.Set("Access-Control-Allow-Origin", "*")
				}
			} else {
				// 检查请求源是否在白名单中
				allowed := false
				for _, o := range origins {
					if o == reqOrigin {
						allowed = true
						break
					}
				}
				if allowed && reqOrigin != "" {
					h.Set("Access-Control-Allow-Origin", reqOrigin)
					h.Set("Access-Control-Allow-Credentials", "true")
				} else if reqOrigin != "" {
					// 不在白名单中，但仍然设置CORS头以便调试（生产环境可以改为拒绝）
					logx.Infof("CORS: 请求源 %s 不在白名单中，但允许通过", reqOrigin)
					h.Set("Access-Control-Allow-Origin", reqOrigin)
					h.Set("Access-Control-Allow-Credentials", "true")
				} else {
					// 没有Origin头，可能是同源请求
					h.Set("Access-Control-Allow-Origin", "*")
				}
			}
			h.Set("Vary", "Origin")

			next(w, r)
		}
	}
}
