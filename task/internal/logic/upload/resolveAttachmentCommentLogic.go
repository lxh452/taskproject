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

type ResolveAttachmentCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 解决附件评论
func NewResolveAttachmentCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResolveAttachmentCommentLogic {
	return &ResolveAttachmentCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResolveAttachmentCommentLogic) ResolveAttachmentComment(req *types.ResolveAttachmentCommentRequest) (resp *types.BaseResponse, err error) {
	if req.CommentID == "" {
		return utils.Response.ValidationError("评论ID不能为空"), nil
	}

	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	err = l.svcCtx.AttachmentCommentModel.MarkResolved(l.ctx, req.CommentID, userID)
	if err != nil {
		logx.Errorf("标记评论已解决失败: %v", err)
		return utils.Response.InternalError("操作失败"), nil
	}

	return utils.Response.Success(nil), nil

	return
}
