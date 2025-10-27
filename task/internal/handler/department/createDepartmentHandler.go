// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package department

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/department"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 创建部门
func CreateDepartmentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateDepartmentRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := department.NewCreateDepartmentLogic(r.Context(), svcCtx)
		resp, err := l.CreateDepartment(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
