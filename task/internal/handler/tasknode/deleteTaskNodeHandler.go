// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tasknode

import (
	"net/http"

	"task_Project/task/internal/logic/tasknode"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 删除任务节点
func DeleteTaskNodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteTaskNodeRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := tasknode.NewDeleteTaskNodeLogic(r.Context(), svcCtx)
		resp, err := l.DeleteTaskNode(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
