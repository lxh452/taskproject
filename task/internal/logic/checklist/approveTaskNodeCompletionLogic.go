package checklist

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ApproveTaskNodeCompletionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApproveTaskNodeCompletionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveTaskNodeCompletionLogic {
	return &ApproveTaskNodeCompletionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApproveTaskNodeCompletionLogic) ApproveTaskNodeCompletion(req *types.ApproveTaskNodeCompletionRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.ApprovalID == "" {
		return utils.Response.BusinessError("approval_id_required"), nil
	}
	if req.Approved != 1 && req.Approved != 2 {
		return utils.Response.BusinessError("approval_result_invalid"), nil
	}

	// 2. 获取当前用户ID
	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}

	// 3. 获取审批记录（使用HandoverApprovalModel，与提交审批保持一致）
	approval, err := l.svcCtx.HandoverApprovalModel.FindOne(l.ctx, req.ApprovalID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("approval_not_found"), nil
		}
		return nil, err
	}

	// 4. 检查审批类型：必须是任务节点完成审批（ApprovalStep=3）
	if approval.ApprovalStep != 3 {
		return utils.Response.BusinessError("approval_type_invalid"), nil
	}

	// 5. 检查审批状态：只有待审批（状态0）的记录才能审批
	if approval.ApprovalType != 0 {
		return utils.Response.BusinessError("approval_already_done"), nil
	}

	// 6. 验证权限：只有审批人（项目负责人）可以审批
	if approval.ApproverId != employeeId {
		return utils.Response.BusinessError("approval_permission_denied"), nil
	}

	// 7. 获取任务节点ID
	taskNodeId := ""
	if approval.TaskNodeId.Valid {
		taskNodeId = approval.TaskNodeId.String
	}
	if taskNodeId == "" {
		return utils.Response.BusinessError("approval_missing_node_id"), nil
	}

	// 8. 获取任务节点信息
	taskNode, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, taskNodeId)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("task_node_not_found"), nil
		}
		return nil, err
	}

	// 9. 获取审批人姓名
	approverName := ""
	employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, employeeId)
	if err == nil {
		approverName = employee.RealName
	}

	// 10. 更新审批记录（使用HandoverApprovalModel）
	approval.ApprovalType = int64(req.Approved)
	approval.ApproverName = approverName
	approval.Comment = sql.NullString{String: req.Comment, Valid: req.Comment != ""}
	approval.UpdateTime = sql.NullTime{Time: time.Now(), Valid: true}
	err = l.svcCtx.HandoverApprovalModel.Update(l.ctx, approval)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新审批记录失败: %v", err)
		return nil, err
	}

	// 11. 如果审批通过，更新节点状态为已完成（状态2）并设置进度为100%
	if req.Approved == 1 {
		// 先更新节点完成时间和进度
		updatedNode := *taskNode
		updatedNode.NodeStatus = 2 // 设置状态为已完成
		updatedNode.Progress = 100 // 设置进度为100%
		updatedNode.NodeFinishTime = sql.NullTime{Time: time.Now(), Valid: true}
		updatedNode.UpdateTime = time.Now()
		err = l.svcCtx.TaskNodeModel.Update(l.ctx, &updatedNode)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务节点失败: %v", err)
			return nil, err
		}

		// 获取当前任务的所有节点
		current, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, taskNodeId)
		if err != nil {
			return nil, err
		}

		// 遍历查看哪个任务节点的前置节点是该节点，如果存在该节点且该节点状态为未开始(0)，则修正为进行中(1)
		nodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, current.TaskId)
		if err == nil && len(nodes) > 1 {
			// 遍历节点查看前置节点是否有该节点
			for _, node := range nodes {
				// 跳过当前已完成的节点，避免覆盖其状态
				if node.TaskNodeId == taskNodeId {
					continue
				}

				// 只有当节点状态为未开始(0)时，才检查是否应该激活
				if node.NodeStatus == 0 && node.ExNodeIds != "" {
					split := strings.Split(node.ExNodeIds, ",")
					for _, v := range split {
						if strings.TrimSpace(v) == taskNodeId {
							// 检查该节点的所有前置节点是否都已完成
							allPrerequisitesCompleted := true
							for _, preNodeId := range split {
								preNodeId = strings.TrimSpace(preNodeId)
								if preNodeId == "" {
									continue
								}
								preNode, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, preNodeId)
								if err != nil || preNode.NodeStatus != 2 {
									allPrerequisitesCompleted = false
									break
								}
							}

							// 只有当所有前置节点都已完成时，才将该节点状态更新为进行中
							if allPrerequisitesCompleted {
								l.svcCtx.TaskNodeModel.UpdateStatus(l.ctx, node.TaskNodeId, 1)
								l.Logger.WithContext(l.ctx).Infof("节点 %s 的所有前置节点已完成，状态更新为进行中", node.TaskNodeId)
							}
							break
						}
					}
				}
			}
		}

		// 更新任务整体进度
		err = l.updateTaskProgress(taskNodeId)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务整体进度失败: %v", err)
		}

		// 发送通知给节点执行人（支持多执行人）
		if l.svcCtx.NotificationMQService != nil && taskNode.ExecutorId != "" {
			// 拆分多个执行人ID
			executorIds := strings.Split(taskNode.ExecutorId, ",")
			notifyEmployees := make([]string, 0, len(executorIds))
			for _, id := range executorIds {
				trimmedId := strings.TrimSpace(id)
				if trimmedId != "" {
					notifyEmployees = append(notifyEmployees, trimmedId)
				}
			}

			if len(notifyEmployees) > 0 {
				notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
					svc.TaskNodeCompleted,
					notifyEmployees,
					taskNodeId,
					svc.NotificationEventOptions{TaskID: taskNode.TaskId, NodeID: taskNodeId},
				)
				notificationEvent.Title = "任务节点审批通过"
				notificationEvent.Content = fmt.Sprintf("任务节点 %s 的完成审批已通过", taskNode.NodeName)
				notificationEvent.Priority = 1
				if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
					l.Logger.WithContext(l.ctx).Errorf("发布通知事件失败: %v", err)
				}
			}
		}
	} else {
		// 如果审批拒绝，将节点状态改回进行中（状态1）并重置进度
		updatedNode := *taskNode
		updatedNode.NodeStatus = 1 // 设置状态为进行中
		updatedNode.Progress = 0   // 重置进度为0，要求重新完成
		updatedNode.UpdateTime = time.Now()
		err = l.svcCtx.TaskNodeModel.Update(l.ctx, &updatedNode)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务节点状态失败: %v", err)
			return nil, err
		}

		// 发送通知给节点执行人（支持多执行人）
		if l.svcCtx.NotificationMQService != nil && taskNode.ExecutorId != "" {
			// 拆分多个执行人ID
			executorIds := strings.Split(taskNode.ExecutorId, ",")
			notifyEmployees := make([]string, 0, len(executorIds))
			for _, id := range executorIds {
				trimmedId := strings.TrimSpace(id)
				if trimmedId != "" {
					notifyEmployees = append(notifyEmployees, trimmedId)
				}
			}

			if len(notifyEmployees) > 0 {
				notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
					svc.TaskNodeCompletionApproval,
					notifyEmployees,
					taskNodeId,
					svc.NotificationEventOptions{TaskID: taskNode.TaskId, NodeID: taskNodeId},
				)
				notificationEvent.Title = "任务节点审批被拒绝"
				notificationEvent.Content = fmt.Sprintf("任务节点 %s 的完成审批被拒绝，请继续完善工作", taskNode.NodeName)
				notificationEvent.Priority = 2
				if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
					l.Logger.WithContext(l.ctx).Errorf("发布通知事件失败: %v", err)
				}
			}
		}
	}

	// 12. 创建任务日志
	logContent := fmt.Sprintf("任务节点 %s 完成审批：%s", taskNode.NodeName, map[int]string{1: "通过", 2: "拒绝"}[req.Approved])
	if req.Comment != "" {
		logContent += fmt.Sprintf("，审批意见：%s", req.Comment)
	}
	taskLog := &task.TaskLog{
		LogId:      utils.Common.GenerateID(),
		TaskId:     taskNode.TaskId,
		TaskNodeId: utils.Common.ToSqlNullString(taskNodeId),
		LogType:    2, // 更新类型
		LogContent: logContent,
		EmployeeId: employeeId,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务日志失败: %v", err)
	}

	return utils.Response.Success(map[string]interface{}{
		"approvalId": req.ApprovalID,
		"approved":   req.Approved,
		"message":    fmt.Sprintf("审批%s成功", map[int]string{1: "通过", 2: "拒绝"}[req.Approved]),
	}), nil
}

// updateTaskProgress 根据所有任务节点进度更新任务整体进度
func (l *ApproveTaskNodeCompletionLogic) updateTaskProgress(taskNodeId string) error {
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
