package dashboard

import (
	"net/http"

	"task_Project/task/internal/logic/dashboard"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取仪表盘统计数据
func GetDashboardStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetDashboardStatsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := dashboard.NewGetDashboardStatsLogic(r.Context(), svcCtx)
		resp, err := l.GetDashboardStats(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
