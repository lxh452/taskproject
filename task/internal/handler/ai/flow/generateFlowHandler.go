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

// 生成流程图
func GenerateFlowHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GenerateFlowRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := flow.NewGenerateFlowLogic(r.Context(), svcCtx)
		resp, err := l.GenerateFlow(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
