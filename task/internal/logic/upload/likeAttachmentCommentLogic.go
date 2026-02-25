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

type LikeAttachmentCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 点赞/取消点赞附件评论
func NewLikeAttachmentCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeAttachmentCommentLogic {
	return &LikeAttachmentCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LikeAttachmentCommentLogic) LikeAttachmentComment(req *types.LikeCommentRequest) (resp *types.BaseResponse, err error) {
	if req.CommentID == "" {
		return utils.Response.ValidationError("评论ID不能为空"), nil
	}

	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	isLike := req.IsLike == 1
	likeCount, err := l.svcCtx.AttachmentCommentModel.ToggleLike(l.ctx, req.CommentID, userID, isLike)
	if err != nil {
		logx.Errorf("附件评论点赞失败: %v", err)
		return utils.Response.InternalError("操作失败"), nil
	}

	return utils.Response.Success(map[string]interface{}{
		"likeCount": likeCount,
	}), nil
}
