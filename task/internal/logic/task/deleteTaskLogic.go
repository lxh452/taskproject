package task

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

type DeleteTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTaskLogic {
	return &DeleteTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTaskLogic) DeleteTask(req *types.DeleteTaskRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.TaskID == "" {
		return utils.Response.BusinessError("task_id_required"), nil
	}

	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}

	// 3. 获取任务信息
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, req.TaskID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("task_not_found"), nil
		}
		return nil, err
	}

	// 4. 验证用户权限（只有任务创建者可以删除）
	if taskInfo.TaskCreator != employeeId {
		return utils.Response.BusinessError("task_delete_denied"), nil
	}

	// 5. 检查任务状态
	if taskInfo.TaskStatus == 3 { // 已完成
		return utils.Response.BusinessError("task_completed_no_delete"), nil
	}

	// 6. 检查是否有进行中的任务节点
	taskNodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, req.TaskID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询任务节点失败: %v", err)
		return nil, err
	}

	for _, node := range taskNodes {
		if node.NodeStatus == 1 || node.NodeStatus == 2 { // 进行中或已完成
			return utils.Response.BusinessError("task_has_active_nodes"), nil
		}
	}

	// 7. 软删除任务
	err = l.svcCtx.TaskModel.SoftDelete(l.ctx, req.TaskID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("删除任务失败: %v", err)
		return nil, err
	}

	// 8. 软删除相关任务节点
	for _, node := range taskNodes {
		err = l.svcCtx.TaskNodeModel.SoftDelete(l.ctx, node.TaskNodeId)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("删除任务节点失败: %v", err)
		}
	}

	// 9. 创建任务日志
	taskLog := &task.TaskLog{
		LogId:      utils.Common.GenerateID(),
		TaskId:     req.TaskID,
		LogType:    4, // 删除类型
		LogContent: fmt.Sprintf("任务 %s 已被删除", taskInfo.TaskTitle),
		EmployeeId: employeeId,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务日志失败: %v", err)
	}

	// 10. 通知相关人员（通过消息队列）
	// 收集需要通知的员工
	employeeIDSet := make(map[string]bool)
	emails := []string{}

	// 通知任务创建者
	if taskInfo.TaskCreator != "" {
		employeeIDSet[taskInfo.TaskCreator] = true
		creator, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, taskInfo.TaskCreator)
		if err == nil && creator.Email.Valid && creator.Email.String != "" {
			emails = append(emails, creator.Email.String)
		}
	}

	// 通知所有节点负责人和执行人
	for _, node := range taskNodes {
		if node.LeaderId != "" && !employeeIDSet[node.LeaderId] {
			employeeIDSet[node.LeaderId] = true
			leader, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, node.LeaderId)
			if err == nil && leader.Email.Valid && leader.Email.String != "" {
				emails = append(emails, leader.Email.String)
			}
		}
		if node.ExecutorId != "" && !employeeIDSet[node.ExecutorId] {
			employeeIDSet[node.ExecutorId] = true
			executor, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, node.ExecutorId)
			if err == nil && executor.Email.Valid && executor.Email.String != "" {
				emails = append(emails, executor.Email.String)
			}
		}
	}

	employeeIDs := make([]string, 0, len(employeeIDSet))
	for id := range employeeIDSet {
		employeeIDs = append(employeeIDs, id)
	}

	// 获取操作者信息
	operatorName := "系统"
	operator, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, employeeId)
	if err == nil {
		operatorName = operator.RealName
	}

	// 发布通知事件（使用工厂方法）
	if l.svcCtx.NotificationMQService != nil && len(employeeIDs) > 0 {
		notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
			svc.TaskDeleted,
			employeeIDs,
			req.TaskID,
			svc.NotificationEventOptions{TaskID: req.TaskID},
		)
		notificationEvent.Title = "任务删除通知"
		notificationEvent.Content = fmt.Sprintf("任务 %s 已被 %s 删除", taskInfo.TaskTitle, operatorName)
		notificationEvent.Priority = 2
		if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
			l.Logger.WithContext(l.ctx).Errorf("发布任务删除通知事件失败: %v", err)
		}
	}

	// 发布邮件事件（使用模板）
	if l.svcCtx.EmailMQService != nil && len(emails) > 0 {
		// 使用模板渲染邮件内容
		body := ""
		if l.svcCtx.EmailTemplateService != nil {
			data := svc.TaskDeletedData{
				TaskTitle:    taskInfo.TaskTitle,
				TaskID:       req.TaskID,
				OperatorName: operatorName,
				DeleteTime:   time.Now().Format("2006-01-02 15:04:05"),
				Year:         time.Now().Year(),
			}
			renderedBody, err := l.svcCtx.EmailTemplateService.RenderTemplate("task_deleted", data)
			if err == nil {
				body = renderedBody
			} else {
				l.Logger.WithContext(l.ctx).Errorf("渲染任务删除邮件模板失败: %v", err)
			}
		}

		emailEvent := &svc.EmailEvent{
			EventType: svc.TaskDeleted,
			To:        emails,
			Subject:   "任务删除通知",
			Body:      body,
			IsHTML:    true,
			TaskID:    req.TaskID,
		}
		if err := l.svcCtx.EmailMQService.PublishEmailEvent(l.ctx, emailEvent); err != nil {
			l.Logger.WithContext(l.ctx).Errorf("发布任务删除邮件事件失败: %v", err)
		}
	}

	return utils.Response.Success(map[string]interface{}{
		"taskId":  req.TaskID,
		"message": "任务删除成功",
	}), nil
}
