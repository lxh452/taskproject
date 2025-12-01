// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/company"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 获取公司信息
func GetCompanyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetCompanyRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := company.NewGetCompanyLogic(r.Context(), svcCtx)
		resp, err := l.GetCompany(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
