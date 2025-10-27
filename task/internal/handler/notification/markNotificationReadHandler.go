// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package notification

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/notification"
	"task_Project/task/internal/svc"
)

// 标记通知为已读
func MarkNotificationReadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := notification.NewMarkNotificationReadLogic(r.Context(), svcCtx)
		resp, err := l.MarkNotificationRead()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
