// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package upload

import (
	"context"
	uploadModel "task_Project/model/upload"
	"task_Project/task/internal/utils"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAttachmentCommentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取附件评论列表
func NewGetAttachmentCommentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAttachmentCommentsLogic {
	return &GetAttachmentCommentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAttachmentCommentsLogic) GetAttachmentComments(req *types.GetAttachmentCommentsRequest) (resp *types.BaseResponse, err error) {
	if req.FileID == "" {
		return utils.Response.ValidationError("文件ID不能为空"), nil
	}

	page := int64(req.Page)
	pageSize := int64(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	comments, total, err := l.svcCtx.AttachmentCommentModel.FindByFileID(l.ctx, req.FileID, page, pageSize)
	if err != nil {
		logx.Errorf("查询附件评论失败: %v", err)
		return utils.Response.InternalError("查询评论失败"), nil
	}

	// 转换为响应格式
	list := make([]types.AttachmentCommentInfo, 0, len(comments))
	for _, c := range comments {
		// 只返回顶级评论
		if c.ParentID != "" {
			continue
		}

		// 获取回复列表
		replies, _ := l.svcCtx.AttachmentCommentModel.FindReplies(l.ctx, c.CommentID)
		replyList := make([]types.AttachmentCommentInfo, 0, len(replies))
		for _, r := range replies {
			replyList = append(replyList, l.convertToCommentInfo(r))
		}

		info := l.convertToCommentInfo(c)
		info.Replies = replyList
		list = append(list, info)
	}

	return utils.Response.Success(map[string]interface{}{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}), nil
}

// convertToCommentInfo 转换评论信息
func (l *GetAttachmentCommentsLogic) convertToCommentInfo(c *uploadModel.Attachment_comment) types.AttachmentCommentInfo {
	var annotationDataReq types.AnnotationDataReq
	if c.AnnotationData != nil {
		annotationDataReq = types.AnnotationDataReq{
			X:      c.AnnotationData.X,
			Y:      c.AnnotationData.Y,
			Width:  c.AnnotationData.Width,
			Height: c.AnnotationData.Height,
			Color:  c.AnnotationData.Color,
			Text:   c.AnnotationData.Text,
			StartX: c.AnnotationData.StartX,
			StartY: c.AnnotationData.StartY,
			EndX:   c.AnnotationData.EndX,
			EndY:   c.AnnotationData.EndY,
		}
	}

	resolvedAt := ""
	if !c.ResolvedAt.IsZero() {
		resolvedAt = c.ResolvedAt.Format(time.RFC3339)
	}

	return types.AttachmentCommentInfo{
		ID:              c.ID.Hex(),
		CommentID:       c.CommentID,
		FileID:          c.FileID,
		TaskID:          c.TaskID,
		TaskNodeID:      c.TaskNodeID,
		UserID:          c.UserID,
		EmployeeID:      c.EmployeeID,
		EmployeeName:    c.EmployeeName,
		Content:         c.Content,
		AtEmployeeIDs:   c.AtEmployeeIDs,
		AtEmployeeNames: c.AtEmployeeNames,
		AnnotationType:  c.AnnotationType,
		AnnotationData:  annotationDataReq,
		PageNumber:      c.PageNumber,
		ParentID:        c.ParentID,
		ReplyToUserID:   c.ReplyToUserID,
		ReplyToName:     c.ReplyToName,
		IsResolved:      c.IsResolved,
		ResolvedBy:      c.ResolvedBy,
		ResolvedAt:      resolvedAt,
		CreateTime:      c.CreateAt.Format(time.RFC3339),
		UpdateTime:      c.UpdateAt.Format(time.RFC3339),
	}
}
