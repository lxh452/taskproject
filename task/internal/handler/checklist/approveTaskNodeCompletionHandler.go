// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package checklist

import (
	"net/http"

	"task_Project/task/internal/logic/checklist"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 审批任务节点完成
func ApproveTaskNodeCompletionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ApproveTaskNodeCompletionRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := checklist.NewApproveTaskNodeCompletionLogic(r.Context(), svcCtx)
		resp, err := l.ApproveTaskNodeCompletion(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
