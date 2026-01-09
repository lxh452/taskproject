package admin

import (
	"net/http"

	"task_Project/task/internal/logic/admin"
	"task_Project/task/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取平台统计概览
func GetPlatformStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := admin.NewGetPlatformStatsLogic(r.Context(), svcCtx)
		resp, err := l.GetPlatformStats()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
