// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"
	"time"

	taskmodel "task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

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
	page := int64(req.Page)
	pageSize := int64(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var comments []*taskmodel.Task_comment
	var total int64

	if req.TaskNodeID != "" {
		comments, total, err = l.svcCtx.TaskCommentModel.FindByTaskNodeID(l.ctx, req.TaskNodeID, page, pageSize)
	} else if req.TaskID != "" {
		comments, total, err = l.svcCtx.TaskCommentModel.FindByTaskID(l.ctx, req.TaskID, page, pageSize)
	} else {
		return utils.Response.ValidationError("任务ID或任务节点ID不能同时为空"), nil
	}

	if err != nil {
		logx.Errorf("查询评论失败: %v", err)
		return utils.Response.InternalError("查询评论失败"), nil
	}

	// 获取当前用户ID（用于判断是否已点赞）
	currentUserID, _ := utils.Common.GetCurrentUserID(l.ctx)

	// 转换为响应格式
	list := make([]types.TaskCommentInfo, 0, len(comments))
	for _, c := range comments {
		// 只返回顶级评论
		if c.ParentID != "" {
			continue
		}

		// 获取回复列表
		replies, _ := l.svcCtx.TaskCommentModel.FindReplies(l.ctx, c.CommentID)
		replyList := make([]types.TaskCommentInfo, 0, len(replies))
		for _, r := range replies {
			replyList = append(replyList, l.convertToCommentInfo(r, currentUserID))
		}

		info := l.convertToCommentInfo(c, currentUserID)
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
func (l *GetTaskCommentsLogic) convertToCommentInfo(c *taskmodel.Task_comment, currentUserID string) types.TaskCommentInfo {
	isLiked := false
	for _, uid := range c.LikedBy {
		if uid == currentUserID {
			isLiked = true
			break
		}
	}

	return types.TaskCommentInfo{
		ID:              c.ID.Hex(),
		CommentID:       c.CommentID,
		TaskID:          c.TaskID,
		TaskNodeID:      c.TaskNodeID,
		UserID:          c.UserID,
		EmployeeID:      c.EmployeeID,
		EmployeeName:    c.EmployeeName,
		Content:         c.Content,
		ContentHTML:     c.ContentHTML,
		AtEmployeeIDs:   c.AtEmployeeIDs,
		AtEmployeeNames: c.AtEmployeeNames,
		ParentID:        c.ParentID,
		ReplyToUserID:   c.ReplyToUserID,
		ReplyToName:     c.ReplyToName,
		AttachmentIDs:   c.AttachmentIDs,
		AttachmentURLs:  c.AttachmentURLs,
		LikeCount:       int64(c.LikeCount),
		IsLiked:         isLiked,
		CreateTime:      c.CreateAt.Format(time.RFC3339),
		UpdateTime:      c.UpdateAt.Format(time.RFC3339),
	}
}
