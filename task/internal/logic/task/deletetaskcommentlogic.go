// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTaskCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除任务评论
func NewDeleteTaskCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTaskCommentLogic {
	return &DeleteTaskCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTaskCommentLogic) DeleteTaskComment(req *types.DeleteTaskCommentRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
