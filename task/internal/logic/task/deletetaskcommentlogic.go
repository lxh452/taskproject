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
	if req.CommentID == "" {
		return utils.Response.ValidationError("评论ID不能为空"), nil
	}

	// 验证是否是评论作者
	comment, err := l.svcCtx.TaskCommentModel.FindByCommentID(l.ctx, req.CommentID)
	if err != nil {
		return utils.Response.NotFoundError("评论不存在"), nil
	}

	currentUserID, _ := utils.Common.GetCurrentUserID(l.ctx)
	if comment.UserID != currentUserID {
		return utils.Response.ForbiddenError("无权删除此评论"), nil
	}

	err = l.svcCtx.TaskCommentModel.SoftDelete(l.ctx, comment.ID.Hex())
	if err != nil {
		logx.Errorf("删除评论失败: %v", err)
		return utils.Response.InternalError("删除评论失败"), nil
	}

	return utils.Response.Success(nil), nil
}
