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

type ConfirmHandoverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmHandoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmHandoverLogic {
	return &ConfirmHandoverLogic{
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

// ConfirmHandover 接收人确认接收交接（第一步 - 同意）
func (l *ConfirmHandoverLogic) ConfirmHandover(req *types.ConfirmHandoverRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.HandoverID == "" {
		return utils.Response.ValidationError("交接ID不能为空"), nil
	}

	// 2. 获取当前用户ID并查找员工信息
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 获取当前员工信息
	currentEmployee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, currentUserID)
	if err != nil {
		l.Logger.Errorf("获取当前员工信息失败: %v", err)
		return utils.Response.ValidationError("用户未绑定员工信息"), nil
	}
	currentEmployeeID := currentEmployee.Id
	approverName := currentEmployee.RealName

	// 3. 获取交接记录
	handover, err := l.svcCtx.TaskHandoverModel.FindOne(l.ctx, req.HandoverID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.ValidationError("交接记录不存在"), nil
		}
		return nil, err
	}

	// 4. 验证交接状态（只有待接收人确认的才能确认）
	if handover.HandoverStatus != 0 {
		return utils.Response.ValidationError("只有待接收人确认的交接才能进行确认操作"), nil
	}

	// 5. 验证是否是交接接收人
	if handover.ToEmployeeId != currentEmployeeID {
		return utils.Response.ValidationError("只有交接接收人才能确认"), nil
	}

	// 6. 获取任务信息
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, handover.TaskId)
	if err != nil {
		l.Logger.Errorf("获取任务信息失败: %v", err)
	}
	taskTitle := ""
	if taskInfo != nil {
		taskTitle = taskInfo.TaskTitle
	}

	// 7. 更新交接状态为待上级审批
	handover.HandoverStatus = 1 // 待上级审批
	handover.UpdateTime = time.Now()
	err = l.svcCtx.TaskHandoverModel.Update(l.ctx, handover)
	if err != nil {
		return nil, err
	}

	// 8. 插入审批记录到数据库
	approvalRecord := &taskModel.HandoverApproval{
		ApprovalId:   utils.Common.GenerateIDWithPrefix("approval"),
		HandoverId:   req.HandoverID,
		ApprovalStep: 1, // 第一步：接收人确认
		ApproverId:   currentEmployeeID,
		ApproverName: approverName,
		ApprovalType: 1, // 同意
		Comment:      sql.NullString{String: "接收人同意接收任务", Valid: true},
		CreateTime:   time.Now(),
	}
	_, err = l.svcCtx.HandoverApprovalModel.Insert(l.ctx, approvalRecord)
	if err != nil {
		l.Logger.Errorf("插入审批记录失败: %v", err)
	}

	// 9. 创建任务日志
	taskLog := &taskModel.TaskLog{
		LogId:      utils.Common.GenerateID(),
		TaskId:     handover.TaskId,
		LogType:    7, // 交接确认
		LogContent: fmt.Sprintf("接收人确认交接: %s -> %s 已同意接收，等待上级审批", handover.FromEmployeeId, handover.ToEmployeeId),
		EmployeeId: currentEmployeeID,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.Errorf("创建任务日志失败: %v", err)
	}

	// 10. 发送通知给发起人（接收人已同意）
	l.Logger.Infof("准备发送通知, NotificationMQService=%v", l.svcCtx.NotificationMQService != nil)
	if l.svcCtx.NotificationMQService != nil {
		notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
			svc.HandoverNotification,
			[]string{handover.FromEmployeeId},
			req.HandoverID,
			svc.NotificationEventOptions{TaskID: handover.TaskId},
		)
		notificationEvent.Title = "交接接收确认"
		notificationEvent.Content = fmt.Sprintf("接收人已同意接收任务「%s」的交接，等待上级审批", taskTitle)
		notificationEvent.Priority = 2
		l.Logger.Infof("发布通知事件: toEmployees=%v, title=%s", []string{handover.FromEmployeeId}, notificationEvent.Title)
		if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
			l.Logger.Errorf("发布通知事件失败: %v", err)
		} else {
			l.Logger.Infof("通知事件发布成功")
		}
	} else {
		l.Logger.Errorf("NotificationMQService 为空，无法发送通知")
	}

	// 11. 发送通知给审批人（如果有）
	if handover.ApproverId.Valid && handover.ApproverId.String != "" {
		if l.svcCtx.NotificationMQService != nil {
			notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
				svc.HandoverNotification,
				[]string{handover.ApproverId.String},
				req.HandoverID,
				svc.NotificationEventOptions{TaskID: handover.TaskId},
			)
			notificationEvent.Title = "交接审批请求"
			notificationEvent.Content = fmt.Sprintf("有一个任务「%s」的交接申请需要您审批", taskTitle)
			notificationEvent.Priority = 2
			if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
				l.Logger.Errorf("发布通知事件失败: %v", err)
			}
		}

		// 发布邮件事件给审批人
		if l.svcCtx.EmailMQService != nil {
			emailEvent := &svc.EmailEvent{
				EventType:   svc.HandoverNotification,
				EmployeeIDs: []string{handover.ApproverId.String},
				RelatedID:   req.HandoverID,
			}
			if err := l.svcCtx.EmailMQService.PublishEmailEvent(l.ctx, emailEvent); err != nil {
				l.Logger.Errorf("发布邮件事件失败: %v", err)
			}
		}
	}

	return utils.Response.Success(map[string]interface{}{
		"handoverId": req.HandoverID,
		"status":     1,
		"statusText": "待上级审批",
		"approverId": handover.ApproverId.String,
		"message":    "已确认接收，等待上级审批",
	}), nil
}
