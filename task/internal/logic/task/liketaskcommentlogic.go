// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

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
	if req.CommentID == "" {
		return utils.Response.ValidationError("评论ID不能为空"), nil
	}

	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	if req.IsLike == 1 {
		err = l.svcCtx.TaskCommentModel.AddLike(l.ctx, req.CommentID, userID)
	} else {
		err = l.svcCtx.TaskCommentModel.RemoveLike(l.ctx, req.CommentID, userID)
	}

	if err != nil {
		logx.Errorf("点赞操作失败: %v", err)
		return utils.Response.InternalError("操作失败"), nil
	}

	return utils.Response.Success(nil), nil
}
