// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	taskmodel "task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateTaskProgressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 任务进度更新
func NewUpdateTaskProgressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTaskProgressLogic {
	return &UpdateTaskProgressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTaskProgressLogic) UpdateTaskProgress(req *types.UpdateTaskProgressRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.TaskNodeID == "" {
		return utils.Response.BusinessError("task_node_not_found"), nil
	}
	if req.Progress < 0 || req.Progress > 100 {
		return utils.Response.BusinessError("progress_range_error"), nil
	}

	// 2. 获取当前用户ID
	// 1. 从上下文获取当前员工ID
	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}

	// 3. 获取任务节点信息
	taskNode, err := l.svcCtx.TaskNodeModel.FindOneSafe(l.ctx, req.TaskNodeID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("task_not_found"), nil
		}
		return nil, err
	}
	// 4. 验证用户权限（只有执行人或负责人可以更新进度）
	exectorIds := strings.Split(taskNode.ExecutorId, ",")
	leaderIds := strings.Split(taskNode.LeaderId, ",")
	i := make(chan bool, 1)
	for _, leaderId := range leaderIds {
		if leaderId == employeeId {
			i <- true
		}
	}
	go func() {
		for _, exectorId := range exectorIds {
			if exectorId == employeeId {
				i <- true
				return
			}
		}
	}()
	if !<-i {
		return utils.Response.BusinessError("taskProgress"), nil
	}

	// 5. 更新任务节点进度
	err = l.svcCtx.TaskNodeModel.UpdateProgress(l.ctx, req.TaskNodeID, req.Progress)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务节点进度失败: %v", err)
		return utils.Response.InternalError("更新进度失败"), nil
	}

	// 6. 如果提供了实际工时，更新实际工时
	if req.ActualHours > 0 {
		err = l.svcCtx.TaskNodeModel.UpdateActualHours(l.ctx, req.TaskNodeID, req.ActualHours)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务节点实际工时失败: %v", err)
			// 不影响主流程，继续执行
		}
	}

	// 7. 注意：进度100%时不自动改为已完成，需要员工手动提交审批，审批通过后才改为已完成

	// 8. 创建任务日志
	logContent := fmt.Sprintf("更新任务节点进度: %d%%", req.Progress)
	if req.ProgressNote != "" {
		logContent += fmt.Sprintf("，备注: %s", req.ProgressNote)
	}
	if req.ActualHours > 0 {
		logContent += fmt.Sprintf("，实际工时: %d小时", req.ActualHours)
	}

	taskLog := &taskmodel.TaskLog{
		LogId:      utils.NewCommon().GenerateIDWithPrefix("task_log"),
		TaskId:     taskNode.TaskId,
		TaskNodeId: utils.Common.ToSqlNullString(req.TaskNodeID),
		EmployeeId: employeeId,
		LogType:    1, // 更新类型
		LogContent: logContent,
		Progress:   sql.NullInt64{Int64: int64(req.Progress), Valid: true},
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务日志失败: %v", err)
		// 不影响主流程，继续执行
	}

	// 9. 如果进度达到100%，发送完成通知给负责人（通过消息队列）
	if req.Progress >= 100 {
		// 获取任务信息和负责人信息
		taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, taskNode.TaskId)
		if err == nil {
			// 收集需要通知的人员（负责人和执行人）
			employeeIDSet := make(map[string]bool)
			if taskNode.LeaderId != "" {
				employeeIDSet[taskNode.LeaderId] = true
			}
			for _, executorId := range exectorIds {
				if executorId != "" {
					employeeIDSet[executorId] = true
				}
			}
			employeeIDs := make([]string, 0, len(employeeIDSet))
			for id := range employeeIDSet {
				employeeIDs = append(employeeIDs, id)
			}

			// 发布通知事件
			if l.svcCtx.NotificationMQService != nil && len(employeeIDs) > 0 {
				notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
					svc.TaskNodeCompleted,
					employeeIDs,
					req.TaskNodeID,
					svc.NotificationEventOptions{TaskID: taskNode.TaskId, NodeID: req.TaskNodeID},
				)
				notificationEvent.Title = "任务节点完成通知"
				notificationEvent.Content = fmt.Sprintf("任务节点 %s 已完成（任务：%s）", taskNode.NodeName, taskInfo.TaskTitle)
				notificationEvent.Priority = 2
				if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
					l.Logger.WithContext(l.ctx).Errorf("发布任务节点完成通知事件失败: %v", err)
				}
			}

			// 发布邮件事件
			if l.svcCtx.EmailMQService != nil {
				emailEvent := &svc.EmailEvent{
					EventType: svc.TaskNodeCompleted,
					TaskID:    taskNode.TaskId,
					NodeID:    req.TaskNodeID,
				}
				if err := l.svcCtx.EmailMQService.PublishEmailEvent(l.ctx, emailEvent); err != nil {
					l.Logger.WithContext(l.ctx).Errorf("发布任务节点完成邮件事件失败: %v", err)
				}
			}

			// 同时也通过 EmailService 直接发送（如果可用）
			leader, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, taskNode.LeaderId)
			if err == nil && leader.Email.Valid && leader.Email.String != "" {
				if l.svcCtx.EmailService != nil {
					completeTime := time.Now().Format("2006-01-02 15:04:05")
					if err := l.svcCtx.EmailService.SendTaskCompletedEmail(l.ctx, leader.Email.String, taskInfo.TaskTitle, taskNode.NodeName, completeTime); err != nil {
						l.Logger.WithContext(l.ctx).Errorf("发送任务完成邮件失败: %v", err)
					}
				}
			}
		}
	}

	// 10. 更新任务整体进度（根据所有节点进度计算平均值）
	err = l.updateTaskProgress(req.TaskNodeID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务整体进度失败: %v", err)
		// 不影响主流程，继续执行
	}

	return utils.Response.Success(map[string]interface{}{
		"taskNodeId": req.TaskNodeID,
		"progress":   req.Progress,
		"message":    "进度更新成功",
	}), nil
}

// updateTaskProgress 根据所有任务节点进度更新任务整体进度（与 updateChecklistLogic 中的实现保持一致）
func (l *UpdateTaskProgressLogic) updateTaskProgress(taskNodeId string) error {
	// 获取任务节点信息
	taskNode, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, taskNodeId)
	if err != nil {
		return err
	}

	// 获取该任务的所有节点
	nodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, taskNode.TaskId)
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return nil
	}

	// 计算平均进度和完成节点数（只统计状态为已完成（状态2）的节点）
	var totalProgress int64
	var completedCount int64
	allNodesCompleted := true
	for _, node := range nodes {
		totalProgress += node.Progress
		if node.NodeStatus == 2 { // 状态为已完成
			completedCount++
		} else {
			allNodesCompleted = false
		}
	}
	avgProgress := int(totalProgress / int64(len(nodes)))

	// 更新任务进度
	err = l.svcCtx.TaskModel.UpdateProgress(l.ctx, taskNode.TaskId, avgProgress)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务进度失败: %v", err)
	}

	// 只有当所有节点都完成（状态2）且平均进度达到100%时，才更新任务状态为已完成
	if allNodesCompleted && avgProgress == 100 {
		err = l.svcCtx.TaskModel.UpdateStatus(l.ctx, taskNode.TaskId, 2)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务状态失败: %v", err)
		}
		l.Logger.WithContext(l.ctx).Infof("任务 %s 所有节点已完成，任务状态更新为已完成", taskNode.TaskId)
	}

	// 更新任务节点统计
	err = l.svcCtx.TaskModel.UpdateNodeCount(l.ctx, taskNode.TaskId, int64(len(nodes)), completedCount)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务节点统计失败: %v", err)
	}

	return nil
}
