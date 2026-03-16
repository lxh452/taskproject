// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package flow

import (
	"fmt"
	"net/http"

	"task_Project/task/internal/logic/ai/flow"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 生成设计方案
func SuggestDesignsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FlowDesignRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 转换为 logic 层请求类型
		logicReq := &flow.SuggestDesignsRequest{
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
			logicReq.Description = desc
		}

		l := flow.NewSuggestDesignsLogic(r.Context(), svcCtx)
		resp, err := l.SuggestDesigns(logicReq)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
