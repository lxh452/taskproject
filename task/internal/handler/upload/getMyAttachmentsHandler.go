package upload

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/upload"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 获取我的附件列表
func GetMyAttachmentsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetMyAttachmentsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := upload.NewGetMyAttachmentsLogic(r.Context(), svcCtx)
		resp, err := l.GetMyAttachments(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
