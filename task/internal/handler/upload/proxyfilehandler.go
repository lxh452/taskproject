package upload

import (
	"context"
	"net/http"

	"task_Project/task/internal/logic/upload"
	"task_Project/task/internal/svc"
)

// 代理文件内容（解决CORS问题）
func ProxyFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取token并验证，将用户信息放入context
		authHeader := r.Header.Get("Authorization")
		ctx := r.Context()

		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token := authHeader[7:]
			// 解析token获取用户信息
			if claims, err := svcCtx.JWTMiddleware.ParseToken(token); err == nil {
				// 验证token是否在Redis中有效
				if err := svcCtx.JWTMiddleware.ValidateTokenWithRedis(token, claims.UserID); err == nil {
					// 将用户信息添加到context（与JWT中间件保持一致）
					ctx = context.WithValue(ctx, "userId", claims.UserID)
					ctx = context.WithValue(ctx, "username", claims.Username)
					ctx = context.WithValue(ctx, "realName", claims.RealName)
					ctx = context.WithValue(ctx, "role", claims.Role)
					ctx = context.WithValue(ctx, "employeeId", claims.EmployeeID)
					ctx = context.WithValue(ctx, "companyId", claims.CompanyID)
				}
			}
		}

		l := upload.NewProxyFileLogic(ctx, svcCtx)
		l.ProxyFile(w, r)
	}
}
