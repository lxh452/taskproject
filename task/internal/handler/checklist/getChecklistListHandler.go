package checklist

import (
	"net/http"

	"task_Project/task/internal/logic/checklist"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取任务节点下的所有清单
func GetChecklistListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetChecklistListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Response.ValidationError(err.Error()))
			return
		}

		l := checklist.NewGetChecklistListLogic(r.Context(), svcCtx)
		resp, err := l.GetChecklistList(&req)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Response.Error(500, err.Error()))
		} else {
			httpx.OkJsonCtx(r.Context(), w, utils.Response.Success(resp))
		}
	}
}
