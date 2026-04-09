package tasknode

import (
	"context"
	"errors"
	"fmt"
	"time"

	"task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteTaskNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTaskNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTaskNodeLogic {
	return &DeleteTaskNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTaskNodeLogic) DeleteTaskNode(req *types.DeleteTaskNodeRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.TaskNodeID == "" {
		return utils.Response.BusinessError("task_node_id_required"), nil
	}

	// 2. 获取当前用户ID
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 3. 获取任务节点信息
	taskNode, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, req.TaskNodeID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("task_node_not_found"), nil
		}
		return nil, err
	}

	// 4. 验证用户权限（只有节点负责人可以删除）
	if taskNode.LeaderId != currentUserID {
		return utils.Response.BusinessError("task_node_delete_denied"), nil
	}

	// 5. 检查任务节点状态
	if taskNode.NodeStatus == 3 { // 已完成
		return utils.Response.BusinessError("task_node_completed_no_delete"), nil
	}

	// 7. 软删除任务节点
	err = l.svcCtx.TaskNodeModel.SoftDelete(l.ctx, req.TaskNodeID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("删除任务节点失败: %v", err)
		return nil, err
	}

	// 8. 创建任务日志
	taskLog := &task.TaskLog{
		LogId:      utils.Common.GenerateID(),
		TaskId:     taskNode.TaskId,
		LogType:    4, // 删除类型
		LogContent: fmt.Sprintf("任务节点 %s 已被删除", taskNode.NodeName),
		EmployeeId: currentUserID,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务日志失败: %v", err)
	}

	// 9. 通知相关人员（通过消息队列）
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, taskNode.TaskId)
	taskTitle := ""
	if err == nil {
		taskTitle = taskInfo.TaskTitle
	}

	// 获取操作者信息
	operatorName := "系统"
	operator, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, currentUserID)
	if err == nil {
		operatorName = operator.RealName
	}

	// 收集需要通知的员工
	employeeIDSet := make(map[string]bool)
	emails := []string{}

	// 通知节点负责人
	if taskNode.LeaderId != "" {
		employeeIDSet[taskNode.LeaderId] = true
		leader, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, taskNode.LeaderId)
		if err == nil && leader.Email.Valid && leader.Email.String != "" {
			emails = append(emails, leader.Email.String)
		}
	}

	// 通知执行人
	if taskNode.ExecutorId != "" && !employeeIDSet[taskNode.ExecutorId] {
		employeeIDSet[taskNode.ExecutorId] = true
		executor, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, taskNode.ExecutorId)
		if err == nil && executor.Email.Valid && executor.Email.String != "" {
			emails = append(emails, executor.Email.String)
		}
	}

	employeeIDs := make([]string, 0, len(employeeIDSet))
	for id := range employeeIDSet {
		employeeIDs = append(employeeIDs, id)
	}

	// 发布通知事件（使用工厂方法）
	if l.svcCtx.NotificationMQService != nil && len(employeeIDs) > 0 {
		notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
			svc.TaskNodeDeleted,
			employeeIDs,
			req.TaskNodeID,
			svc.NotificationEventOptions{TaskID: taskNode.TaskId, NodeID: req.TaskNodeID},
		)
		notificationEvent.Title = "任务节点删除通知"
		notificationEvent.Content = fmt.Sprintf("任务节点 %s（任务：%s）已被 %s 删除", taskNode.NodeName, taskTitle, operatorName)
		notificationEvent.Priority = 2
		if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
			l.Logger.WithContext(l.ctx).Errorf("发布任务节点删除通知事件失败: %v", err)
		}
	}

	// 发送邮件通知（使用新的统一邮件服务）
	if l.svcCtx.UnifiedEmailService != nil && len(emails) > 0 {
		taskData := map[string]interface{}{
			"taskId":       taskNode.TaskId,
			"taskTitle":    taskTitle,
			"nodeName":     taskNode.NodeName,
			"operatorName": operatorName,
			"deleteTime":   time.Now().Format("2006-01-02 15:04:05"),
			"message":      fmt.Sprintf("任务节点 %s（任务：%s）已被 %s 删除", taskNode.NodeName, taskTitle, operatorName),
		}

		if err := l.svcCtx.UnifiedEmailService.SendTaskNotification(l.ctx, "task_node_deleted", emails, taskData); err != nil {
			l.Logger.WithContext(l.ctx).Errorf("发送任务节点删除邮件失败: %v", err)
		}
	}

	return utils.Response.Success(map[string]interface{}{
		"taskNodeId": req.TaskNodeID,
		"message":    "任务节点删除成功",
	}), nil
}
