// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tasknode

import (
	"encoding/json"
	"net/http"

	"task_Project/task/internal/logic/tasknode"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取用户的任务节点信息
func GetUserTaskNodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PageReq

		// 尝试从 JSON body 解析（POST 请求）
		if r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				// 如果 JSON 解析失败，尝试从 query 参数解析
				if err := httpx.Parse(r, &req); err != nil {
					httpx.ErrorCtx(r.Context(), w, err)
					return
				}
			}
		} else {
			// GET 请求从 query 参数解析
			if err := httpx.Parse(r, &req); err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}
		}

		l := tasknode.NewGetUserTaskNodeLogic(r.Context(), svcCtx)
		resp, err := l.GetUserTaskNode(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
