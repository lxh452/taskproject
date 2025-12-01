// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"net/http"

	"task_Project/task/internal/logic/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取任务信息
func GetTaskHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetTaskRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := task.NewGetTaskLogic(r.Context(), svcCtx)
		resp, err := l.GetTask(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
