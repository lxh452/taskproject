package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAiSuggestionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAiSuggestionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAiSuggestionLogic {
	return &GetAiSuggestionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// TaskSummary 任务摘要
type TaskSummary struct {
	TaskNodeID   string `json:"taskNodeId"`
	TaskNodeName string `json:"taskNodeName"`
	TaskTitle    string `json:"taskTitle"`
	Priority     int    `json:"priority"`
	Progress     int    `json:"progress"`
	Deadline     string `json:"deadline"`
	DaysLeft     int    `json:"daysLeft"`
	Status       int    `json:"status"`
	IsLeader     bool   `json:"isLeader"`
	IsExecutor   bool   `json:"isExecutor"`
}

// NotificationSummary 通知摘要
type NotificationSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	CreatedAt string `json:"createdAt"`
	IsRead    bool   `json:"isRead"`
}

// ApprovalSummary 审批摘要
type ApprovalSummary struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	RequestedBy string `json:"requestedBy"`
	CreatedAt   string `json:"createdAt"`
}

// AISuggestionResult AI建议结果
type AISuggestionResult struct {
	Greeting         string                `json:"greeting"`
	TodayFocus       []string              `json:"todayFocus"`
	PriorityTasks    []TaskSummary         `json:"priorityTasks"`
	TimeAllocation   []TimeBlock           `json:"timeAllocation"`
	PendingApprovals []ApprovalSummary     `json:"pendingApprovals"`
	UnreadNotices    []NotificationSummary `json:"unreadNotices"`
	AIAnalysis       string                `json:"aiAnalysis"`
	Suggestions      []string              `json:"suggestions"`
}

// TimeBlock 时间块建议
type TimeBlock struct {
	TimeRange  string `json:"timeRange"`
	TaskName   string `json:"taskName"`
	TaskNodeID string `json:"taskNodeId"`
	Priority   string `json:"priority"`
	Reason     string `json:"reason"`
}

func (l *GetAiSuggestionLogic) GetAiSuggestion(req *types.GetAiSuggestionRequest) (resp *types.BaseResponse, err error) {
	// 获取当前员工ID
	employeeID, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeID == "" {
		return utils.Response.UnauthorizedError(), nil
	}

	// 收集用户数据
	tasks, err := l.getUserTasks(employeeID)
	if err != nil {
		l.Logger.Errorf("获取用户任务失败: %v", err)
	}

	notifications, err := l.getUserNotifications(employeeID)
	if err != nil {
		l.Logger.Errorf("获取用户通知失败: %v", err)
	}

	approvals, err := l.getPendingApprovals(employeeID)
	if err != nil {
		l.Logger.Errorf("获取待审批失败: %v", err)
	}

	// 如果GLM服务可用，使用AI生成建议
	if l.svcCtx.GLMService != nil {
		// 创建独立的context，设置30秒超时（prompt较长需要更多时间）
		aiCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 使用AI生成建议
		result, err := l.generateAISuggestionWithCtx(aiCtx, tasks, notifications, approvals)
		if err != nil {
			l.Logger.Errorf("AI生成建议失败: %v, 使用默认建议", err)
			result = l.generateDefaultSuggestion(tasks, notifications, approvals)
		}
		return utils.Response.Success(result), nil
	}

	// 无AI服务，使用默认建议
	result := l.generateDefaultSuggestion(tasks, notifications, approvals)
	return utils.Response.Success(result), nil
}

// getUserTasks 获取用户任务
func (l *GetAiSuggestionLogic) getUserTasks(employeeID string) ([]TaskSummary, error) {
	var tasks []TaskSummary

	// 获取作为执行人的任务节点
	executorNodes, _, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, employeeID, 1, 100)
	if err == nil {
		for _, node := range executorNodes {
			if node.NodeStatus == 2 { // 跳过已完成
				continue
			}
			task, _ := l.svcCtx.TaskModel.FindOne(l.ctx, node.TaskId)
			taskTitle := ""
			if task != nil {
				taskTitle = task.TaskTitle
			}
			daysLeft := int(time.Until(node.NodeDeadline).Hours() / 24)
			tasks = append(tasks, TaskSummary{
				TaskNodeID:   node.TaskNodeId,
				TaskNodeName: node.NodeName,
				TaskTitle:    taskTitle,
				Priority:     int(node.NodePriority),
				Progress:     int(node.Progress),
				Deadline:     node.NodeDeadline.Format("2006-01-02"),
				DaysLeft:     daysLeft,
				Status:       int(node.NodeStatus),
				IsExecutor:   true,
			})
		}
	}

	// 获取作为负责人的任务节点
	leaderNodes, _, err := l.svcCtx.TaskNodeModel.FindByLeader(l.ctx, employeeID, 1, 100)
	if err == nil {
		for _, node := range leaderNodes {
			if node.NodeStatus == 2 { // 跳过已完成
				continue
			}
			// 检查是否已添加
			exists := false
			for _, t := range tasks {
				if t.TaskNodeID == node.TaskNodeId {
					exists = true
					break
				}
			}
			if exists {
				continue
			}
			task, _ := l.svcCtx.TaskModel.FindOne(l.ctx, node.TaskId)
			taskTitle := ""
			if task != nil {
				taskTitle = task.TaskTitle
			}
			daysLeft := int(time.Until(node.NodeDeadline).Hours() / 24)
			tasks = append(tasks, TaskSummary{
				TaskNodeID:   node.TaskNodeId,
				TaskNodeName: node.NodeName,
				TaskTitle:    taskTitle,
				Priority:     int(node.NodePriority),
				Progress:     int(node.Progress),
				Deadline:     node.NodeDeadline.Format("2006-01-02"),
				DaysLeft:     daysLeft,
				Status:       int(node.NodeStatus),
				IsLeader:     true,
			})
		}
	}

	// 按优先级和截止日期排序
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Priority != tasks[j].Priority {
			return tasks[i].Priority < tasks[j].Priority
		}
		return tasks[i].DaysLeft < tasks[j].DaysLeft
	})

	return tasks, nil
}

// getUserNotifications 获取用户通知
func (l *GetAiSuggestionLogic) getUserNotifications(employeeID string) ([]NotificationSummary, error) {
	var notifications []NotificationSummary

	// 使用 FindByEmployee 方法，传入 isRead=0 只获取未读通知
	isRead := 0
	notifs, _, err := l.svcCtx.NotificationModel.FindByEmployee(l.ctx, employeeID, &isRead, nil, 1, 10)
	if err != nil {
		return notifications, err
	}

	for _, n := range notifs {
		category := ""
		if n.Category.Valid {
			category = n.Category.String
		}
		notifications = append(notifications, NotificationSummary{
			ID:        n.Id,
			Title:     n.Title,
			Type:      category,
			CreatedAt: n.CreateTime.Format("2006-01-02 15:04"),
			IsRead:    n.IsRead == 1,
		})
	}

	return notifications, nil
}

// getPendingApprovals 获取待审批
func (l *GetAiSuggestionLogic) getPendingApprovals(employeeID string) ([]ApprovalSummary, error) {
	var approvals []ApprovalSummary

	// 获取任务节点完成审批（使用 FindTaskNodeApprovalsByApprover）
	nodeApprovals, _, err := l.svcCtx.HandoverApprovalModel.FindTaskNodeApprovalsByApprover(l.ctx, employeeID, 1, 10)
	if err == nil {
		for _, h := range nodeApprovals {
			if h.ApprovalType != 0 { // 只获取待审批 (0-待审批)
				continue
			}
			approvals = append(approvals, ApprovalSummary{
				ID:        h.ApprovalId,
				Type:      "node_completion",
				Title:     "任务节点完成审批",
				CreatedAt: h.CreateTime.Format("2006-01-02 15:04"),
			})
		}
	}

	// 获取交接审批（查询作为审批人的交接记录）
	handovers, _, err := l.svcCtx.TaskHandoverModel.FindByEmployeeInvolved(l.ctx, employeeID, 1, 10)
	if err == nil {
		for _, h := range handovers {
			// 只获取待审批状态(1)且当前员工是审批人的记录
			if h.HandoverStatus != 1 || !h.ApproverId.Valid || h.ApproverId.String != employeeID {
				continue
			}
			approvals = append(approvals, ApprovalSummary{
				ID:        h.HandoverId,
				Type:      "handover",
				Title:     "任务交接审批",
				CreatedAt: h.CreateTime.Format("2006-01-02 15:04"),
			})
		}
	}

	return approvals, nil
}

// generateAISuggestionWithCtx 使用AI生成建议（带独立context）
func (l *GetAiSuggestionLogic) generateAISuggestionWithCtx(ctx context.Context, tasks []TaskSummary, notifications []NotificationSummary, approvals []ApprovalSummary) (*AISuggestionResult, error) {
	// 构建提示词
	prompt := l.buildSuggestionPrompt(tasks, notifications, approvals)

	// 打印发送给GLM的prompt用于调试
	l.Logger.Infof("发送给GLM的prompt: 任务数=%d, 通知数=%d, 审批数=%d", len(tasks), len(notifications), len(approvals))
	l.Logger.Infof("Prompt内容: %s", prompt[:min(500, len(prompt))])

	// 调用GLM（使用传入的context）
	response, err := l.svcCtx.GLMService.CallGLMWithPrompt(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 打印GLM原始响应用于调试
	l.Logger.Infof("GLM原始响应: %s", response)

	// 解析响应
	result, err := l.parseAIResponse(response, tasks, notifications, approvals)
	if err != nil {
		l.Logger.Errorf("解析GLM响应失败: %v", err)
		return nil, err
	}

	l.Logger.Infof("AI建议生成成功, greeting: %s", result.Greeting)
	return result, nil
}

// generateAISuggestion 使用AI生成建议
func (l *GetAiSuggestionLogic) generateAISuggestion(tasks []TaskSummary, notifications []NotificationSummary, approvals []ApprovalSummary) (*AISuggestionResult, error) {
	return l.generateAISuggestionWithCtx(l.ctx, tasks, notifications, approvals)
}

// buildSuggestionPrompt 构建建议提示词
func (l *GetAiSuggestionLogic) buildSuggestionPrompt(tasks []TaskSummary, notifications []NotificationSummary, approvals []ApprovalSummary) string {
	now := time.Now()
	hour := now.Hour()

	var greeting string
	if hour < 12 {
		greeting = "早上好"
	} else if hour < 18 {
		greeting = "下午好"
	} else {
		greeting = "晚上好"
	}

	// 构建详细任务信息（包含ID用于时间分配关联）
	var taskInfo strings.Builder
	var taskList []map[string]interface{}
	for i, t := range tasks {
		if i >= 10 { // 最多10个任务
			break
		}
		priorityText := []string{"紧急", "高", "中", "低"}[min(t.Priority-1, 3)]

		// 计算任务紧迫度评分（用于AI参考）
		urgencyScore := 100 - t.Progress // 进度越低越需要关注
		if t.DaysLeft <= 0 {
			urgencyScore += 50 // 已逾期
		} else if t.DaysLeft <= 1 {
			urgencyScore += 40 // 明天截止
		} else if t.DaysLeft <= 3 {
			urgencyScore += 20 // 3天内截止
		}
		if t.Priority == 1 {
			urgencyScore += 30 // 紧急
		} else if t.Priority == 2 {
			urgencyScore += 15 // 高优先级
		}

		taskInfo.WriteString(fmt.Sprintf("- [ID:%s] %s (所属任务:%s, 优先级:%s, 进度:%d%%, 剩余%d天, 紧迫度:%d)\n",
			t.TaskNodeID, t.TaskNodeName, t.TaskTitle, priorityText, t.Progress, t.DaysLeft, urgencyScore))

		taskList = append(taskList, map[string]interface{}{
			"id":           t.TaskNodeID,
			"name":         t.TaskNodeName,
			"taskTitle":    t.TaskTitle,
			"priority":     priorityText,
			"progress":     t.Progress,
			"daysLeft":     t.DaysLeft,
			"urgencyScore": urgencyScore,
		})
	}

	// 计算今日可用工作时间
	workStartHour := 9
	if hour > 9 {
		workStartHour = hour
	}
	workEndHour := 18
	availableHours := workEndHour - workStartHour
	if availableHours < 0 {
		availableHours = 0
	}

	prompt := fmt.Sprintf(`你是智能工作助手，根据用户实际任务数据提供今日工作建议。

## 当前: %s %s，剩余工作时间约%d小时（%d:00-%d:00）

## 待办任务 (%d个)
%s
## 未读通知: %d条 | 待审批: %d个

## 输出要求（纯JSON，无markdown标记）
{
  "greeting": "个性化问候",
  "todayFocus": ["重点1", "重点2", "重点3"],
  "timeAllocation": [
    {"timeRange": "HH:MM-HH:MM", "taskName": "任务名", "taskNodeId": "任务ID", "priority": "高/中/低", "reason": "安排理由"}
  ],
  "suggestions": ["建议1", "建议2"],
  "aiAnalysis": "工作分析(50字内)"
}

规则：紧迫度高的优先安排；已逾期必须优先；timeAllocation只放实际任务，必须有taskNodeId`,
		now.Format("2006-01-02"), greeting, availableHours, workStartHour, workEndHour,
		len(tasks), taskInfo.String(),
		len(notifications),
		len(approvals))

	return prompt
}

// parseAIResponse 解析AI响应
func (l *GetAiSuggestionLogic) parseAIResponse(response string, tasks []TaskSummary, notifications []NotificationSummary, approvals []ApprovalSummary) (*AISuggestionResult, error) {
	result := &AISuggestionResult{
		PriorityTasks:    tasks,
		UnreadNotices:    notifications,
		PendingApprovals: approvals,
	}

	// 尝试解析JSON
	type AITimeBlock struct {
		TimeRange  string `json:"timeRange"`
		TaskName   string `json:"taskName"`
		TaskNodeID string `json:"taskNodeId"`
		Priority   string `json:"priority"`
		Reason     string `json:"reason"`
	}

	type AIOutput struct {
		Greeting       string        `json:"greeting"`
		TodayFocus     []string      `json:"todayFocus"`
		TimeAllocation []AITimeBlock `json:"timeAllocation"`
		Suggestions    []string      `json:"suggestions"`
		AIAnalysis     string        `json:"aiAnalysis"`
	}

	var output AIOutput

	// 移除 markdown 代码块标记
	cleanResponse := response
	cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
	cleanResponse = strings.TrimPrefix(cleanResponse, "```")
	cleanResponse = strings.TrimSuffix(cleanResponse, "```")
	cleanResponse = strings.TrimSpace(cleanResponse)

	// 提取JSON
	jsonStart := strings.Index(cleanResponse, "{")
	jsonEnd := strings.LastIndex(cleanResponse, "}")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		jsonStr := cleanResponse[jsonStart : jsonEnd+1]
		if err := json.Unmarshal([]byte(jsonStr), &output); err == nil {
			result.Greeting = output.Greeting
			result.TodayFocus = output.TodayFocus
			result.Suggestions = output.Suggestions
			result.AIAnalysis = output.AIAnalysis

			// 转换时间分配，保留taskNodeId
			for _, tb := range output.TimeAllocation {
				result.TimeAllocation = append(result.TimeAllocation, TimeBlock{
					TimeRange:  tb.TimeRange,
					TaskName:   tb.TaskName,
					TaskNodeID: tb.TaskNodeID,
					Priority:   tb.Priority,
					Reason:     tb.Reason,
				})
			}
		} else {
			l.Logger.Errorf("JSON解析失败: %v, jsonStr: %s", err, jsonStr[:min(200, len(jsonStr))])
		}
	}

	// 如果解析失败，使用默认值
	if result.Greeting == "" {
		result.Greeting = "今天也要加油哦！"
	}
	if len(result.TodayFocus) == 0 {
		result.TodayFocus = []string{"完成紧急任务", "处理待审批事项", "查看未读通知"}
	}

	return result, nil
}

// generateDefaultSuggestion 生成默认建议
func (l *GetAiSuggestionLogic) generateDefaultSuggestion(tasks []TaskSummary, notifications []NotificationSummary, approvals []ApprovalSummary) *AISuggestionResult {
	now := time.Now()
	hour := now.Hour()

	var greeting string
	if hour < 12 {
		greeting = "早上好！新的一天，高效工作从现在开始。"
	} else if hour < 18 {
		greeting = "下午好！继续保持专注，完成今日目标。"
	} else {
		greeting = "晚上好！辛苦了，注意劳逸结合。"
	}

	// 生成今日重点
	var todayFocus []string
	urgentCount := 0
	overdueCount := 0
	for _, t := range tasks {
		if t.Priority == 1 {
			urgentCount++
		}
		if t.DaysLeft < 0 {
			overdueCount++
		}
	}

	if urgentCount > 0 {
		todayFocus = append(todayFocus, fmt.Sprintf("处理 %d 个紧急任务", urgentCount))
	}
	if overdueCount > 0 {
		todayFocus = append(todayFocus, fmt.Sprintf("关注 %d 个已逾期任务", overdueCount))
	}
	if len(approvals) > 0 {
		todayFocus = append(todayFocus, fmt.Sprintf("处理 %d 个待审批事项", len(approvals)))
	}
	if len(notifications) > 0 {
		todayFocus = append(todayFocus, fmt.Sprintf("查看 %d 条未读通知", len(notifications)))
	}
	if len(todayFocus) == 0 {
		todayFocus = []string{"保持工作节奏", "及时跟进任务进度"}
	}

	// 生成时间分配建议
	var timeAllocation []TimeBlock
	currentHour := hour
	if currentHour < 9 {
		currentHour = 9
	}

	for i, t := range tasks {
		if i >= 4 || currentHour >= 18 {
			break
		}
		priorityText := []string{"紧急", "高", "中", "低"}[min(t.Priority-1, 3)]
		var reason string
		if t.DaysLeft <= 0 {
			reason = "已逾期，需立即处理"
		} else if t.DaysLeft <= 2 {
			reason = "即将到期，优先处理"
		} else if t.Priority == 1 {
			reason = "紧急任务，优先安排"
		} else {
			reason = "按计划推进"
		}

		timeAllocation = append(timeAllocation, TimeBlock{
			TimeRange:  fmt.Sprintf("%02d:00-%02d:30", currentHour, currentHour+1),
			TaskName:   t.TaskNodeName,
			TaskNodeID: t.TaskNodeID,
			Priority:   priorityText,
			Reason:     reason,
		})
		currentHour += 2
	}

	// 生成建议
	var suggestions []string
	if urgentCount > 0 {
		suggestions = append(suggestions, "建议优先处理紧急任务，避免影响项目进度")
	}
	if overdueCount > 0 {
		suggestions = append(suggestions, "有逾期任务需要关注，建议与相关人员沟通调整计划")
	}
	if len(approvals) > 0 {
		suggestions = append(suggestions, "及时处理审批事项，避免阻塞他人工作")
	}
	if len(suggestions) == 0 {
		suggestions = []string{"保持良好的工作节奏", "定期检查任务进度", "及时与团队沟通"}
	}

	// 生成分析
	var analysis string
	if len(tasks) == 0 {
		analysis = "当前没有待办任务，可以主动了解团队需求或学习提升。"
	} else if urgentCount > 2 {
		analysis = fmt.Sprintf("当前有%d个紧急任务，建议集中精力优先处理，必要时寻求支援。", urgentCount)
	} else {
		analysis = fmt.Sprintf("当前有%d个待办任务，工作量适中，按优先级有序推进即可。", len(tasks))
	}

	// 限制任务数量
	priorityTasks := tasks
	if len(priorityTasks) > 5 {
		priorityTasks = priorityTasks[:5]
	}

	return &AISuggestionResult{
		Greeting:         greeting,
		TodayFocus:       todayFocus,
		PriorityTasks:    priorityTasks,
		TimeAllocation:   timeAllocation,
		PendingApprovals: approvals,
		UnreadNotices:    notifications,
		AIAnalysis:       analysis,
		Suggestions:      suggestions,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
