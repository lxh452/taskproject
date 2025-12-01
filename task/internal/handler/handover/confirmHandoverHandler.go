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

// 确认交接
func ConfirmHandoverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ConfirmHandoverRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := handover.NewConfirmHandoverLogic(r.Context(), svcCtx)
		resp, err := l.ConfirmHandover(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
