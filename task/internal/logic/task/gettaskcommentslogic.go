// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskCommentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取任务评论列表
func NewGetTaskCommentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskCommentsLogic {
	return &GetTaskCommentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTaskCommentsLogic) GetTaskComments(req *types.GetTaskCommentsRequest) (resp *types.BaseResponse, err error) {
	// 使用 TaskCommentLogic 来处理评论列表查询
	commentLogic := NewTaskCommentLogic(l.ctx, l.svcCtx)
	return commentLogic.GetComments(req)
}
