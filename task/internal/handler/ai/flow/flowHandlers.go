package flow

import (
	"encoding/json"
	"fmt"
	"net/http"

	flowLogic "task_Project/task/internal/logic/ai/flow"
	"task_Project/task/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// SuggestDesignsHandler 流程设计建议处理器
func SuggestDesignsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Tasks       []string               `json:"tasks"`
			Constraints map[string]interface{} `json:"constraints,optional"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 转换为内部请求格式
		designReq := &flowLogic.SuggestDesignsRequest{
			FlowType:    "task_flow",
			Description: "",
			Context:     req.Constraints,
		}

		// 如果有任务列表，拼接为描述
		if len(req.Tasks) > 0 {
			desc := "任务列表:\n"
			for i, task := range req.Tasks {
				desc += fmt.Sprintf("%d. %s\n", i+1, task)
			}
			designReq.Description = desc
		}

		l := flowLogic.NewSuggestDesignsLogic(r.Context(), svcCtx)
		resp, err := l.SuggestDesigns(designReq)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
