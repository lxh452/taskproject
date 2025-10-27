package svc

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"task_Project/model/task"
	"task_Project/task/internal/middleware"

	"github.com/zeromicro/go-zero/core/logx"
)

// DispatchService 任务派发服务
type DispatchService struct {
	svcCtx *ServiceContext
}

// NewDispatchService 创建派发服务
func NewDispatchService(svcCtx *ServiceContext) *DispatchService {
	return &DispatchService{
		svcCtx: svcCtx,
	}
}

// EmployeeScore 员工评分
type EmployeeScore struct {
	EmployeeID string  `json:"employeeId"`
	Score      float64 `json:"score"`
	Reason     string  `json:"reason"`
}

// DispatchConfig 派发配置
type DispatchConfig struct {
	// 权重配置
	LastMonthTaskRatio      float64 // 上个月完成任务数量占比 0.1
	PriorityTaskRate        float64 // 优先级任务节点完成率 0.2
	TenureWeight            float64 // 任职时长 0.2
	TotalTaskCount          float64 // 总任务完成数量 0.1
	AutoAddPersonnelCount   float64 // 任务节点增派人手次数 0.2
	ManualAddPersonnelCount float64 // 手动需要增派人手次数 0.2
}

// GetDefaultDispatchConfig 获取默认派发配置
func GetDefaultDispatchConfig() *DispatchConfig {
	return &DispatchConfig{
		LastMonthTaskRatio:      0.1,
		PriorityTaskRate:        0.2,
		TenureWeight:            0.2,
		TotalTaskCount:          0.1,
		AutoAddPersonnelCount:   0.2,
		ManualAddPersonnelCount: 0.2,
	}
}

// CalculateEmployeeScore 计算员工评分
func (d *DispatchService) CalculateEmployeeScore(ctx context.Context, employeeID string, taskNode *task.TaskNode) (*EmployeeScore, error) {
	config := GetDefaultDispatchConfig()

	// 获取员工信息
	employee, err := d.svcCtx.EmployeeModel.FindOne(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	// 获取员工上个月完成的任务数量
	lastMonthTasks, err := d.getLastMonthCompletedTasks(ctx, employeeID)
	if err != nil {
		logx.Errorf("获取员工上个月完成任务失败: %v", err)
		lastMonthTasks = 0
	}

	// 获取员工总完成任务数量
	totalTasks, err := d.getTotalCompletedTasks(ctx, employeeID)
	if err != nil {
		logx.Errorf("获取员工总完成任务失败: %v", err)
		totalTasks = 0
	}

	// 获取员工优先级任务完成率
	priorityRate, err := d.getPriorityTaskCompletionRate(ctx, employeeID)
	if err != nil {
		logx.Errorf("获取员工优先级任务完成率失败: %v", err)
		priorityRate = 0.5 // 默认50%
	}

	// 计算任职时长（月）
	var hireTime *time.Time
	if employee.HireDate.Valid {
		hireTime = &employee.HireDate.Time
	}
	tenureMonths := d.calculateTenureMonths(hireTime)

	// 获取任务节点增派人手次数
	autoAddCount, err := d.getTaskNodeAddPersonnelCount(ctx, taskNode.TaskNodeId, true)
	if err != nil {
		logx.Errorf("获取任务节点自动增派人手次数失败: %v", err)
		autoAddCount = 0
	}

	manualAddCount, err := d.getTaskNodeAddPersonnelCount(ctx, taskNode.TaskNodeId, false)
	if err != nil {
		logx.Errorf("获取任务节点手动增派人手次数失败: %v", err)
		manualAddCount = 0
	}

	// 计算各项得分
	lastMonthScore := d.calculateLastMonthTaskScore(lastMonthTasks) * config.LastMonthTaskRatio
	priorityScore := priorityRate * config.PriorityTaskRate
	tenureScore := d.calculateTenureScore(tenureMonths) * config.TenureWeight
	totalTaskScore := d.calculateTotalTaskScore(totalTasks) * config.TotalTaskCount
	autoAddScore := d.calculateAddPersonnelScore(autoAddCount, true) * config.AutoAddPersonnelCount
	manualAddScore := d.calculateAddPersonnelScore(manualAddCount, false) * config.ManualAddPersonnelCount

	// 计算总分
	totalScore := lastMonthScore + priorityScore + tenureScore + totalTaskScore + autoAddScore + manualAddScore

	// 生成评分说明
	reason := d.generateScoreReason(lastMonthTasks, priorityRate, tenureMonths, totalTasks, autoAddCount, manualAddCount)

	return &EmployeeScore{
		EmployeeID: employeeID,
		Score:      totalScore,
		Reason:     reason,
	}, nil
}

// SelectBestEmployee 选择最佳员工
func (d *DispatchService) SelectBestEmployee(ctx context.Context, candidateIDs []string, taskNode *task.TaskNode) (*EmployeeScore, error) {
	var scores []*EmployeeScore

	for _, employeeID := range candidateIDs {
		score, err := d.CalculateEmployeeScore(ctx, employeeID, taskNode)
		if err != nil {
			logx.Errorf("计算员工 %s 评分失败: %v", employeeID, err)
			continue
		}
		scores = append(scores, score)
	}

	if len(scores) == 0 {
		return nil, nil
	}

	// 按评分排序，选择最高分
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	return scores[0], nil
}

// AutoDispatchTask 自动派发任务
func (d *DispatchService) AutoDispatchTask(ctx context.Context, taskNodeID string) error {
	// 获取任务节点信息
	taskNode, err := d.svcCtx.TaskNodeModel.FindOne(ctx, taskNodeID)
	if err != nil {
		return err
	}

	// 获取候选员工列表
	candidates, err := d.getCandidateEmployees(ctx, taskNode)
	if err != nil {
		return err
	}

	if len(candidates) == 0 {
		return nil // 没有候选员工
	}

	// 选择最佳员工
	bestEmployee, err := d.SelectBestEmployee(ctx, candidates, taskNode)
	if err != nil {
		return err
	}

	if bestEmployee == nil {
		return nil // 没有合适的员工
	}

	// 更新任务节点执行人
	err = d.svcCtx.TaskNodeModel.UpdateExecutor(ctx, taskNodeID, bestEmployee.EmployeeID)
	if err != nil {
		return err
	}

	// 发送派发通知邮件
	go d.sendDispatchNotification(ctx, taskNode, bestEmployee.EmployeeID, bestEmployee.Reason)

	return nil
}

// 获取员工上个月完成的任务数量
func (d *DispatchService) getLastMonthCompletedTasks(ctx context.Context, employeeID string) (int, error) {
	// 计算上个月的时间范围
	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	startTime := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, lastMonth.Location())
	_ = startTime.AddDate(0, 1, 0).Add(-time.Second) // endTime

	// 查询上个月完成的任务节点数量
	// 这里需要根据实际的TaskLogModel实现来查询
	// 假设有FindCompletedTasksByEmployeeAndTimeRange方法
	return 0, nil // TODO: 实现具体查询逻辑
}

// 获取员工总完成任务数量
func (d *DispatchService) getTotalCompletedTasks(ctx context.Context, employeeID string) (int, error) {
	// 查询员工总完成的任务节点数量
	// 假设有FindCompletedTasksByEmployee方法
	return 0, nil // TODO: 实现具体查询逻辑
}

// 获取员工优先级任务完成率
func (d *DispatchService) getPriorityTaskCompletionRate(ctx context.Context, employeeID string) (float64, error) {
	// 查询员工优先级任务的完成情况
	// 计算完成率 = 按时完成数 / 总任务数
	return 0.8, nil // TODO: 实现具体查询逻辑
}

// 计算任职时长（月）
func (d *DispatchService) calculateTenureMonths(hireDate *time.Time) int {
	if hireDate == nil {
		return 0
	}

	now := time.Now()
	months := int(now.Sub(*hireDate).Hours() / 24 / 30)
	return months
}

// 获取任务节点增派人手次数
func (d *DispatchService) getTaskNodeAddPersonnelCount(ctx context.Context, taskNodeID string, isAuto bool) (int, error) {
	// 查询任务节点的增派人手记录
	// 假设有FindAddPersonnelCountByTaskNode方法
	return 0, nil // TODO: 实现具体查询逻辑
}

// 计算上个月任务得分
func (d *DispatchService) calculateLastMonthTaskScore(taskCount int) float64 {
	// 根据任务数量计算得分，任务越多得分越高
	// 这里使用对数函数来避免分数过高
	return math.Log(float64(taskCount + 1))
}

// 计算任职时长得分
func (d *DispatchService) calculateTenureScore(months int) float64 {
	// 任职时长越长得分越高，但有上限
	if months <= 0 {
		return 0
	}

	// 使用对数函数，12个月为满分
	return math.Min(math.Log(float64(months+1))/math.Log(13), 1.0)
}

// 计算总任务得分
func (d *DispatchService) calculateTotalTaskScore(taskCount int) float64 {
	// 根据总任务数量计算得分
	return math.Log(float64(taskCount + 1))
}

// 计算增派人手得分
func (d *DispatchService) calculateAddPersonnelScore(count int, isAuto bool) float64 {
	// 增派人手次数越少得分越高（表示员工能力强，不需要增派人手）
	if count == 0 {
		return 1.0
	}

	// 使用倒数函数，次数越多得分越低
	return 1.0 / (float64(count) + 1.0)
}

// 生成评分说明
func (d *DispatchService) generateScoreReason(lastMonthTasks int, priorityRate float64, tenureMonths int, totalTasks int, autoAddCount int, manualAddCount int) string {
	reason := "评分详情："

	reason += fmt.Sprintf("上个月完成任务数量: %d; ", lastMonthTasks)
	reason += fmt.Sprintf("优先级任务完成率: %.1f%%; ", priorityRate*100)
	reason += fmt.Sprintf("任职时长: %d个月; ", tenureMonths)
	reason += fmt.Sprintf("总完成任务数量: %d; ", totalTasks)
	reason += fmt.Sprintf("自动增派人手次数: %d; ", autoAddCount)
	reason += fmt.Sprintf("手动增派人手次数: %d", manualAddCount)

	return reason
}

// 获取候选员工列表
func (d *DispatchService) getCandidateEmployees(ctx context.Context, taskNode *task.TaskNode) ([]string, error) {
	// 根据任务节点的要求获取候选员工
	// 1. 根据部门筛选
	// 2. 根据技能要求筛选
	// 3. 根据角色标签筛选
	// 4. 排除已离职员工

	// 这里需要根据实际的业务逻辑实现
	// 假设有GetCandidatesByTaskNode方法
	return []string{}, nil // TODO: 实现具体查询逻辑
}

// 发送派发通知邮件
func (d *DispatchService) sendDispatchNotification(ctx context.Context, taskNode *task.TaskNode, employeeID string, reason string) {
	// 获取员工信息
	employee, err := d.svcCtx.EmployeeModel.FindOne(ctx, employeeID)
	if err != nil {
		logx.Errorf("获取员工信息失败: %v", err)
		return
	}

	// 获取任务信息
	taskInfo, err := d.svcCtx.TaskModel.FindOne(ctx, taskNode.TaskId)
	if err != nil {
		logx.Errorf("获取任务信息失败: %v", err)
		return
	}

	// 构建邮件内容
	subject := "任务派发通知"
	body := fmt.Sprintf(`
		<h2>任务派发通知</h2>
		<p>您好 %s，</p>
		<p>您已被分配执行以下任务节点：</p>
		<ul>
			<li><strong>任务名称：</strong>%s</li>
			<li><strong>节点名称：</strong>%s</li>
			<li><strong>节点详情：</strong>%s</li>
			<li><strong>截止时间：</strong>%s</li>
		</ul>
		<p><strong>派发原因：</strong>%s</p>
		<p>请及时查看并开始执行任务。</p>
		<p>祝工作顺利！</p>
	`, employee.RealName, taskInfo.TaskTitle, taskNode.NodeName, taskNode.NodeDetail.String,
		taskNode.NodeDeadline.Format("2006-01-02 15:04:05"), reason)

	// 发送邮件
	emailMsg := middleware.EmailMessage{
		To:      []string{employee.Email.String},
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	}

	if err := d.svcCtx.EmailMiddleware.SendEmail(ctx, emailMsg); err != nil {
		logx.Errorf("发送任务派发通知邮件失败: %v", err)
	}
}
