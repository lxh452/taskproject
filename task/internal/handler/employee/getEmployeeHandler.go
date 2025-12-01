// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package employee

import (
	"net/http"

	"task_Project/task/internal/logic/employee"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取员工信息
func GetEmployeeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetEmployeeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := employee.NewGetEmployeeLogic(r.Context(), svcCtx)
		resp, err := l.GetEmployee(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
