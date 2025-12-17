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
	// todo: add your logic here and delete this line

	return
}
