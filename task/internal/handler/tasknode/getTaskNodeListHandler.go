// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tasknode

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/tasknode"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 获取任务节点列表
func GetTaskNodeListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TaskNodeListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := tasknode.NewGetTaskNodeListLogic(r.Context(), svcCtx)
		resp, err := l.GetTaskNodeList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
