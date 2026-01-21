package tasknode

import (
	"context"
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

type UpdateTaskNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTaskNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTaskNodeLogic {
	return &UpdateTaskNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTaskNodeLogic) UpdateTaskNode(req *types.UpdateTaskNodeRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.NodeID == "" {
		return utils.Response.BusinessError("task_node_not_found"), nil
	}

	// 2. 获取当前用户ID
	currentEmpID, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 3. 获取任务节点信息
	taskNode, err := l.svcCtx.TaskNodeModel.FindOneSafe(l.ctx, req.NodeID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("task_node_not_found"), nil
		}
		return nil, err
	}

	// 4. 获取任务信息，用于验证任务负责人权限
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, taskNode.TaskId)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("获取任务信息失败: %v", err)
		return nil, err
	}

	// 5. 验证用户权限：节点负责人、执行人或任务负责人可以更新
	hasPermission := false

	// 检查是否是节点负责人
	if taskNode.LeaderId == currentEmpID {
		hasPermission = true
	}

	// 检查是否是任务负责人
	if !hasPermission && taskInfo.LeaderId.Valid && taskInfo.LeaderId.String == currentEmpID {
		hasPermission = true
	}

	// 检查是否是任务创建者
	if !hasPermission && taskInfo.TaskCreator == currentEmpID {
		hasPermission = true
	}

	// 检查是否是节点执行人（ExecutorId可能是逗号分隔的多个ID）
	if !hasPermission {
		executorIdStr := taskNode.ExecutorId
		if executorIdStr != "" {
			executorIds := strings.Split(executorIdStr, ",")
			for _, executorId := range executorIds {
				if strings.TrimSpace(executorId) == currentEmpID {
					hasPermission = true
					break
				}
			}
		}
	}

	if !hasPermission {
		return utils.Response.BusinessError("permission_denied"), nil
	}

	// 5. 构建更新数据
	updateData := make(map[string]interface{})
	updateFields := []string{}

	// 更新节点名称
	if req.NodeName != "" {
		updateData["node_name"] = req.NodeName
		updateFields = append(updateFields, "节点名称")
	}

	// 更新节点详情  只有节点负责人或任务负责人可以修改
	canUpdateDetail := taskNode.LeaderId == currentEmpID ||
		(taskInfo.LeaderId.Valid && taskInfo.LeaderId.String == currentEmpID) ||
		taskInfo.TaskCreator == currentEmpID
	if req.NodeDetail != "" && canUpdateDetail {
		updateData["node_detail"] = req.NodeDetail
		updateFields = append(updateFields, "节点详情")
	}

	// 更新执行人 只有节点负责人或任务负责人可以修改
	canUpdateExecutor := taskNode.LeaderId == currentEmpID ||
		(taskInfo.LeaderId.Valid && taskInfo.LeaderId.String == currentEmpID) ||
		taskInfo.TaskCreator == currentEmpID
	if len(req.ExecutorID) > 0 && canUpdateExecutor {
		// 支持多个执行人，用逗号分隔
		executorIDs := make([]string, 0)
		for _, executorID := range req.ExecutorID {
			executorID = strings.TrimSpace(executorID)
			if executorID == "" {
				continue
			}
			// 验证执行人是否存在
			_, err = l.svcCtx.EmployeeModel.FindOne(l.ctx, executorID)
			if err != nil {
				if errors.Is(err, sqlx.ErrNotFound) {
					return utils.Response.BusinessError("executor_not_found"), nil
				}
				return nil, err
			}
			executorIDs = append(executorIDs, executorID)
		}
		if len(executorIDs) > 0 {
			newExecutorID := strings.Join(executorIDs, ",")
			updateData["executor_id"] = newExecutorID
			updateFields = append(updateFields, "执行人")
		}
	}

	// 更新节点状态
	if len(req.NodeStatus) > 0 {
		updateData["node_status"] = req.NodeStatus[0]
		updateFields = append(updateFields, "节点状态")
	}

	// 更新截止时间
	if req.NodeDeadline != "" {
		deadline, err := time.Parse("2006-01-02", req.NodeDeadline)
		if err != nil {
			return utils.Response.BusinessError("Format error"), nil
		}
		updateData["node_deadline"] = deadline
		updateFields = append(updateFields, "截止时间")
	}

	// 更新完成时间
	if req.NodeFinishTime != "" {
		finishTime, err := time.Parse("2006-01-02 15:04:05", req.NodeFinishTime)
		if err != nil {
			return utils.Response.BusinessError("Format error"), nil
		}
		updateData["node_finish_time"] = finishTime
		updateFields = append(updateFields, "完成时间")
	}

	// 更新前置节点（保存到 ex_node_ids）
	if req.PrerequisiteNodes != "" {
		if err := l.svcCtx.TaskNodeModel.UpdateExNodeIds(l.ctx, req.NodeID, req.PrerequisiteNodes); err != nil {
			return utils.Response.InternalError("更新前置节点失败"), nil
		}
		updateFields = append(updateFields, "前置节点")
	}

	if len(updateData) == 0 {
		return utils.Response.BusinessError("invalid_params"), nil
	}

	updateData["update_time"] = time.Now()

	// 6. 更新任务节点
	updatedTaskNode := *taskNode
	if req.NodeName != "" {
		updatedTaskNode.NodeName = req.NodeName
	}
	canUpdateDetail = taskNode.LeaderId == currentEmpID ||
		(taskInfo.LeaderId.Valid && taskInfo.LeaderId.String == currentEmpID) ||
		taskInfo.TaskCreator == currentEmpID
	if req.NodeDetail != "" && canUpdateDetail {
		updatedTaskNode.NodeDetail = utils.Common.ToSqlNullString(req.NodeDetail)
	}
	canUpdateExecutor = taskNode.LeaderId == currentEmpID ||
		(taskInfo.LeaderId.Valid && taskInfo.LeaderId.String == currentEmpID) ||
		taskInfo.TaskCreator == currentEmpID
	if len(req.ExecutorID) > 0 && canUpdateExecutor {
		// 支持多个执行人，用逗号分隔
		executorIDs := make([]string, 0)
		for _, executorID := range req.ExecutorID {
			executorID = strings.TrimSpace(executorID)
			if executorID != "" {
				executorIDs = append(executorIDs, executorID)
			}
		}
		if len(executorIDs) > 0 {
			updatedTaskNode.ExecutorId = strings.Join(executorIDs, ",")
		}
	}
	if len(req.NodeStatus) > 0 {
		updatedTaskNode.NodeStatus = int64(req.NodeStatus[0])
	}
	if req.NodeDeadline != "" {
		deadline, err := time.Parse("2006-01-02", req.NodeDeadline)
		if err == nil {
			updatedTaskNode.NodeDeadline = deadline
		}
	}
	if req.NodeFinishTime != "" {
		finishTime, err := time.Parse("2006-01-02 15:04:05", req.NodeFinishTime)
		if err == nil {
			updatedTaskNode.NodeFinishTime = utils.Common.ToSqlNullTime(finishTime.Format("2006-01-02 15:04:05"))
		}
	}
	if req.PrerequisiteNodes != "" {
		// 暂时注释掉，因为字段名可能不匹配
		// updatedTaskNode.PrerequisiteNodes = utils.Common.ToSqlNullString(req.PrerequisiteNodes)
	}
	updatedTaskNode.UpdateTime = time.Now()

	err = l.svcCtx.TaskNodeModel.Update(l.ctx, &updatedTaskNode)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务节点失败: %v", err)
		return nil, err
	}

	// 6.5 如果更新了节点状态，同步更新任务整体进度
	if len(req.NodeStatus) > 0 {
		err = l.updateTaskProgress(req.NodeID)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务整体进度失败: %v", err)
			// 不影响主流程，继续执行
		}
	}

	// 7. 创建任务日志
	logContent := fmt.Sprintf("任务节点 %s 已更新：%s", taskNode.NodeName, fmt.Sprintf("%v", updateFields))
	taskLog := &task.TaskLog{
		LogId:      utils.Common.GenerateID(),
		TaskId:     taskNode.TaskId,
		LogType:    2, // 更新类型
		LogContent: logContent,
		EmployeeId: currentEmpID,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务日志失败: %v", err)
	}

	// 8. 如果更新了执行人，发送通知和邮件（通过消息队列，消费者会查询并发送）
	newExecutorID := ""
	if len(req.ExecutorID) > 0 {
		executorIDs := make([]string, 0)
		for _, executorID := range req.ExecutorID {
			executorID = strings.TrimSpace(executorID)
			if executorID != "" {
				executorIDs = append(executorIDs, executorID)
			}
		}
		if len(executorIDs) > 0 {
			newExecutorID = strings.Join(executorIDs, ",")
		}
	}
	if newExecutorID != "" && newExecutorID != taskNode.ExecutorId {
		// 发布邮件事件（消费者会查询新执行人并发送）
		if l.svcCtx.EmailMQService != nil {
			emailEvent := &svc.EmailEvent{
				EventType: "task.node.executor.changed",
				NodeID:    req.NodeID,
			}
			if err := l.svcCtx.EmailMQService.PublishEmailEvent(l.ctx, emailEvent); err != nil {
				l.Logger.WithContext(l.ctx).Errorf("发布邮件事件失败: %v", err)
			}
		}

		// 发布通知事件（消费者会查询新执行人并创建通知）
		if l.svcCtx.NotificationMQService != nil {
			event := &svc.NotificationEvent{
				EventType:   "task.node.executor.changed",
				NodeID:      req.NodeID,
				Type:        3,
				Category:    "handover",
				Priority:    2,
				RelatedID:   req.NodeID,
				RelatedType: "task",
			}
			if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, event); err != nil {
				l.Logger.WithContext(l.ctx).Errorf("发布通知事件失败: %v", err)
			}
		}
	}

	return utils.Response.Success(map[string]interface{}{
		"taskNodeId":    req.NodeID,
		"message":       "任务节点更新成功",
		"updatedFields": updateFields,
	}), nil
}

// updateTaskProgress 根据所有任务节点进度更新任务整体进度
func (l *UpdateTaskNodeLogic) updateTaskProgress(taskNodeId string) error {
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

	// 计算平均进度和完成节点数
	// 节点状态：0-未开始，1-进行中，2-已完成，3-已逾期
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
	// 任务状态：0-未开始，1-进行中，2-已完成，3-逾期完成
	if allNodesCompleted && avgProgress == 100 {
		err = l.svcCtx.TaskModel.UpdateStatus(l.ctx, taskNode.TaskId, 2)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务状态失败: %v", err)
		}
		l.Logger.WithContext(l.ctx).Infof("任务 %s 所有节点已完成，任务状态更新为已完成", taskNode.TaskId)
	} else if completedCount > 0 || avgProgress > 0 {
		// 如果有节点完成或有进度，但不是全部完成，确保任务状态为进行中
		taskInfo, taskErr := l.svcCtx.TaskModel.FindOne(l.ctx, taskNode.TaskId)
		if taskErr == nil && taskInfo.TaskStatus == 0 {
			err = l.svcCtx.TaskModel.UpdateStatus(l.ctx, taskNode.TaskId, 1)
			if err != nil {
				l.Logger.WithContext(l.ctx).Errorf("更新任务状态失败: %v", err)
			}
			l.Logger.WithContext(l.ctx).Infof("任务 %s 有进展，任务状态更新为进行中", taskNode.TaskId)
		}
	}

	// 更新任务节点统计
	err = l.svcCtx.TaskModel.UpdateNodeCount(l.ctx, taskNode.TaskId, int64(len(nodes)), completedCount)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务节点统计失败: %v", err)
	}

	return nil
}
