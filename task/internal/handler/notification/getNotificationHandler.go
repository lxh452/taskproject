// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package notification

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/notification"
	"task_Project/task/internal/svc"
)

// 获取通知信息
func GetNotificationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := notification.NewGetNotificationLogic(r.Context(), svcCtx)
		resp, err := l.GetNotification()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
