package checklist

import (
	"net/http"

	"task_Project/task/internal/logic/checklist"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 更新任务清单
func UpdateChecklistHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateChecklistRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Response.ValidationError(err.Error()))
			return
		}

		l := checklist.NewUpdateChecklistLogic(r.Context(), svcCtx)
		resp, err := l.UpdateChecklist(&req)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Response.Error(500, err.Error()))
		} else {
			httpx.OkJsonCtx(r.Context(), w, utils.Response.Success(resp))
		}
	}
}
