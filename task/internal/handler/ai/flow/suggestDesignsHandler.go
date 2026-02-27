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

// 生成设计方案
func SuggestDesignsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FlowDesignRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := flow.NewSuggestDesignsLogic(r.Context(), svcCtx)
		resp, err := l.SuggestDesigns(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
