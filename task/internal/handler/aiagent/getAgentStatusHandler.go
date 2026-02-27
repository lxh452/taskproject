// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package aiagent

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/aiagent"
	"task_Project/task/internal/svc"
)

// 获取 Agent 状态
func GetAgentStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := aiagent.NewGetAgentStatusLogic(r.Context(), svcCtx)
		resp, err := l.GetAgentStatus()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
