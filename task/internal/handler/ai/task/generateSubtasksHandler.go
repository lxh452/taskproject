// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/ai/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 生成子任务
func GenerateSubtasksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GenerateSubtasksRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 转换为 logic 层请求类型
		logicReq := &task.GenerateSubtasksRequest{
			TaskID:     "",
			TaskTitle:  req.TaskDescription,
			TaskDetail: req.TaskDescription,
			Context:    nil,
		}

		l := task.NewGenerateSubtasksLogic(r.Context(), svcCtx)
		resp, err := l.GenerateSubtasks(logicReq)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
