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

type CompleteTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCompleteTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompleteTaskLogic {
	return &CompleteTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CompleteTaskLogic) CompleteTask(req *types.CompleteTaskRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.TaskID == "" {
		return utils.Response.BusinessError("任务ID不能为空"), nil
	}

	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}

	// 3. 获取任务信息
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, req.TaskID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("任务不存在"), nil
		}
		return nil, err
	}

	// 4. 验证用户权限（只有任务创建者可以标记任务完成）
	if taskInfo.TaskCreator != employeeId {
		return utils.Response.BusinessError("无权限完成此任务"), nil
	}

	// 5. 检查任务状态
	if taskInfo.TaskStatus == 3 { // 已完成
		return utils.Response.BusinessError("任务已经完成"), nil
	}

	// 6. 检查所有任务节点是否都已完成
	taskNodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, req.TaskID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询任务节点失败: %v", err)
		return nil, err
	}

	allNodesCompleted := true
	for _, node := range taskNodes {
		if node.NodeStatus != 3 { // 未完成
			allNodesCompleted = false
			break
		}
	}

	if !allNodesCompleted {
		return utils.Response.BusinessError("所有任务节点完成后才能标记任务完成"), nil
	}

	// 7. 更新任务状态为已完成
	updatedTask := *taskInfo
	updatedTask.TaskStatus = 3 // 已完成
	updatedTask.UpdateTime = time.Now()

	err = l.svcCtx.TaskModel.Update(l.ctx, &updatedTask)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务状态失败: %v", err)
		return nil, err
	}

	// 8. 创建任务日志
	taskLog := &task.TaskLog{
		LogId:      utils.Common.GenerateID(),
		TaskId:     req.TaskID,
		LogType:    3, // 完成类型
		LogContent: fmt.Sprintf("任务 %s 已完成", taskInfo.TaskTitle),
		EmployeeId: employeeId,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务日志失败: %v", err)
	}

	// 9. 通知相关人员（通过消息队列，消费者会查询并发送）
	// 收集所有相关人员ID
	employeeIDSet := make(map[string]bool)
	if taskInfo.TaskCreator != "" {
		employeeIDSet[taskInfo.TaskCreator] = true
	}
	for _, node := range taskNodes {
		if node.LeaderId != "" {
			employeeIDSet[node.LeaderId] = true
		}
		if node.ExecutorId != "" {
			employeeIDSet[node.ExecutorId] = true
		}
	}
	employeeIDs := make([]string, 0, len(employeeIDSet))
	for id := range employeeIDSet {
		employeeIDs = append(employeeIDs, id)
	}

	// 发布通知事件
	if l.svcCtx.NotificationMQService != nil && len(employeeIDs) > 0 {
		notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
			svc.TaskCompleted,
			employeeIDs,
			req.TaskID,
			svc.NotificationEventOptions{TaskID: req.TaskID},
		)
		notificationEvent.Title = "任务完成通知"
		notificationEvent.Content = fmt.Sprintf("任务 %s 已完成", taskInfo.TaskTitle)
		notificationEvent.Priority = 2
		if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
			l.Logger.WithContext(l.ctx).Errorf("发布任务完成通知事件失败: %v", err)
		}
	}

	// 发布邮件事件
	if l.svcCtx.EmailMQService != nil {
		emailEvent := &svc.EmailEvent{
			EventType: svc.TaskCompleted,
			TaskID:    req.TaskID,
		}
		if err := l.svcCtx.EmailMQService.PublishEmailEvent(l.ctx, emailEvent); err != nil {
			l.Logger.WithContext(l.ctx).Errorf("发布任务完成邮件事件失败: %v", err)
		}
	}

	return utils.Response.Success(map[string]interface{}{
		"taskId":  req.TaskID,
		"message": "任务完成成功",
	}), nil
}
