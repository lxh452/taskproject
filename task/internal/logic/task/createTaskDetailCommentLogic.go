// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTaskDetailCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建任务评论
func NewCreateTaskDetailCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTaskDetailCommentLogic {
	return &CreateTaskDetailCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTaskDetailCommentLogic) CreateTaskDetailComment(req *types.CreateTaskDetailComment) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
