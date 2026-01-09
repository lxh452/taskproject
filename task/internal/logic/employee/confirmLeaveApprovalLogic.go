// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package employee

import (
	"context"
	"database/sql"
	"task_Project/model/task"
	"task_Project/task/internal/utils"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmLeaveApprovalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 确认离职审批
func NewConfirmLeaveApprovalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmLeaveApprovalLogic {
	return &ConfirmLeaveApprovalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ConfirmLeaveApproval 确认离职审批（审批通过后调用）
func (l *ConfirmLeaveApprovalLogic) ConfirmLeaveApproval(req *types.ConfirmLeaveApprovalRequest) (*types.BaseResponse, error) {
	// 1. 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.ApprovalID) {
		return utils.Response.ValidationError("审批ID不能为空"), nil
	}

	// 2. 查询审批记录
	approval, err := l.svcCtx.TaskHandoverModel.FindOne(l.ctx, req.ApprovalID)
	if err != nil {
		logx.Errorf("查询审批记录失败: %v", err)
		return utils.Response.ErrorWithKey("approval_not_found"), nil
	}

	// 3. 检查审批状态
	if approval.HandoverStatus != 1 { // 不是待处理状态
		return utils.Response.BusinessError("data_not_found"), nil
	}

	// 4. 获取员工信息
	employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, approval.FromEmployeeId)
	if err != nil {
		logx.Errorf("查询员工信息失败: %v", err)
		return utils.Response.ErrorWithKey("employee_not_found"), nil
	}

	// 5. 更新员工状态为离职
	updateData := map[string]interface{}{
		"status":     0, // 离职
		"leave_date": time.Now(),
	}

	err = l.svcCtx.EmployeeModel.SelectiveUpdate(l.ctx, approval.FromEmployeeId, updateData)
	if err != nil {
		logx.Errorf("更新员工离职状态失败: %v", err)
		return utils.Response.InternalError("更新员工离职状态失败"), err
	}

	// 6. 更新审批状态为已通过
	logx.Infof("离职审批 %s 已通过", req.ApprovalID)

	// 7. 处理员工的任务重新派发
	err = l.handleTaskRedispatch(approval.FromEmployeeId)
	if err != nil {
		logx.Errorf("处理任务重新派发失败: %v", err)
		// 不影响主流程，只记录错误
	}

	return utils.Response.Success(map[string]interface{}{
		"message":      "离职审批已通过，任务已重新派发",
		"employeeName": employee.RealName,
		"approvalId":   req.ApprovalID,
	}), nil
}

// handleTaskRedispatch 处理任务重新派发
func (l *ConfirmLeaveApprovalLogic) handleTaskRedispatch(employeeID string) error {
	// 1. 查找员工当前负责的任务节点
	taskNodes, _, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, employeeID, 1, 100)
	if err != nil {
		logx.Errorf("查询员工任务失败: %v", err)
		return err
	}

	// 2. 对每个进行中的任务节点进行处理
	for _, node := range taskNodes {
		if node.NodeStatus == 2 { // 进行中
			// 清空执行人，让任务进入闲置状态
			err = l.svcCtx.TaskNodeModel.UpdateExecutor(l.ctx, node.TaskNodeId, "")
			if err != nil {
				logx.Errorf("清空任务节点执行人失败: %v", err)
				continue
			}

			// 创建交接记录等待手动分配
			l.createTaskHandover(node, employeeID, "员工离职，任务待重新分配")
		}
	}

	return nil
}

// createTaskHandover 创建任务交接记录
func (l *ConfirmLeaveApprovalLogic) createTaskHandover(node *task.TaskNode, fromEmployeeID, reason string) error {
	handover := &task.TaskHandover{
		HandoverId:     utils.Common.GenId("handover"),
		TaskId:         node.TaskId,
		FromEmployeeId: fromEmployeeID,
		ToEmployeeId:   "", // 待分配
		HandoverReason: sql.NullString{String: reason, Valid: true},
		HandoverNote:   sql.NullString{String: "等待管理者分配接替者", Valid: true},
		HandoverStatus: 1, // 待处理
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}

	_, err := l.svcCtx.TaskHandoverModel.Insert(l.ctx, handover)
	return err
}
