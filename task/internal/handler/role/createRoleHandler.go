// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/role"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 创建角色
func CreateRoleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateRoleRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := role.NewCreateRoleLogic(r.Context(), svcCtx)
		resp, err := l.CreateRole(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
