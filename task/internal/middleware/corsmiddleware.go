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
	// 构建 set 加速查找
	originSet := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		originSet[o] = struct{}{}
	}

	isAllowed := func(origin string) bool {
		if allowAll {
			return true
		}
		_, ok := originSet[origin]
		return ok
	}

	setCORSHeaders := func(h http.Header, origin string) {
		if origin != "" {
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
		} else if allowAll {
			h.Set("Access-Control-Allow-Origin", "*")
		}
		h.Set("Vary", "Origin")
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			reqOrigin := r.Header.Get("Origin")

			// 预检请求
			if r.Method == http.MethodOptions {
				if reqOrigin == "" || isAllowed(reqOrigin) {
					setCORSHeaders(h, reqOrigin)
					h.Set("Access-Control-Allow-Headers", strings.Join([]string{
						"Authorization", "Content-Type", "Accept",
						"X-Requested-With", "X-CSRF-Token",
					}, ", "))
					h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
					h.Set("Access-Control-Max-Age", "86400")
					w.WriteHeader(http.StatusNoContent)
				} else {
					logx.Infof("[CORS] 拒绝预检请求，来源不在白名单: %s", reqOrigin)
					w.WriteHeader(http.StatusForbidden)
				}
				return
			}

			// 实际请求
			if reqOrigin == "" || isAllowed(reqOrigin) {
				setCORSHeaders(h, reqOrigin)
				next(w, r)
			} else {
				logx.Infof("[CORS] 拒绝请求，来源不在白名单: %s", reqOrigin)
				http.Error(w, "CORS origin not allowed", http.StatusForbidden)
			}
		}
	}
}
