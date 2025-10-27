// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package notification

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateNotificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建通知
func NewCreateNotificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateNotificationLogic {
	return &CreateNotificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateNotificationLogic) CreateNotification(req *types.CreateNotificationRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
