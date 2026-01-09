package admin

import (
	"net/http"

	"task_Project/task/internal/logic/admin"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// LogListHandler 系统日志列表
func LogListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SystemLogListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.WriteJson(w, http.StatusOK, utils.Response.ValidationError(err.Error()))
			return
		}

		l := admin.NewLogListLogic(r.Context(), svcCtx)
		resp, err := l.LogList(&req)
		if err != nil {
			httpx.WriteJson(w, http.StatusOK, utils.Response.InternalError(err.Error()))
		} else {
			httpx.WriteJson(w, http.StatusOK, resp)
		}
	}
}
