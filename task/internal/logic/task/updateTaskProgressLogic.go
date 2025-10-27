// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTaskProgressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 任务进度更新
func NewUpdateTaskProgressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTaskProgressLogic {
	return &UpdateTaskProgressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTaskProgressLogic) UpdateTaskProgress(req *types.UpdateTaskProgressRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
