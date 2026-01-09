package admin

import (
	"net/http"

	"task_Project/task/internal/logic/admin"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取服务器性能指标
func MetricsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := admin.NewMetricsLogic(r.Context(), svcCtx)
		resp, err := l.GetMetrics()
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Response.Error(500, err.Error()))
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
