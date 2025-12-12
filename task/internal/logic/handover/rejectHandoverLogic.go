package handover

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	taskModel "task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RejectHandoverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRejectHandoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectHandoverLogic {
	return &RejectHandoverLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 交接状态定义:
// 0 = 待接收人确认
// 1 = 待上级审批
// 2 = 已通过
// 3 = 已拒绝
// 4 = 已完成

// RejectHandover 接收人拒绝交接（第一步 - 拒绝）
func (l *RejectHandoverLogic) RejectHandover(req *types.ApproveHandoverRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.HandoverID == "" {
		return utils.Response.BusinessError("交接ID不能为空"), nil
	}

	// 2. 获取交接记录
	handover, err := l.svcCtx.TaskHandoverModel.FindOne(l.ctx, req.HandoverID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("交接记录不存在"), nil
		}
		return nil, err
	}

	// 3. 验证交接状态（只有待接收人确认的才能拒绝）
	if handover.HandoverStatus != 0 {
		return utils.Response.BusinessError("只有待接收人确认的交接才能进行拒绝操作"), nil
	}

	// 4. 获取当前用户ID（接收人）
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 5. 验证是否是交接接收人
	if handover.ToEmployeeId != currentUserID {
		return utils.Response.BusinessError("只有交接接收人才能拒绝"), nil
	}

	// 6. 获取当前用户信息
	currentEmployee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, currentUserID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("获取当前用户信息失败: %v", err)
	}
	approverName := ""
	if currentEmployee != nil {
		approverName = currentEmployee.RealName
	}

	// 7. 获取任务信息
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, handover.TaskId)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("获取任务信息失败: %v", err)
	}
	taskTitle := ""
	if taskInfo != nil {
		taskTitle = taskInfo.TaskTitle
	}

	// 8. 更新交接状态为已拒绝
	handover.HandoverStatus = 3 // 已拒绝
	handover.UpdateTime = time.Now()
	err = l.svcCtx.TaskHandoverModel.Update(l.ctx, handover)
	if err != nil {
		return nil, err
	}

	// 9. 插入审批记录到数据库
	comment := req.Comment
	if comment == "" {
		comment = "接收人拒绝接收任务"
	}
	approvalRecord := &taskModel.HandoverApproval{
		ApprovalId:   utils.Common.GenerateIDWithPrefix("approval"),
		HandoverId:   req.HandoverID,
		ApprovalStep: 1, // 第一步：接收人确认
		ApproverId:   currentUserID,
		ApproverName: approverName,
		ApprovalType: 2, // 拒绝
		Comment:      sql.NullString{String: comment, Valid: true},
		CreateTime:   time.Now(),
	}
	_, err = l.svcCtx.HandoverApprovalModel.Insert(l.ctx, approvalRecord)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("插入审批记录失败: %v", err)
	}

	// 10. 创建任务日志
	taskLog := &taskModel.TaskLog{
		LogId:      utils.Common.GenerateID(),
		TaskId:     handover.TaskId,
		LogType:    7, // 交接确认
		LogContent: fmt.Sprintf("接收人拒绝交接: %s -> %s, 原因: %s", handover.FromEmployeeId, handover.ToEmployeeId, comment),
		EmployeeId: currentUserID,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务日志失败: %v", err)
	}

	// 11. 发送通知给发起人（接收人已拒绝）
	if l.svcCtx.NotificationMQService != nil {
		notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
			svc.HandoverNotification,
			[]string{handover.FromEmployeeId},
			req.HandoverID,
			svc.NotificationEventOptions{TaskID: handover.TaskId},
		)
		notificationEvent.Title = "交接被拒绝"
		notificationEvent.Content = fmt.Sprintf("接收人拒绝接收任务「%s」的交接，原因：%s", taskTitle, comment)
		notificationEvent.Priority = 2
		if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
			l.Logger.WithContext(l.ctx).Errorf("发布通知事件失败: %v", err)
		}
	}

	// 发布邮件事件给发起人
	if l.svcCtx.EmailMQService != nil {
		emailEvent := &svc.EmailEvent{
			EventType:   svc.HandoverNotification,
			EmployeeIDs: []string{handover.FromEmployeeId},
			RelatedID:   req.HandoverID,
		}
		if err := l.svcCtx.EmailMQService.PublishEmailEvent(l.ctx, emailEvent); err != nil {
			l.Logger.WithContext(l.ctx).Errorf("发布邮件事件失败: %v", err)
		}
	}

	return utils.Response.Success(map[string]interface{}{
		"handoverId": req.HandoverID,
		"status":     3,
		"statusText": "已拒绝",
		"message":    "已拒绝接收此交接",
	}), nil
}
