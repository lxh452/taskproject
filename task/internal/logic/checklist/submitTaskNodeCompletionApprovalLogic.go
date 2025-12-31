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

	// 4. 检查节点状态：只有进行中（状态1）的节点才能提交审批
	if taskNode.NodeStatus != 1 {
		return nil, errors.New("只有进行中的任务节点才能提交审批")
	}

	// 5. 检查节点进度：只有进度100%的节点才能提交审批
	if taskNode.Progress < 100 {
		return nil, errors.New("任务节点进度未达到100%，无法提交审批")
	}

	// 6. 验证权限：只有节点执行人可以提交审批
	hasPermission := false
	// ExecutorId可能是逗号分隔的多个ID
	executorIdStr := taskNode.ExecutorId
	if executorIdStr != "" {
		executorIds := strings.Split(executorIdStr, ",")
		for _, executorId := range executorIds {
			if strings.TrimSpace(executorId) == employeeId {
				hasPermission = true
				break
			}
		}
	}
	if !hasPermission {
		return nil, errors.New("无权限提交审批，只有节点执行人可以提交")
	}

	// 7. 检查是否已有待审批的记录（使用HandoverApprovalModel）
	existingApproval, err := l.svcCtx.HandoverApprovalModel.FindLatestByTaskNodeId(l.ctx, req.NodeID)
	if err == nil && existingApproval != nil && existingApproval.ApprovalType == 0 {
		return nil, errors.New("该节点已有待审批的记录，请勿重复提交")
	}

	// 8. 使用通用的上级审核逻辑查找审批人
	approverFinder := utils.NewApproverFinder(
		l.svcCtx.EmployeeModel,
		l.svcCtx.DepartmentModel,
		l.svcCtx.CompanyModel,
	)

	approverResult, err := approverFinder.FindApprover(l.ctx, employeeId)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查找审批人失败: %v", err)
	}

	approverId := ""
	approverName := ""

	if approverResult != nil {
		approverId = approverResult.ApproverID
		approverName = approverResult.ApproverName
	} else {
		// 如果找不到上级，回退到任务负责人
		taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, taskNode.TaskId)
		if err != nil {
			if errors.Is(err, sqlx.ErrNotFound) {
				return nil, errors.New("任务不存在")
			}
			return nil, err
		}

		if taskInfo.LeaderId.Valid && taskInfo.LeaderId.String != "" {
			approverId = taskInfo.LeaderId.String
		} else {
			approverId = taskInfo.TaskCreator
		}

		if approverId != "" {
			employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, approverId)
			if err == nil {
				approverName = employee.RealName
			}
		}
	}

	if approverId == "" {
		return nil, errors.New("无法找到审批人，请先设置直属上级或部门经理")
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
