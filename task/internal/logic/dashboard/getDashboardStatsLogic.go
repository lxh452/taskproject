package dashboard

import (
	"context"
	"time"

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
	TotalTasks        int64           `json:"totalTasks"`        // 总任务数（执行人+负责人）
	PendingApprovals  int64           `json:"pendingApprovals"`  // 待审批数量
	CompletedTasks    int64           `json:"completedTasks"`    // 已完成任务数
	CriticalTasks     int64           `json:"criticalTasks"`     // 紧急任务数
	OverdueTasks      int64           `json:"overdueTasks"`      // 逾期任务数
	AvgCompletionDays int64           `json:"avgCompletionDays"` // 平均完成天数
	OnTimeRate        int64           `json:"onTimeRate"`        // 按时完成率（百分比）
	ActiveMembers     int64           `json:"activeMembers"`     // 活跃成员数
	TaskTrend         []TaskTrendData `json:"taskTrend"`         // 任务趋势数据
}

// TaskTrendData 任务趋势数据点
type TaskTrendData struct {
	Date      string `json:"date"`      // 日期 YYYY-MM-DD
	Created   int64  `json:"created"`   // 创建的任务数
	Completed int64  `json:"completed"` // 完成的任务数
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

	// 5. 计算逾期任务数
	stats.OverdueTasks = l.getOverdueTaskCount(employeeID, departmentID, req.Scope)

	// 6. 计算平均完成天数
	stats.AvgCompletionDays = l.getAvgCompletionDays(employeeID, departmentID, req.Scope)

	// 7. 计算按时完成率
	stats.OnTimeRate = l.getOnTimeRate(employeeID, departmentID, req.Scope)

	// 8. 计算活跃成员数（仅部门范围有效）
	if req.Scope == "department" && departmentID != "" {
		stats.ActiveMembers = l.getActiveMembersCount(departmentID)
	}

	// 9. 计算任务趋势数据（最近7天）
	stats.TaskTrend = l.getTaskTrend(employeeID, 7)

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

// getOverdueTaskCount 获取逾期任务数
func (l *GetDashboardStatsLogic) getOverdueTaskCount(employeeID, departmentID, scope string) int64 {
	var count int64
	now := utils.Common.GetCurrentTime()

	if scope == "department" && departmentID != "" {
		// 部门范围：统计部门内所有逾期任务节点
		nodes, _, err := l.svcCtx.TaskNodeModel.FindByDepartment(l.ctx, departmentID, 1, 1000)
		if err == nil {
			for _, node := range nodes {
				// 未完成且已过截止时间
				if node.NodeStatus != 2 && !node.NodeDeadline.IsZero() {
					if node.NodeDeadline.Before(now) {
						count++
					}
				}
			}
		}
	} else {
		// 个人范围：统计自己作为执行人或负责人的逾期任务节点
		nodeMap := make(map[string]bool)

		// 获取作为执行人的逾期任务节点
		executorNodes, _, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, employeeID, 1, 1000)
		if err == nil {
			for _, node := range executorNodes {
				if node.NodeStatus != 2 && !node.NodeDeadline.IsZero() {
					if node.NodeDeadline.Before(now) {
						nodeMap[node.TaskNodeId] = true
					}
				}
			}
		}

		// 获取作为负责人的逾期任务节点
		leaderNodes, _, err := l.svcCtx.TaskNodeModel.FindByLeader(l.ctx, employeeID, 1, 1000)
		if err == nil {
			for _, node := range leaderNodes {
				if node.NodeStatus != 2 && !node.NodeDeadline.IsZero() {
					if node.NodeDeadline.Before(now) {
						nodeMap[node.TaskNodeId] = true
					}
				}
			}
		}

		count = int64(len(nodeMap))
	}

	return count
}

// getAvgCompletionDays 获取平均完成天数
func (l *GetDashboardStatsLogic) getAvgCompletionDays(employeeID, departmentID, scope string) int64 {
	var totalDays int64
	var completedCount int64

	if scope == "department" && departmentID != "" {
		// 部门范围
		nodes, _, err := l.svcCtx.TaskNodeModel.FindByDepartment(l.ctx, departmentID, 1, 1000)
		if err == nil {
			for _, node := range nodes {
				if node.NodeStatus == 2 { // 已完成
					days := node.UpdateTime.Sub(node.CreateTime).Hours() / 24
					if days > 0 {
						totalDays += int64(days)
						completedCount++
					}
				}
			}
		}
	} else {
		// 个人范围
		nodeMap := make(map[string]*struct {
			createTime string
			updateTime string
		})

		// 获取作为执行人的已完成任务节点
		executorNodes, _, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, employeeID, 1, 1000)
		if err == nil {
			for _, node := range executorNodes {
				if node.NodeStatus == 2 {
					nodeMap[node.TaskNodeId] = &struct {
						createTime string
						updateTime string
					}{
						createTime: node.CreateTime.Format("2006-01-02 15:04:05"),
						updateTime: node.UpdateTime.Format("2006-01-02 15:04:05"),
					}
				}
			}
		}

		// 获取作为负责人的已完成任务节点
		leaderNodes, _, err := l.svcCtx.TaskNodeModel.FindByLeader(l.ctx, employeeID, 1, 1000)
		if err == nil {
			for _, node := range leaderNodes {
				if node.NodeStatus == 2 {
					nodeMap[node.TaskNodeId] = &struct {
						createTime string
						updateTime string
					}{
						createTime: node.CreateTime.Format("2006-01-02 15:04:05"),
						updateTime: node.UpdateTime.Format("2006-01-02 15:04:05"),
					}
				}
			}
		}

		// 计算平均天数
		for _, times := range nodeMap {
			createTime, _ := utils.Common.ParseTime(times.createTime)
			updateTime, _ := utils.Common.ParseTime(times.updateTime)
			days := updateTime.Sub(createTime).Hours() / 24
			if days > 0 {
				totalDays += int64(days)
				completedCount++
			}
		}
	}

	if completedCount == 0 {
		return 0
	}
	return totalDays / completedCount
}

// getOnTimeRate 获取按时完成率（百分比）
func (l *GetDashboardStatsLogic) getOnTimeRate(employeeID, departmentID, scope string) int64 {
	var totalWithDeadline int64
	var onTimeCompleted int64

	if scope == "department" && departmentID != "" {
		// 部门范围
		nodes, _, err := l.svcCtx.TaskNodeModel.FindByDepartment(l.ctx, departmentID, 1, 1000)
		if err == nil {
			for _, node := range nodes {
				if node.NodeStatus == 2 && !node.NodeDeadline.IsZero() { // 已完成且有截止时间
					totalWithDeadline++
					if node.UpdateTime.Before(node.NodeDeadline) || node.UpdateTime.Equal(node.NodeDeadline) {
						onTimeCompleted++
					}
				}
			}
		}
	} else {
		// 个人范围
		nodeMap := make(map[string]*struct {
			updateTime string
			deadline   string
		})

		// 获取作为执行人的已完成任务节点
		executorNodes, _, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, employeeID, 1, 1000)
		if err == nil {
			for _, node := range executorNodes {
				if node.NodeStatus == 2 && !node.NodeDeadline.IsZero() {
					nodeMap[node.TaskNodeId] = &struct {
						updateTime string
						deadline   string
					}{
						updateTime: node.UpdateTime.Format("2006-01-02 15:04:05"),
						deadline:   node.NodeDeadline.Format("2006-01-02 15:04:05"),
					}
				}
			}
		}

		// 获取作为负责人的已完成任务节点
		leaderNodes, _, err := l.svcCtx.TaskNodeModel.FindByLeader(l.ctx, employeeID, 1, 1000)
		if err == nil {
			for _, node := range leaderNodes {
				if node.NodeStatus == 2 && !node.NodeDeadline.IsZero() {
					nodeMap[node.TaskNodeId] = &struct {
						updateTime string
						deadline   string
					}{
						updateTime: node.UpdateTime.Format("2006-01-02 15:04:05"),
						deadline:   node.NodeDeadline.Format("2006-01-02 15:04:05"),
					}
				}
			}
		}

		// 计算按时完成率
		for _, times := range nodeMap {
			totalWithDeadline++
			updateTime, _ := utils.Common.ParseTime(times.updateTime)
			deadline, _ := utils.Common.ParseTime(times.deadline)
			if updateTime.Before(deadline) || updateTime.Equal(deadline) {
				onTimeCompleted++
			}
		}
	}

	if totalWithDeadline == 0 {
		return 0
	}
	return (onTimeCompleted * 100) / totalWithDeadline
}

// getActiveMembersCount 获取活跃成员数（最近30天有更新任务的成员）
func (l *GetDashboardStatsLogic) getActiveMembersCount(departmentID string) int64 {
	// 获取部门所有员工
	employees, err := l.svcCtx.EmployeeModel.FindByDepartmentID(l.ctx, departmentID)
	if err != nil {
		l.Logger.Errorf("获取部门员工失败: %v", err)
		return 0
	}

	// 计算最近30天的时间点
	thirtyDaysAgo := utils.Common.GetCurrentTime().AddDate(0, 0, -30)
	activeMembers := make(map[string]bool)

	// 检查每个员工是否有最近30天更新的任务节点
	for _, emp := range employees {
		// 获取作为执行人的任务节点
		executorNodes, _, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, emp.Id, 1, 100)
		if err == nil {
			for _, node := range executorNodes {
				if node.UpdateTime.After(thirtyDaysAgo) {
					activeMembers[emp.Id] = true
					break
				}
			}
		}

		// 如果还没标记为活跃，检查作为负责人的任务节点
		if !activeMembers[emp.Id] {
			leaderNodes, _, err := l.svcCtx.TaskNodeModel.FindByLeader(l.ctx, emp.Id, 1, 100)
			if err == nil {
				for _, node := range leaderNodes {
					if node.UpdateTime.After(thirtyDaysAgo) {
						activeMembers[emp.Id] = true
						break
					}
				}
			}
		}
	}

	return int64(len(activeMembers))
}

// getTaskTrend 获取任务趋势数据
func (l *GetDashboardStatsLogic) getTaskTrend(employeeID string, days int) []TaskTrendData {
	var trendData []TaskTrendData
	now := time.Now()

	// 遍历每一天统计数据
	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		startTime := date.Format("2006-01-02 00:00:00")
		endTime := date.Format("2006-01-02 23:59:59")

		// 统计当天创建的任务节点数(作为执行人或负责人)
		createdCount := l.countTaskNodesByDate(employeeID, startTime, endTime, "create")

		// 统计当天完成的任务节点数(作为执行人或负责人)
		completedCount := l.countTaskNodesByDate(employeeID, startTime, endTime, "complete")

		trendData = append(trendData, TaskTrendData{
			Date:      dateStr,
			Created:   createdCount,
			Completed: completedCount,
		})
	}

	return trendData
}

// countTaskNodesByDate 统计指定日期范围内的任务节点数
func (l *GetDashboardStatsLogic) countTaskNodesByDate(employeeID, startTime, endTime, countType string) int64 {
	var count int64
	nodeMap := make(map[string]bool)

	// 获取作为执行人的任务节点
	executorNodes, _, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, employeeID, 1, 1000)
	if err == nil {
		for _, node := range executorNodes {
			var timeToCheck time.Time
			if countType == "create" {
				timeToCheck = node.CreateTime
			} else if countType == "complete" {
				if node.NodeStatus == 2 { // 已完成
					timeToCheck = node.UpdateTime
				} else {
					continue
				}
			}

			// 检查时间是否在范围内
			if l.isTimeInRange(timeToCheck, startTime, endTime) {
				nodeMap[node.TaskNodeId] = true
			}
		}
	}

	// 获取作为负责人的任务节点
	leaderNodes, _, err := l.svcCtx.TaskNodeModel.FindByLeader(l.ctx, employeeID, 1, 1000)
	if err == nil {
		for _, node := range leaderNodes {
			var timeToCheck time.Time
			if countType == "create" {
				timeToCheck = node.CreateTime
			} else if countType == "complete" {
				if node.NodeStatus == 2 { // 已完成
					timeToCheck = node.UpdateTime
				} else {
					continue
				}
			}

			// 检查时间是否在范围内
			if l.isTimeInRange(timeToCheck, startTime, endTime) {
				nodeMap[node.TaskNodeId] = true
			}
		}
	}

	count = int64(len(nodeMap))
	return count
}

// isTimeInRange 检查时间是否在指定范围内
func (l *GetDashboardStatsLogic) isTimeInRange(t time.Time, startStr, endStr string) bool {
	start, err1 := time.Parse("2006-01-02 15:04:05", startStr)
	end, err2 := time.Parse("2006-01-02 15:04:05", endStr)
	if err1 != nil || err2 != nil {
		return false
	}
	return (t.Equal(start) || t.After(start)) && (t.Equal(end) || t.Before(end))
}
