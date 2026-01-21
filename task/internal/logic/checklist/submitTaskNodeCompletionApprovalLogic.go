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

type SubmitTaskNodeCompletionApprovalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitTaskNodeCompletionApprovalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitTaskNodeCompletionApprovalLogic {
	return &SubmitTaskNodeCompletionApprovalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitTaskNodeCompletionApprovalLogic) SubmitTaskNodeCompletionApproval(req *types.SubmitTaskNodeCompletionApprovalRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.NodeID == "" {
		return utils.Response.ValidationError("任务节点ID不能为空"), nil
	}

	// 2. 获取当前用户ID
	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}

	// 3. 获取任务节点信息
	taskNode, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, req.NodeID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("task_node_not_found"), nil
		}
		return nil, err
	}

	// 4. 检查节点状态：只有进行中（状态1）或已完成（状态2）的节点才能提交审批
	// 允许已完成的节点重新提交，以支持新增任务后需要重新审批的场景
	if taskNode.NodeStatus != 1 && taskNode.NodeStatus != 2 {
		return nil, errors.New("只有进行中或已完成的任务节点才能提交审批")
	}

	// 5. 检查节点进度：只有进度100%的节点才能提交审批
	if taskNode.Progress < 100 {
		return nil, errors.New("任务节点进度未达到100%，无法提交审批")
	}

	// 6. 验证权限：只有节点执行人可以提交审批
	hasPermission := false
	// ExecutorId可能是逗号分隔的多个ID
	executorIdStr := taskNode.ExecutorId
	l.Logger.WithContext(l.ctx).Infof("[权限检查] 当前用户ID: %s, 节点执行人ID: %s", employeeId, executorIdStr)
	if executorIdStr != "" {
		executorIds := strings.Split(executorIdStr, ",")
		for _, executorId := range executorIds {
			trimmedExecutorId := strings.TrimSpace(executorId)
			l.Logger.WithContext(l.ctx).Infof("[权限检查] 比较: 当前用户=%s, 执行人=%s", employeeId, trimmedExecutorId)
			if trimmedExecutorId == employeeId {
				hasPermission = true
				l.Logger.WithContext(l.ctx).Infof("[权限检查] 权限验证通过")
				break
			}
		}
	}
	if !hasPermission {
		l.Logger.WithContext(l.ctx).Errorf("[权限检查] 权限验证失败: 当前用户ID=%s 不在执行人列表中=%s", employeeId, executorIdStr)
		return nil, errors.New("无权限提交审批，只有节点执行人可以提交")
	}

	// 7. 检查是否已有待审批的记录（使用HandoverApprovalModel）
	// 注意：只检查待审批状态(0)，已通过(1)和已拒绝(2)的都允许重新提交
	// 这样可以支持：1) 审批被拒后重新提交 2) 已完成的节点因新增任务需要重新审批
	existingApproval, err := l.svcCtx.HandoverApprovalModel.FindLatestByTaskNodeId(l.ctx, req.NodeID)
	l.Logger.WithContext(l.ctx).Infof("[审批检查] 节点ID: %s, 查询结果: err=%v, approval=%+v", req.NodeID, err, existingApproval)
	if err == nil && existingApproval != nil {
		l.Logger.WithContext(l.ctx).Infof("[审批检查] 找到审批记录: ApprovalId=%s, ApprovalType=%d (0=待审批,1=已通过,2=已拒绝)", existingApproval.ApprovalId, existingApproval.ApprovalType)
		if existingApproval.ApprovalType == 0 {
			// 只有待审批状态才阻止重复提交
			l.Logger.WithContext(l.ctx).Errorf("[审批检查] 阻止提交: 存在待审批记录 ApprovalId=%s", existingApproval.ApprovalId)
			return nil, errors.New("该节点已有待审批的记录，请等待审批完成后再提交")
		}
		// 如果上次审批已通过(1)或被拒绝(2)，允许重新提交
		// 这样可以处理节点需要重新审批的场景（如新增任务后需要重新确认完成）
		l.Logger.WithContext(l.ctx).Infof("[审批检查] 允许重新提交: 上次审批状态=%d (非待审批状态)", existingApproval.ApprovalType)
	} else {
		l.Logger.WithContext(l.ctx).Infof("[审批检查] 未找到审批记录或查询出错，允许提交")
	}

	// 8. 查找审批人：使用节点负责人（leader_id）
	approverId := ""
	approverName := ""

	// 优先使用节点负责人
	if taskNode.LeaderId != "" {
		approverId = taskNode.LeaderId
		l.Logger.WithContext(l.ctx).Infof("[审批人查找] 使用节点负责人: %s", approverId)
	}

	// 如果节点没有负责人，返回错误
	if approverId == "" {
		return nil, errors.New("该任务节点未设置负责人，无法提交审批")
	}

	// 获取审批人姓名
	if approverId != "" {
		employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, approverId)
		if err == nil {
			approverName = employee.RealName
		} else {
			l.Logger.WithContext(l.ctx).Errorf("获取审批人信息失败: %v", err)
			approverName = "未知"
		}
	}

	// 8.5. 检查审批人是否是提交人自己
	if approverId == employeeId {
		// 如果审批人就是提交人自己，直接自动通过审批
		l.Logger.WithContext(l.ctx).Infof("节点执行人与审批人相同，自动通过审批: nodeId=%s, employeeId=%s", req.NodeID, employeeId)

		// 直接更新节点状态为已完成
		updatedNode := *taskNode
		updatedNode.NodeStatus = 2 // 设置状态为已完成
		updatedNode.Progress = 100 // 确保进度为100%
		updatedNode.NodeFinishTime = sql.NullTime{Time: time.Now(), Valid: true}
		updatedNode.UpdateTime = time.Now()
		err = l.svcCtx.TaskNodeModel.Update(l.ctx, &updatedNode)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务节点失败: %v", err)
			return nil, err
		}

		// 创建已通过的审批记录
		approvalId := utils.Common.GenId("approval")
		approval := &task.HandoverApproval{
			ApprovalId:   approvalId,
			HandoverId:   "",
			TaskNodeId:   sql.NullString{String: req.NodeID, Valid: true},
			ApprovalStep: 3,
			ApproverId:   approverId,
			ApproverName: approverName,
			ApprovalType: 1, // 1-已同意（自动通过）
			Comment:      sql.NullString{String: "执行人与审批人相同，自动通过", Valid: true},
			CreateTime:   time.Now(),
			UpdateTime:   sql.NullTime{Time: time.Now(), Valid: true},
		}
		_, err = l.svcCtx.HandoverApprovalModel.Insert(l.ctx, approval)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("创建审批记录失败: %v", err)
		}

		// 创建任务日志
		taskLog := &task.TaskLog{
			LogId:      utils.Common.GenerateID(),
			TaskId:     taskNode.TaskId,
			TaskNodeId: utils.Common.ToSqlNullString(req.NodeID),
			LogType:    2,
			LogContent: fmt.Sprintf("任务节点 %s 已完成（执行人与审批人相同，自动通过）", taskNode.NodeName),
			EmployeeId: employeeId,
			CreateTime: time.Now(),
		}
		_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("创建任务日志失败: %v", err)
		}

		// 检查并激活后续节点
		l.activateSubsequentNodes(req.NodeID, taskNode.TaskId)

		// 更新任务整体进度
		err = l.updateTaskProgress(req.NodeID)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务整体进度失败: %v", err)
		}

		return utils.Response.Success(map[string]interface{}{
			"approvalId":   approvalId,
			"nodeId":       req.NodeID,
			"message":      "任务节点已完成（自动通过审批）",
			"autoApproved": true,
		}), nil
	}

	// 9. 创建审批记录（使用HandoverApprovalModel）
	approvalId := utils.Common.GenId("approval")
	approval := &task.HandoverApproval{
		ApprovalId:   approvalId,
		HandoverId:   "",                                              // 任务节点完成审批不关联交接
		TaskNodeId:   sql.NullString{String: req.NodeID, Valid: true}, // 关联任务节点
		ApprovalStep: 3,                                               // 3-任务节点完成审批
		ApproverId:   approverId,
		ApproverName: approverName,
		ApprovalType: 0, // 0-待审批
		CreateTime:   time.Now(),
		UpdateTime:   sql.NullTime{Time: time.Now(), Valid: true},
	}
	_, err = l.svcCtx.HandoverApprovalModel.Insert(l.ctx, approval)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建审批记录失败: %v", err)
		return nil, err
	}

	// 10. 注意：节点状态保持为进行中（状态1），不改为待审批状态
	// 审批状态通过审批记录（HandoverApproval）来管理，而不是节点状态

	// 11. 创建任务日志
	taskLog := &task.TaskLog{
		LogId:      utils.Common.GenerateID(),
		TaskId:     taskNode.TaskId,
		TaskNodeId: utils.Common.ToSqlNullString(req.NodeID),
		LogType:    2, // 更新类型
		LogContent: fmt.Sprintf("任务节点 %s 已提交完成审批，等待%s审批", taskNode.NodeName, approverName),
		EmployeeId: employeeId,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务日志失败: %v", err)
	}

	// 12. 发送通知给项目负责人（通过消息队列）
	if l.svcCtx.NotificationMQService != nil && approverId != "" {
		notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
			svc.TaskNodeCompletionApproval,
			[]string{approverId},
			approvalId, // 使用审批ID作为RelatedID，便于前端直接获取审批记录
			svc.NotificationEventOptions{TaskID: taskNode.TaskId, NodeID: taskNode.TaskNodeId},
		)
		notificationEvent.Title = "任务节点完成审批"
		notificationEvent.Content = fmt.Sprintf("任务节点 %s 已完成，等待您的审批", taskNode.NodeName)
		notificationEvent.Priority = 2
		notificationEvent.Category = "task_approval" // 设置分类，便于前端过滤
		if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
			l.Logger.WithContext(l.ctx).Errorf("发布审批通知事件失败: %v", err)
		}
	}

	return utils.Response.Success(map[string]interface{}{
		"approvalId": approvalId,
		"nodeId":     req.NodeID,
		"message":    "提交审批成功，等待项目负责人审批",
	}), nil
}

// activateSubsequentNodes 激活后续节点
func (l *SubmitTaskNodeCompletionApprovalLogic) activateSubsequentNodes(completedNodeId, taskId string) {
	nodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, taskId)
	if err != nil || len(nodes) <= 1 {
		return
	}

	for _, node := range nodes {
		// 跳过当前已完成的节点
		if node.TaskNodeId == completedNodeId {
			continue
		}

		// 只有当节点状态为未开始(0)时，才检查是否应该激活
		if node.NodeStatus == 0 && node.ExNodeIds != "" {
			split := strings.Split(node.ExNodeIds, ",")
			for _, v := range split {
				if strings.TrimSpace(v) == completedNodeId {
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

// updateTaskProgress 根据所有任务节点进度更新任务整体进度
func (l *SubmitTaskNodeCompletionApprovalLogic) updateTaskProgress(taskNodeId string) error {
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
