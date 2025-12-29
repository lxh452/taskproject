// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package upload

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/upload"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 获取附件评论列表
func GetAttachmentCommentsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetAttachmentCommentsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := upload.NewGetAttachmentCommentsLogic(r.Context(), svcCtx)
		resp, err := l.GetAttachmentComments(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
