// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package notification

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNotificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取通知信息
func NewGetNotificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNotificationLogic {
	return &GetNotificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNotificationLogic) GetNotification(req *types.GetNotificationRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
