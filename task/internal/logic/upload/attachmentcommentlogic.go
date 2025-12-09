package upload

import (
	"context"
	"time"

	uploadModel "task_Project/model/upload"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
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

// CreateComment 创建附件评论
func (l *AttachmentCommentLogic) CreateComment(req *types.CreateAttachmentCommentRequest) (resp *types.BaseResponse, err error) {
	if req.FileID == "" {
		return utils.Response.ValidationError("文件ID不能为空"), nil
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

	// 如果从上下文获取不到employeeID或realName，尝试从数据库查询
	if employeeID == "" || realName == "" {
		logx.Infof("上下文中缺少员工信息, userID=%s, employeeID=%s, realName=%s, 尝试从数据库查询", userID, employeeID, realName)
		employee, dbErr := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, userID)
		if dbErr != nil {
			logx.Errorf("通过userID查询员工失败: userID=%s, error=%v", userID, dbErr)
		} else if employee != nil {
			logx.Infof("查询到员工信息: Id=%s, RealName=%s", employee.Id, employee.RealName)
			if employeeID == "" {
				employeeID = employee.Id
			}
			if realName == "" {
				realName = employee.RealName
			}
		}
	}

	// 如果仍然没有realName，尝试从User表获取
	if realName == "" {
		username, _ := utils.Common.GetCurrentUsername(l.ctx)
		if username != "" {
			realName = username
		}
	}

	// 最终兜底：使用 userID 的一部分
	if realName == "" {
		if len(userID) > 8 {
			realName = "用户" + userID[:8]
		} else {
			realName = "用户" + userID
		}
		logx.Infof("无法获取用户真实姓名，使用兜底值: %s", realName)
	}

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
		parentComment, err := l.svcCtx.AttachmentCommentModel.FindByCommentID(l.ctx, req.ParentID)
		if err == nil && parentComment != nil {
			replyToUserID = parentComment.UserID
			replyToName = parentComment.EmployeeName
		}
	}

	commentID := utils.Common.GenId("attcomment")

	// 转换标注数据
	var annotationData *uploadModel.AnnotationData
	if req.AnnotationData != nil {
		annotationData = &uploadModel.AnnotationData{
			X:      req.AnnotationData.X,
			Y:      req.AnnotationData.Y,
			Width:  req.AnnotationData.Width,
			Height: req.AnnotationData.Height,
			Color:  req.AnnotationData.Color,
			Text:   req.AnnotationData.Text,
			StartX: req.AnnotationData.StartX,
			StartY: req.AnnotationData.StartY,
			EndX:   req.AnnotationData.EndX,
			EndY:   req.AnnotationData.EndY,
		}
	}

	// 创建评论
	comment := &uploadModel.Attachment_comment{
		CommentID:       commentID,
		FileID:          req.FileID,
		TaskID:          req.TaskID,
		TaskNodeID:      req.TaskNodeID,
		UserID:          userID,
		EmployeeID:      employeeID,
		EmployeeName:    realName,
		Content:         req.Content,
		AtEmployeeIDs:   req.AtEmployeeIDs,
		AtEmployeeNames: atEmployeeNames,
		AnnotationType:  req.AnnotationType,
		AnnotationData:  annotationData,
		PageNumber:      req.PageNumber,
		ParentID:        req.ParentID,
		ReplyToUserID:   replyToUserID,
		ReplyToName:     replyToName,
		IsResolved:      false,
		IsDeleted:       false,
	}

	err = l.svcCtx.AttachmentCommentModel.Insert(l.ctx, comment)
	if err != nil {
		logx.Errorf("创建附件评论失败: %v", err)
		return utils.Response.InternalError("创建评论失败"), nil
	}

	// 发送@通知
	if len(req.AtEmployeeIDs) > 0 && l.svcCtx.NotificationMQService != nil {
		go func() {
			event := &svc.NotificationEvent{
				EventType:   "comment.mention",
				EmployeeIDs: req.AtEmployeeIDs,
				Title:       "附件评论中被@提及",
				Content:     realName + "在附件评论中@了你: " + req.Content,
				Type:        3, // 类型: 系统通知
				Category:    "comment",
				Priority:    2, // 优先级: 普通
				RelatedID:   commentID,
				RelatedType: "attachment_comment",
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

// GetComments 获取附件评论列表
func (l *AttachmentCommentLogic) GetComments(req *types.GetAttachmentCommentsRequest) (resp *types.BaseResponse, err error) {
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

// ResolveComment 标记评论已解决
func (l *AttachmentCommentLogic) ResolveComment(req *types.ResolveAttachmentCommentRequest) (resp *types.BaseResponse, err error) {
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
}

// DeleteComment 删除评论（软删除）
func (l *AttachmentCommentLogic) DeleteComment(req *types.DeleteAttachmentCommentRequest) (resp *types.BaseResponse, err error) {
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

// convertToCommentInfo 转换评论信息
func (l *AttachmentCommentLogic) convertToCommentInfo(c *uploadModel.Attachment_comment) types.AttachmentCommentInfo {
	var annotationData *types.AnnotationDataReq
	if c.AnnotationData != nil {
		annotationData = &types.AnnotationDataReq{
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
		AnnotationData:  annotationData,
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
