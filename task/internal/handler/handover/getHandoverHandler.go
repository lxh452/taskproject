// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handover

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/handover"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 获取交接信息
func GetHandoverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetHandoverRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := handover.NewGetHandoverLogic(r.Context(), svcCtx)
		resp, err := l.GetHandover(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
