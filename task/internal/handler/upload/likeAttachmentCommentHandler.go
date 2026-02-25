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

// 点赞/取消点赞附件评论
func LikeAttachmentCommentHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LikeCommentRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := upload.NewLikeAttachmentCommentLogic(r.Context(), svcCtx)
		resp, err := l.LikeAttachmentComment(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
