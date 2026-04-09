package task

import (
	"encoding/json"
	"fmt"
	"net/http"

	taskLogic "task_Project/task/internal/logic/ai/task"
	"task_Project/task/internal/svc"
)

// StreamAIChatHandler 流式AI对话处理器（直接传递prompt给GLM）
func StreamAIChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Prompt string `json:"prompt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Prompt == "" {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Write([]byte("event: error\ndata: {\"message\": \"解析请求失败或prompt为空\"}\n\n"))
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if svcCtx.GLMService == nil {
			w.Write([]byte("event: error\ndata: {\"message\": \"AI服务未配置\"}\n\n"))
			return
		}

		// 发送开始事件
		sendSSEEvent(w, "start", map[string]interface{}{"message": "开始AI对话"})

		// 收集完整内容
		var fullContent string
		chunkCount := 0

		err := svcCtx.GLMService.StreamCallGLMWithPrompt(r.Context(), req.Prompt, func(chunk string) {
			fullContent += chunk
			chunkCount++
			sendSSEEvent(w, "chunk", map[string]interface{}{
				"content": chunk,
				"index":   chunkCount,
			})
		})

		if err != nil {
			sendSSEEvent(w, "error", map[string]interface{}{"message": err.Error()})
			return
		}

		sendSSEEvent(w, "complete", map[string]interface{}{
			"content": fullContent,
		})
	}
}

func sendSSEEvent(w http.ResponseWriter, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(jsonData))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// StreamGenerateSubtasksHandler 流式生成子任务处理器
func StreamGenerateSubtasksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TaskDescription string                 `json:"taskDescription"`
			TaskID          string                 `json:"taskId,optional"`
			Context         map[string]interface{} `json:"context,optional"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// 以 SSE 格式返回错误
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Write([]byte("event: error\ndata: {\"message\": \"解析请求失败\"}\n\n"))
			return
		}

		// 转换为内部请求格式
		subtaskReq := &taskLogic.GenerateSubtasksRequest{
			TaskID:     req.TaskID,
			TaskTitle:  req.TaskDescription,
			TaskDetail: req.TaskDescription,
			Context:    req.Context,
		}

		l := taskLogic.NewStreamGenerateSubtasksLogic(r.Context(), svcCtx)
		l.StreamGenerateSubtasks(subtaskReq, w)
	}
}
