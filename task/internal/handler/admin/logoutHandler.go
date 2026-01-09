package admin

import (
	"net/http"

	"task_Project/task/internal/logic/admin"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// AdminLogoutHandler 管理员登出
func AdminLogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminLogoutRequest
		l := admin.NewAdminLogoutLogic(r.Context(), svcCtx)
		resp, err := l.AdminLogout(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
