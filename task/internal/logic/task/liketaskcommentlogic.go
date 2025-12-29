// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeTaskCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 点赞/取消点赞任务评论
func NewLikeTaskCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeTaskCommentLogic {
	return &LikeTaskCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LikeTaskCommentLogic) LikeTaskComment(req *types.LikeCommentRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
