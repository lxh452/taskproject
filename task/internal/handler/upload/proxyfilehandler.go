package upload

import (
	"net/http"

	"task_Project/task/internal/logic/upload"
	"task_Project/task/internal/svc"
)

// 代理文件内容（解决CORS问题）
func ProxyFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := upload.NewProxyFileLogic(r.Context(), svcCtx)
		l.ProxyFile(w, r)
	}
}
