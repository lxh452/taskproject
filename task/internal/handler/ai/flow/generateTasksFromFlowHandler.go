// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package flow

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/ai/flow"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 根据流程生成任务
func GenerateTasksFromFlowHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GenerateTasksFromFlowRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := flow.NewGenerateTasksFromFlowLogic(r.Context(), svcCtx)
		resp, err := l.GenerateTasksFromFlow(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
