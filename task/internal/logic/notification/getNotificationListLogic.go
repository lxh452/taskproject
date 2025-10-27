// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package notification

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNotificationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取通知列表
func NewGetNotificationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNotificationListLogic {
	return &GetNotificationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNotificationListLogic) GetNotificationList(req *types.NotificationListRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
