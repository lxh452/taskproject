package dashboard

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDashboardStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDashboardStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDashboardStatsLogic {
	return &GetDashboardStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DashboardStats 仪表盘统计数据
type DashboardStats struct {
	TotalTasks       int64 `json:"totalTasks"`       // 总任务数（执行人+负责人）
	PendingApprovals int64 `json:"pendingApprovals"` // 待审批数量
	CompletedTasks   int64 `json:"completedTasks"`   // 已完成任务数
	CriticalTasks    int64 `json:"criticalTasks"`    // 紧急任务数
}

func (l *GetDashboardStatsLogic) GetDashboardStats(req *types.GetDashboardStatsRequest) (resp *types.BaseResponse, err error) {
	// 获取当前员工ID
	employeeID, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeID == "" {
		return utils.Response.UnauthorizedError(), nil
	}

	// 获取员工信息
	employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, employeeID)
	if err != nil {
		l.Logger.Errorf("获取员工信息失败: %v", err)
		return utils.Response.BusinessError("employee_not_found"), nil
	}

	stats := DashboardStats{}

	// 1. 计算总任务数（作为执行人和负责人的任务节点，去重）
	stats.TotalTasks = l.getTotalTaskCount(employeeID)

	// 2. 计算待审批数量
	stats.PendingApprovals = l.getPendingApprovalCount(employeeID)

	// 3. 计算已完成任务数（可以是部门或个人）
	departmentID := ""
	if employee.DepartmentId.Valid {
		departmentID = employee.DepartmentId.String
	}
	stats.CompletedTasks = l.getCompletedTaskCount(employeeID, departmentID, req.Scope)

	// 4. 计算紧急任务数（可以是部门或个人）
	stats.CriticalTasks = l.getCriticalTaskCount(employeeID, departmentID, req.Scope)

	return utils.Response.Success(stats), nil
}

// getTotalTaskCount 获取总任务数（执行人+负责人，去重）
func (l *GetDashboardStatsLogic) getTotalTaskCount(employeeID string) int64 {
	// 使用已有的 GetTaskNodeCountByEmployee 方法，它已经处理了去重
	count, err := l.svcCtx.TaskNodeModel.GetTaskNodeCountByEmployee(l.ctx, employeeID)
	if err != nil {
		l.Logger.Errorf("获取任务节点数量失败: %v", err)
		return 0
	}
	return count
}

// getPendingApprovalCount 获取待审批数量（自己发起的+审批人是自己的，只统计未审批的）
func (l *GetDashboardStatsLogic) getPendingApprovalCount(employeeID string) int64 {
	var count int64

	// 1. 获取任务节点完成审批（审批人是自己，且状态为待审批）
	nodeApprovals, _, err := l.svcCtx.HandoverApprovalModel.FindTaskNodeApprovalsByApprover(l.ctx, employeeID, 1, 1000)
	if err == nil {
		for _, approval := range nodeApprovals {
			if approval.ApprovalType == 0 { // 0-待审批
				count++
			}
		}
	}

	// 2. 获取交接审批（审批人是自己，且状态为待审批）
	handovers, _, err := l.svcCtx.TaskHandoverModel.FindByEmployeeInvolved(l.ctx, employeeID, 1, 1000)
	if err == nil {
		for _, h := range handovers {
			// 只统计待审批状态(1)且审批人是自己的
			if h.HandoverStatus == 1 && h.ApproverId.Valid && h.ApproverId.String == employeeID {
				count++
			}
		}
	}

	// 3. 获取自己发起的待审批（交接申请，状态为待审批）
	myHandovers, _, err := l.svcCtx.TaskHandoverModel.FindByFromEmployee(l.ctx, employeeID, 1, 1000)
	if err == nil {
		for _, h := range myHandovers {
			// 自己发起的，状态为待审批(1)
			if h.HandoverStatus == 1 {
				count++
			}
		}
	}

	// 4. 获取员工加入申请审批（如果当前用户是管理者）
	// 检查当前用户是否有审批权限（是否是部门经理或公司管理员）
	employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, employeeID)
	if err == nil && employee.CompanyId != "" && employee.DepartmentId.Valid {
		dept, err := l.svcCtx.DepartmentModel.FindOne(l.ctx, employee.DepartmentId.String)
		if err == nil && dept.ManagerId.Valid && dept.ManagerId.String == employeeID {
			// 是部门经理，统计待审批的加入申请
			status := 0 // 待审批状态
			pendingApps, err := l.svcCtx.JoinApplicationModel.FindByCompanyId(l.ctx, employee.CompanyId, &status)
			if err == nil {
				count += int64(len(pendingApps))
			}
		}
	}

	return count
}

// getCompletedTaskCount 获取已完成任务数
func (l *GetDashboardStatsLogic) getCompletedTaskCount(employeeID, departmentID, scope string) int64 {
	var count int64

	if scope == "department" && departmentID != "" {
		// 部门范围：统计部门内所有已完成的任务节点
		// 直接使用 FindByDepartment 方法获取部门的任务节点
		nodes, _, err := l.svcCtx.TaskNodeModel.FindByDepartment(l.ctx, departmentID, 1, 1000)
		if err == nil {
			for _, node := range nodes {
				if node.NodeStatus == 2 { // 已完成
					count++
				}
			}
		}
	} else {
		// 个人范围：统计自己作为执行人或负责人的已完成任务节点
		nodeMap := make(map[string]bool)

		// 获取作为执行人的已完成任务节点
		executorNodes, _, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, employeeID, 1, 1000)
		if err == nil {
			for _, node := range executorNodes {
				if node.NodeStatus == 2 { // 已完成
					nodeMap[node.TaskNodeId] = true
				}
			}
		}

		// 获取作为负责人的已完成任务节点
		leaderNodes, _, err := l.svcCtx.TaskNodeModel.FindByLeader(l.ctx, employeeID, 1, 1000)
		if err == nil {
			for _, node := range leaderNodes {
				if node.NodeStatus == 2 { // 已完成
					nodeMap[node.TaskNodeId] = true
				}
			}
		}

		count = int64(len(nodeMap))
	}

	return count
}

// getCriticalTaskCount 获取紧急任务数
func (l *GetDashboardStatsLogic) getCriticalTaskCount(employeeID, departmentID, scope string) int64 {
	var count int64

	if scope == "department" && departmentID != "" {
		// 部门范围：统计部门内所有紧急任务节点（未完成的）
		// 先获取部门的任务节点
		nodes, _, err := l.svcCtx.TaskNodeModel.FindByDepartment(l.ctx, departmentID, 1, 1000)
		if err == nil {
			for _, node := range nodes {
				if node.NodeStatus != 2 && node.NodePriority == 1 { // 未完成且紧急
					count++
				}
			}
		}
	} else {
		// 个人范围：统计自己作为执行人或负责人的紧急任务节点
		nodeMap := make(map[string]bool)

		// 获取作为执行人的紧急任务节点
		executorNodes, _, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, employeeID, 1, 1000)
		if err == nil {
			for _, node := range executorNodes {
				if node.NodeStatus != 2 && node.NodePriority == 1 { // 未完成且紧急
					nodeMap[node.TaskNodeId] = true
				}
			}
		}

		// 获取作为负责人的紧急任务节点
		leaderNodes, _, err := l.svcCtx.TaskNodeModel.FindByLeader(l.ctx, employeeID, 1, 1000)
		if err == nil {
			for _, node := range leaderNodes {
				if node.NodeStatus != 2 && node.NodePriority == 1 { // 未完成且紧急
					nodeMap[node.TaskNodeId] = true
				}
			}
		}

		count = int64(len(nodeMap))
	}

	return count
}
