package task

import (
	"encoding/json"
	"net/http"

	taskLogic "task_Project/task/internal/logic/ai/task"
	"task_Project/task/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// PolishTaskHandler 任务润色处理器
func PolishTaskHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RawDescription string                 `json:"rawDescription"`
			Context        map[string]interface{} `json:"context,optional"`
		}
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 转换为内部请求格式
		polishReq := &taskLogic.PolishTaskRequest{
			TaskTitle:  req.RawDescription,
			TaskDetail: req.RawDescription,
			PolishType: "clarity",
		}

		l := taskLogic.NewPolishTaskLogic(r.Context(), svcCtx)
		resp, err := l.PolishTask(polishReq)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// GenerateSubtasksHandler 生成子任务处理器
func GenerateSubtasksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TaskDescription string `json:"taskDescription"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 转换为内部请求格式
		subtaskReq := &taskLogic.GenerateSubtasksRequest{
			TaskTitle:  req.TaskDescription,
			TaskDetail: req.TaskDescription,
		}

		l := taskLogic.NewGenerateSubtasksLogic(r.Context(), svcCtx)
		resp, err := l.GenerateSubtasks(subtaskReq)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
