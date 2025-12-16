package checklist

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
		return utils.Response.BusinessError("任务节点ID不能为空"), nil
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
			return utils.Response.BusinessError("任务节点不存在"), nil
		}
		return nil, err
	}

	// 4. 检查节点状态：只有进行中（状态1）的节点才能提交审批
	if taskNode.NodeStatus != 1 {
		return utils.Response.BusinessError("只有进行中的任务节点才能提交审批"), nil
	}

	// 5. 检查节点进度：只有进度100%的节点才能提交审批
	if taskNode.Progress < 100 {
		return utils.Response.BusinessError("任务节点进度未达到100%，无法提交审批"), nil
	}

	// 6. 验证权限：只有节点执行人可以提交审批
	hasPermission := false
	executorIds := []string{taskNode.ExecutorId}
	for _, executorId := range executorIds {
		if executorId == employeeId {
			hasPermission = true
			break
		}
	}
	if !hasPermission {
		return utils.Response.BusinessError("无权限提交审批，只有节点执行人可以提交"), nil
	}

	// 7. 检查是否已有待审批的记录
	existingApproval, err := l.svcCtx.TaskNodeCompletionApprovalModel.FindLatestByTaskNodeId(l.ctx, req.NodeID)
	if err == nil && existingApproval != nil && existingApproval.ApprovalType == 0 {
		return utils.Response.BusinessError("该节点已有待审批的记录，请勿重复提交"), nil
	}

	// 8. 获取任务信息，找到项目负责人（任务负责人）
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, taskNode.TaskId)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("任务不存在"), nil
		}
		return nil, err
	}

	// 获取项目负责人（任务负责人）
	approverId := ""
	if taskInfo.LeaderId.Valid {
		approverId = taskInfo.LeaderId.String
	} else {
		// 如果没有任务负责人，使用任务创建者
		approverId = taskInfo.TaskCreator
	}

	if approverId == "" {
		return utils.Response.BusinessError("无法找到项目负责人，无法提交审批"), nil
	}

	// 获取审批人姓名
	approverName := ""
	if approverId != "" {
		employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, approverId)
		if err == nil {
			approverName = employee.RealName
		}
	}

	// 9. 创建审批记录
	approvalId := utils.Common.GenId("approval")
	approval := &task.TaskNodeCompletionApproval{
		ApprovalId:   approvalId,
		TaskNodeId:   req.NodeID,
		ApproverId:   approverId,
		ApproverName: approverName,
		ApprovalType: 0, // 0-待审批
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}
	_, err = l.svcCtx.TaskNodeCompletionApprovalModel.Insert(l.ctx, approval)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建审批记录失败: %v", err)
		return nil, err
	}

	// 10. 设置节点状态为待审批（状态4）
	err = l.svcCtx.TaskNodeModel.UpdateStatus(l.ctx, req.NodeID, 4)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新节点状态失败: %v", err)
		return nil, err
	}

	// 11. 创建任务日志
	taskLog := &task.TaskLog{
		LogId:      utils.Common.GenerateID(),
		TaskId:     taskNode.TaskId,
		TaskNodeId: utils.Common.ToSqlNullString(req.NodeID),
		LogType:    2, // 更新类型
		LogContent: fmt.Sprintf("任务节点 %s 已提交完成审批，等待项目负责人审批", taskNode.NodeName),
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
			req.NodeID,
			svc.NotificationEventOptions{TaskID: taskNode.TaskId, NodeID: req.NodeID},
		)
		notificationEvent.Title = "任务节点完成审批"
		notificationEvent.Content = fmt.Sprintf("任务节点 %s 已完成，等待您的审批", taskNode.NodeName)
		notificationEvent.Priority = 2
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
