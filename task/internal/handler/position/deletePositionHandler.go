// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package position

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/position"
	"task_Project/task/internal/svc"
)

// 删除职位
func DeletePositionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := position.NewDeletePositionLogic(r.Context(), svcCtx)
		resp, err := l.DeletePosition()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
