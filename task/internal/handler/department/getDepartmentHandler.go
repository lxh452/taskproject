// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package department

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/department"
	"task_Project/task/internal/svc"
)

// 获取部门信息
func GetDepartmentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := department.NewGetDepartmentLogic(r.Context(), svcCtx)
		resp, err := l.GetDepartment()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
