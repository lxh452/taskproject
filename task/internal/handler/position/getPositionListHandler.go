// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package position

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/position"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 获取职位列表
func GetPositionListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PositionListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := position.NewGetPositionListLogic(r.Context(), svcCtx)
		resp, err := l.GetPositionList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
