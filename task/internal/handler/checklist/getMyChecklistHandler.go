package checklist

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/checklist"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"
)

// 获取我的清单列表
func GetMyChecklistHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetMyChecklistRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Response.ValidationError(err.Error()))
			return
		}

		l := checklist.NewGetMyChecklistLogic(r.Context(), svcCtx)
		resp, err := l.GetMyChecklist(&req)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Response.Error(500, err.Error()))
		} else {
			httpx.OkJsonCtx(r.Context(), w, utils.Response.Success(resp))
		}
	}
}
