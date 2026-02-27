// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	_ "task_Project/task/internal/logic/ai/task"
	taskLogic "task_Project/task/internal/logic/ai/task"
	"task_Project/task/internal/svc"
	_ "task_Project/task/internal/types"
)

// 润色任务
func PolishTaskHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RawDescription string                 `json:"rawDescription"`
			PolishType     string                 `json:"polishType,optional"`
			Context        map[string]interface{} `json:"context,optional"`
		}
		if err := httpx.Parse(r, &req); err != nil {
			// 即使是错误也要以 SSE 格式返回
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Write([]byte("event: error\ndata: {\"message\": \"解析请求失败\"}\n\n"))
			return
		}

		// 转换为内部请求格式
		polishReq := &taskLogic.PolishTaskRequest{
			TaskTitle:  req.RawDescription,
			TaskDetail: req.RawDescription,
			PolishType: req.PolishType,
		}
		if polishReq.PolishType == "" {
			polishReq.PolishType = "clarity"
		}

		l := taskLogic.NewStreamPolishTaskLogic(r.Context(), svcCtx)
		l.StreamPolishTask(polishReq, w)
	}
}
