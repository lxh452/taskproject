package admin

import (
	"net/http"

	"task_Project/task/internal/logic/admin"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// EnableCompanyHandler 启用公司
func EnableCompanyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.EnableCompanyRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Response.ValidationError(err.Error()))
			return
		}

		l := admin.NewEnableCompanyLogic(r.Context(), svcCtx)
		resp, err := l.EnableCompany(&req)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Response.Error(500, err.Error()))
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
