package task

import (
	"context"
	"strings"
	"time"

	"task_Project/model/task"
	"task_Project/model/user"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type AutoDispatchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 任务自动派发
func NewAutoDispatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AutoDispatchLogic {
	return &AutoDispatchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AutoDispatchResult 自动派发结果
type AutoDispatchResult struct {
	TaskID          string                       `json:"taskId"`
	TaskTitle       string                       `json:"taskTitle"`
	Recommendations []svc.DispatchRecommendation `json:"recommendations"`
	Message         string                       `json:"message"`
}

// AutoDispatch 自动派发 - 使用GLM-4模型推荐最优人选
func (l *AutoDispatchLogic) AutoDispatch(req *types.AutoDispatchRequest) (resp *types.BaseResponse, err error) {
	// 权限检查
	employeeID, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeID == "" {
		return utils.Response.UnauthorizedError(), nil
	}

	// 获取任务信息
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, req.TaskID)
	if err != nil {
		l.Logger.Errorf("获取任务信息失败: %v", err)
		return utils.Response.BusinessError("task_not_found"), nil
	}

	// 检查用户权限
	if !l.hasDispatchPermission(employeeID, taskInfo) {
		return utils.Response.BusinessError("auto_dispatch_denied"), nil
	}

	// 检查GLM服务是否可用
	if l.svcCtx.GLMService == nil {
		return utils.Response.BusinessError("ai_service_unavailable"), nil
	}

	var recommendations []svc.DispatchRecommendation

	// 如果指定了节点ID，只处理该节点
	if req.NodeID != "" {
		taskNode, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, req.NodeID)
		if err != nil {
			l.Logger.Errorf("获取任务节点失败: %v", err)
			return utils.Response.BusinessError("task_node_not_found"), nil
		}

		// 验证节点属于该任务
		if taskNode.TaskId != req.TaskID {
			return utils.Response.BusinessError("node_not_belong_to_task"), nil
		}

		// 获取候选员工
		candidates, err := l.getCandidateEmployees(taskNode)
		if err != nil || len(candidates) == 0 {
			l.Logger.Infof("节点 %s 没有候选员工", taskNode.NodeName)
			return utils.Response.BusinessError("no_candidates"), nil
		}

		// 构建任务节点信息
		nodeInfo := svc.TaskNodeInfo{
			NodeID:       taskNode.TaskNodeId,
			NodeName:     taskNode.NodeName,
			NodeDetail:   taskNode.NodeDetail.String,
			Priority:     int(taskNode.NodePriority),
			Deadline:     taskNode.NodeDeadline.Format("2006-01-02"),
			RequiredDays: int(taskNode.EstimatedDays),
			TaskTitle:    taskInfo.TaskTitle,
		}

		// 调用GLM获取推荐
		recommendation, err := l.svcCtx.GLMService.GetDispatchRecommendation(l.ctx, nodeInfo, candidates)
		if err != nil {
			l.Logger.Errorf("GLM推荐失败: %v", err)
			// 使用默认推荐
			recommendation = l.getDefaultRecommendation(taskNode, candidates)
		}

		recommendations = append(recommendations, *recommendation)
	} else {
		// 未指定节点ID，获取任务的所有节点
		taskNodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, req.TaskID)
		if err != nil {
			l.Logger.Errorf("获取任务节点失败: %v", err)
			return utils.Response.BusinessError("task_nodes_fetch_failed"), nil
		}

		// 遍历需要派发的节点
		for _, taskNode := range taskNodes {
			// 跳过已有执行人的节点
			if taskNode.ExecutorId != "" {
				continue
			}

			// 获取候选员工
			candidates, err := l.getCandidateEmployees(taskNode)
			if err != nil || len(candidates) == 0 {
				l.Logger.Infof("节点 %s 没有候选员工", taskNode.NodeName)
				continue
			}

			// 构建任务节点信息
			nodeInfo := svc.TaskNodeInfo{
				NodeID:       taskNode.TaskNodeId,
				NodeName:     taskNode.NodeName,
				NodeDetail:   taskNode.NodeDetail.String,
				Priority:     int(taskNode.NodePriority),
				Deadline:     taskNode.NodeDeadline.Format("2006-01-02"),
				RequiredDays: int(taskNode.EstimatedDays),
				TaskTitle:    taskInfo.TaskTitle,
			}

			// 调用GLM获取推荐
			recommendation, err := l.svcCtx.GLMService.GetDispatchRecommendation(l.ctx, nodeInfo, candidates)
			if err != nil {
				l.Logger.Errorf("GLM推荐失败: %v", err)
				// 使用默认推荐
				recommendation = l.getDefaultRecommendation(taskNode, candidates)
			}

			recommendations = append(recommendations, *recommendation)
		}
	}

	result := AutoDispatchResult{
		TaskID:          req.TaskID,
		TaskTitle:       taskInfo.TaskTitle,
		Recommendations: recommendations,
		Message:         "AI已分析完成，请选择合适的执行人",
	}

	return utils.Response.Success(result), nil
}

// getCandidateEmployees 获取候选员工列表
func (l *AutoDispatchLogic) getCandidateEmployees(taskNode *task.TaskNode) ([]svc.EmployeeCandidate, error) {
	var candidates []svc.EmployeeCandidate

	// 根据部门获取员工
	employees, err := l.svcCtx.EmployeeModel.FindByDepartmentID(l.ctx, taskNode.DepartmentId)
	if err != nil {
		return nil, err
	}

	for _, emp := range employees {
		// 排除非在职员工
		if emp.Status != 1 {
			continue
		}

		// 获取员工统计信息
		activeTasks := l.getActiveTaskCount(emp.Id)
		completedTasks := l.getCompletedTaskCount(emp.Id)
		tenureMonths := l.calculateTenureMonths(emp)

		// 解析技能
		var skills []string
		if emp.Skills.Valid && emp.Skills.String != "" {
			skills = strings.Split(emp.Skills.String, ",")
		}

		// 获取职位名称
		positionName := ""
		if emp.PositionId.Valid {
			pos, err := l.svcCtx.PositionModel.FindOne(l.ctx, emp.PositionId.String)
			if err == nil {
				positionName = pos.PositionName
			}
		}

		// 获取部门名称
		deptName := ""
		if emp.DepartmentId.Valid {
			dept, err := l.svcCtx.DepartmentModel.FindOne(l.ctx, emp.DepartmentId.String)
			if err == nil {
				deptName = dept.DepartmentName
			}
		}

		candidates = append(candidates, svc.EmployeeCandidate{
			EmployeeID:     emp.Id,
			Name:           emp.RealName,
			Department:     deptName,
			Position:       positionName,
			Skills:         skills,
			TenureMonths:   tenureMonths,
			ActiveTasks:    activeTasks,
			CompletedTasks: completedTasks,
			AvgCompletion:  0.85, // TODO: 计算实际完成率
		})
	}

	return candidates, nil
}

// getActiveTaskCount 获取活跃任务数
func (l *AutoDispatchLogic) getActiveTaskCount(employeeID string) int {
	nodes, _, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, employeeID, 1, 1000)
	if err != nil {
		return 0
	}
	count := 0
	for _, n := range nodes {
		if n.NodeStatus == 1 { // 进行中
			count++
		}
	}
	return count
}

// getCompletedTaskCount 获取已完成任务数
func (l *AutoDispatchLogic) getCompletedTaskCount(employeeID string) int {
	nodes, _, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, employeeID, 1, 1000)
	if err != nil {
		return 0
	}
	count := 0
	for _, n := range nodes {
		if n.NodeStatus == 2 { // 已完成
			count++
		}
	}
	return count
}

// calculateTenureMonths 计算任职月数
func (l *AutoDispatchLogic) calculateTenureMonths(emp *user.Employee) int {
	if !emp.HireDate.Valid {
		return 0
	}
	months := int(time.Since(emp.HireDate.Time).Hours() / 24 / 30)
	return months
}

// getDefaultRecommendation 获取默认推荐
func (l *AutoDispatchLogic) getDefaultRecommendation(taskNode *task.TaskNode, candidates []svc.EmployeeCandidate) *svc.DispatchRecommendation {
	recommendation := &svc.DispatchRecommendation{
		TaskNodeID:   taskNode.TaskNodeId,
		TaskNodeName: taskNode.NodeName,
		AIAnalysis:   "基于工作负载的默认推荐",
	}

	// 按活跃任务数排序
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].ActiveTasks < candidates[i].ActiveTasks {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	for i, c := range candidates {
		if i >= 5 {
			break
		}
		recommendation.Candidates = append(recommendation.Candidates, svc.RecommendedEmployee{
			EmployeeID: c.EmployeeID,
			Name:       c.Name,
			Score:      float64(100 - c.ActiveTasks*10),
			Reason:     "当前工作负载较低",
			Rank:       i + 1,
		})
	}

	return recommendation
}

// hasDispatchPermission 检查派发权限
func (l *AutoDispatchLogic) hasDispatchPermission(employeeID string, taskInfo *task.Task) bool {
	// 创建者可派发
	if taskInfo.TaskCreator == employeeID {
		return true
	}

	// 任务负责人可派发
	if taskInfo.LeaderId.Valid && taskInfo.LeaderId.String == employeeID {
		return true
	}

	// 获取员工信息
	employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, employeeID)
	if err != nil {
		return false
	}

	// 创始人可派发
	if l.isFounder(employee) {
		return true
	}

	return false
}

// isFounder 检查是否是创始人
func (l *AutoDispatchLogic) isFounder(employee *user.Employee) bool {
	if employee.PositionId.Valid {
		pos, err := l.svcCtx.PositionModel.FindOne(l.ctx, employee.PositionId.String)
		if err == nil && pos != nil && pos.PositionCode.Valid {
			if pos.PositionCode.String == "FOUNDER" {
				return true
			}
		}
	}
	company, err := l.svcCtx.CompanyModel.FindOne(l.ctx, employee.CompanyId)
	if err == nil && company != nil && company.Owner == employee.UserId {
		return true
	}
	return false
}
