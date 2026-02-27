// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package aiagent

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/aiagent"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 执行 Agent
func ExecuteAgentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ExecuteAgentRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := aiagent.NewExecuteAgentLogic(r.Context(), svcCtx)
		resp, err := l.ExecuteAgent(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
