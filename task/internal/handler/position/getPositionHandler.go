// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package position

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/position"
	"task_Project/task/internal/svc"
)

// 获取职位信息
func GetPositionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := position.NewGetPositionLogic(r.Context(), svcCtx)
		resp, err := l.GetPosition()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
