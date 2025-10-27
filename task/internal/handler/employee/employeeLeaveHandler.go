// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package employee

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/employee"
	"task_Project/task/internal/svc"
)

// 员工离职
func EmployeeLeaveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := employee.NewEmployeeLeaveLogic(r.Context(), svcCtx)
		resp, err := l.EmployeeLeave()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
