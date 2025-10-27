// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package employee

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/employee"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 获取员工列表
func GetEmployeeListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.EmployeeListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := employee.NewGetEmployeeListLogic(r.Context(), svcCtx)
		resp, err := l.GetEmployeeList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
