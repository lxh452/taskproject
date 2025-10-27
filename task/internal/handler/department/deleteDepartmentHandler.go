// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package department

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/department"
	"task_Project/task/internal/svc"
)

// 删除部门
func DeleteDepartmentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := department.NewDeleteDepartmentLogic(r.Context(), svcCtx)
		resp, err := l.DeleteDepartment()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
