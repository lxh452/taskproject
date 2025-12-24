// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package upload

import (
	"context"
	"task_Project/task/internal/utils"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAttachmentCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除附件评论
func NewDeleteAttachmentCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAttachmentCommentLogic {
	return &DeleteAttachmentCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAttachmentCommentLogic) DeleteAttachmentComment(req *types.DeleteAttachmentCommentRequest) (resp *types.BaseResponse, err error) {
	if req.CommentID == "" {
		return utils.Response.ValidationError("评论ID不能为空"), nil
	}

	// 验证是否是评论作者
	comment, err := l.svcCtx.AttachmentCommentModel.FindByCommentID(l.ctx, req.CommentID)
	if err != nil {
		return utils.Response.NotFoundError("评论不存在"), nil
	}

	currentUserID, _ := utils.Common.GetCurrentUserID(l.ctx)
	if comment.UserID != currentUserID {
		return utils.Response.ForbiddenError("无权删除此评论"), nil
	}

	err = l.svcCtx.AttachmentCommentModel.SoftDelete(l.ctx, comment.ID.Hex())
	if err != nil {
		logx.Errorf("删除评论失败: %v", err)
		return utils.Response.InternalError("删除评论失败"), nil
	}

	return utils.Response.Success(nil), nil
}
