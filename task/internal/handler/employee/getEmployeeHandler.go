// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package employee

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/employee"
	"task_Project/task/internal/svc"
)

// 获取员工信息
func GetEmployeeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := employee.NewGetEmployeeLogic(r.Context(), svcCtx)
		resp, err := l.GetEmployee()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
