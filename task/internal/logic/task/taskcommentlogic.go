package task

import (
	"context"
	"time"

	"task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type TaskCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 任务评论逻辑
func NewTaskCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TaskCommentLogic {
	return &TaskCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateComment 创建任务评论
func (l *TaskCommentLogic) CreateComment(req *types.CreateTaskCommentRequest) (resp *types.BaseResponse, err error) {
	if req.TaskID == "" {
		return utils.Response.ValidationError("任务ID不能为空"), nil
	}
	if req.Content == "" {
		return utils.Response.ValidationError("评论内容不能为空"), nil
	}

	// 获取当前用户信息
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}
	employeeID, _ := utils.Common.GetCurrentEmployeeID(l.ctx)
	realName, _ := utils.Common.GetCurrentRealName(l.ctx)

	// 获取@的员工姓名列表
	var atEmployeeNames []string
	if len(req.AtEmployeeIDs) > 0 {
		for _, empID := range req.AtEmployeeIDs {
			emp, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, empID)
			if err == nil && emp != nil {
				atEmployeeNames = append(atEmployeeNames, emp.RealName)
			}
		}
	}

	// 获取回复目标信息
	var replyToUserID, replyToName string
	if req.ParentID != "" {
		parentComment, err := l.svcCtx.TaskCommentModel.FindByCommentID(l.ctx, req.ParentID)
		if err == nil && parentComment != nil {
			replyToUserID = parentComment.UserID
			replyToName = parentComment.EmployeeName
		}
	}

	// 获取附件URL列表
	var attachmentURLs []string
	if len(req.AttachmentIDs) > 0 {
		for _, fileID := range req.AttachmentIDs {
			file, err := l.svcCtx.UploadFileModel.FindByFileID(l.ctx, fileID)
			if err == nil && file != nil {
				attachmentURLs = append(attachmentURLs, file.FileURL)
			}
		}
	}

	commentID := utils.Common.GenId("comment")

	// 创建评论
	comment := &task.Task_comment{
		CommentID:       commentID,
		TaskID:          req.TaskID,
		TaskNodeID:      req.TaskNodeID,
		UserID:          userID,
		EmployeeID:      employeeID,
		EmployeeName:    realName,
		Content:         req.Content,
		ContentHTML:     req.ContentHTML,
		AtEmployeeIDs:   req.AtEmployeeIDs,
		AtEmployeeNames: atEmployeeNames,
		ParentID:        req.ParentID,
		ReplyToUserID:   replyToUserID,
		ReplyToName:     replyToName,
		AttachmentIDs:   req.AttachmentIDs,
		AttachmentURLs:  attachmentURLs,
		LikeCount:       0,
		LikedBy:         []string{},
		IsDeleted:       false,
	}

	err = l.svcCtx.TaskCommentModel.Insert(l.ctx, comment)
	if err != nil {
		logx.Errorf("创建评论失败: %v", err)
		return utils.Response.InternalError("创建评论失败"), nil
	}

	// 发送@通知
	if len(req.AtEmployeeIDs) > 0 && l.svcCtx.NotificationMQService != nil {
		go func() {
			event := &svc.NotificationEvent{
				EventType:   "comment.mention",
				EmployeeIDs: req.AtEmployeeIDs,
				Title:       "评论中被@提及",
				Content:     realName + "在任务评论中@了你: " + req.Content,
				Type:        3, // 类型: 系统通知
				Category:    "comment",
				Priority:    2, // 优先级: 普通
				RelatedID:   commentID,
				RelatedType: "task_comment",
				TaskID:      req.TaskID,
				NodeID:      req.TaskNodeID,
			}
			if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, event); err != nil {
				logx.Errorf("发送@通知失败: %v", err)
			}
		}()
	}

	return utils.Response.Success(map[string]interface{}{
		"commentId": commentID,
	}), nil
}

// GetComments 获取任务评论列表
func (l *TaskCommentLogic) GetComments(req *types.GetTaskCommentsRequest) (resp *types.BaseResponse, err error) {
	page := int64(req.Page)
	pageSize := int64(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var comments []*task.Task_comment
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

// LikeComment 点赞/取消点赞评论
func (l *TaskCommentLogic) LikeComment(req *types.LikeCommentRequest) (resp *types.BaseResponse, err error) {
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

// DeleteComment 删除评论（软删除）
func (l *TaskCommentLogic) DeleteComment(req *types.DeleteTaskCommentRequest) (resp *types.BaseResponse, err error) {
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

// convertToCommentInfo 转换评论信息
func (l *TaskCommentLogic) convertToCommentInfo(c *task.Task_comment, currentUserID string) types.TaskCommentInfo {
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
