package upload

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"task_Project/task/internal/svc"
)

type AttachmentCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 附件评论逻辑
func NewAttachmentCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AttachmentCommentLogic {
	return &AttachmentCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
