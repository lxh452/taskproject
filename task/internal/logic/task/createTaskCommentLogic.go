// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"

	taskmodel "task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTaskCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建任务评论
func NewCreateTaskCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTaskCommentLogic {
	return &CreateTaskCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTaskCommentLogic) CreateTaskComment(req *types.CreateTaskCommentRequest) (resp *types.BaseResponse, err error) {
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
	comment := &taskmodel.Task_comment{
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
