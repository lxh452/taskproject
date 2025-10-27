// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package notification

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/notification"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 创建通知
func CreateNotificationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateNotificationRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := notification.NewCreateNotificationLogic(r.Context(), svcCtx)
		resp, err := l.CreateNotification(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
