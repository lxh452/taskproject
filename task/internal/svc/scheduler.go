package svc

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"task_Project/model/task"
	"task_Project/model/user"
	"task_Project/model/user_auth"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

// SchedulerService 定时任务服务
type SchedulerService struct {
	svcCtx *ServiceContext
	stopCh chan struct{}
}

// NewSchedulerService 创建定时任务服务
func NewSchedulerService(svcCtx *ServiceContext) *SchedulerService {
	return &SchedulerService{
		svcCtx: svcCtx,
		stopCh: make(chan struct{}),
	}
}

// Start 兼容外部启动调用
func (s *SchedulerService) Start() {
	s.StartScheduler()
}

// Stop 预留优雅关闭（当前 ticker goroutine 无状态，可按需扩展）
func (s *SchedulerService) Stop() {
	select {
	case <-s.stopCh:
		// already closed
	default:
		close(s.stopCh)
	}
}

// StartScheduler 启动定时任务
func (s *SchedulerService) StartScheduler() {
	// 启动任务截止提醒定时任务
	go s.startDeadlineReminder()

	// 启动每日汇报提醒定时任务
	go s.startDailyReportReminder()

	// 启动进度缓慢检测定时任务
	go s.startSlowProgressDetection()

	// 启动员工离职检测定时任务
	go s.startEmployeeLeaveDetection()

	// 启动任务节点闲置检查定时任务
	go s.startTaskNodeIdleCheck()
}

// 任务截止提醒定时任务
func (s *SchedulerService) startDeadlineReminder() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.checkTaskDeadlines()
		}
	}
}

// 检查任务截止时间
func (s *SchedulerService) checkTaskDeadlines() {
	ctx := context.Background()

	// 获取即将截止的任务节点（24小时内）
	deadline := time.Now().Add(24 * time.Hour)
	taskNodes, err := s.svcCtx.TaskNodeModel.FindByDeadlineRange(ctx,
		time.Now().Format("2006-01-02 15:04:05"),
		deadline.Format("2006-01-02 15:04:05"))

	if err != nil {
		logx.Errorf("查询即将截止的任务节点失败: %v", err)
		return
	}

	// 发送截止提醒（通过消息队列，消费者会查询并发送）
	for _, taskNode := range taskNodes {
		if taskNode.NodeStatus == 1 { // 进行中（状态1）
			// 发布邮件事件（消费者会查询执行人并发送）
			if s.svcCtx.EmailMQService != nil {
				emailEvent := &EmailEvent{
					EventType: "task.deadline.reminder",
					NodeID:    taskNode.TaskNodeId,
				}
				if err := s.svcCtx.EmailMQService.PublishEmailEvent(ctx, emailEvent); err != nil {
					logx.Errorf("发布任务截止提醒邮件事件失败: %v", err)
				}
			}

			// 发布通知事件（消费者会查询执行人并创建通知）
			if s.svcCtx.NotificationMQService != nil {
				event := &NotificationEvent{
					EventType:   "task.deadline.reminder",
					NodeID:      taskNode.TaskNodeId,
					Type:        2,
					Category:    "task",
					Priority:    3,
					RelatedID:   taskNode.TaskId,
					RelatedType: "task",
				}
				if err := s.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
					logx.Errorf("发布通知事件失败: %v", err)
				}
			}
		}
	}
}

// 每日汇报提醒定时任务
func (s *SchedulerService) startDailyReportReminder() {
	// 每天下午5点30分提醒
	ticker := time.NewTicker(20 * time.Minute)

	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			now := time.Now()
			if now.Hour() == 17 && now.Minute() >= 00 && now.Minute() < 31 {
				s.sendDailyReportReminders()
			}
		}
	}
}

// 发送每日汇报提醒
func (s *SchedulerService) sendDailyReportReminders() {
	ctx := context.Background()

	// 获取所有在职员工
	employees, err := s.svcCtx.EmployeeModel.FindByStatus(ctx, 1) // 在职
	if err != nil {
		logx.Errorf("查询在职员工失败: %v", err)
		return
	}

	// 发送提醒（通过消息队列，消费者会查询并发送）
	for _, employee := range employees {
		// 发布邮件事件（消费者会查询员工并发送）
		if s.svcCtx.EmailMQService != nil {
			emailEvent := &EmailEvent{
				EventType:  "daily.report.reminder",
				EmployeeID: employee.Id,
			}
			if err := s.svcCtx.EmailMQService.PublishEmailEvent(ctx, emailEvent); err != nil {
				logx.Errorf("发布每日汇报提醒邮件事件失败: %v", err)
			}
		}
	}
}

// 进度缓慢检测定时任务
func (s *SchedulerService) startSlowProgressDetection() {
	ticker := time.NewTicker(2 * time.Hour) // 每2小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.checkSlowProgress()
		}
	}
}

// 检查进度缓慢的任务
func (s *SchedulerService) checkSlowProgress() {
	ctx := context.Background()

	// 获取进行中的任务节点
	taskNodes, err := s.svcCtx.TaskNodeModel.FindByStatus(ctx, 1) // 进行中
	if err != nil {
		logx.Errorf("查询进行中的任务节点失败: %v", err)
		return
	}

	// 检查进度缓慢的任务（通过消息队列）
	for _, taskNode := range taskNodes {
		// 计算任务开始时间
		startTime := taskNode.CreateTime

		// 计算预期进度
		deadline := taskNode.NodeDeadline
		totalDuration := deadline.Sub(startTime)
		elapsed := time.Since(startTime)

		if totalDuration > 0 {
			expectedProgress := float64(elapsed) / float64(totalDuration)
			if expectedProgress > 0.5 && taskNode.Progress < int64(expectedProgress*0.5) {
				// 进度缓慢，发送提醒（通过消息队列，消费者会查询并发送）
				if s.svcCtx.EmailMQService != nil {
					emailEvent := &EmailEvent{
						EventType: "task.slow.progress",
						NodeID:    taskNode.TaskNodeId,
					}
					if err := s.svcCtx.EmailMQService.PublishEmailEvent(ctx, emailEvent); err != nil {
						logx.Errorf("发布进度缓慢提醒邮件事件失败: %v", err)
					}
				}
			}
		}
	}
}

// 员工离职检测定时任务
func (s *SchedulerService) startEmployeeLeaveDetection() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.checkEmployeeLeave()
		}
	}
}

// 检查员工离职
func (s *SchedulerService) checkEmployeeLeave() {
	ctx := context.Background()

	// 获取离职员工
	employees, err := s.svcCtx.EmployeeModel.FindByStatus(ctx, 0) // 离职
	if err != nil {
		logx.Errorf("查询离职员工失败: %v", err)
		return
	}

	// 发送离职通知（通过消息队列）
	for _, employee := range employees {
		// 检查是否已经发送过通知
		// TODO: 实现检查逻辑
		// lastNotification, err := s.svcCtx.NotificationModel.FindByEmployeeAndType(ctx, employee.Id, "employee_leave")
		// if err == nil && lastNotification != nil {
		// 	// 已经发送过通知，跳过
		// 	continue
		// }

		// 获取员工当前负责的任务节点
		taskNodes := []string{} // TODO: 实现获取任务节点逻辑

		// 发送邮件（通过消息队列）
		recipientEmail := ""
		if employee.DepartmentId.Valid && employee.DepartmentId.String != "" {
			department, err := s.svcCtx.DepartmentModel.FindOne(ctx, employee.DepartmentId.String)
			if err == nil && department.ManagerId.Valid && department.ManagerId.String != "" {
				manager, err := s.svcCtx.EmployeeModel.FindOne(ctx, department.ManagerId.String)
				if err == nil && manager.Email.Valid && manager.Email.String != "" {
					recipientEmail = manager.Email.String
				}
			}
		}

		if recipientEmail == "" && employee.Email.Valid && employee.Email.String != "" {
			recipientEmail = employee.Email.String
		}

		if recipientEmail != "" && s.svcCtx.EmailService != nil {
			if err := s.svcCtx.EmailService.SendEmployeeLeaveEmail(ctx, recipientEmail, employee.RealName, taskNodes); err != nil {
				logx.Errorf("发送离职邮件失败: %v", err)
			}
		}

		// 创建系统通知（通过消息队列，消费者会查询并创建）
		if s.svcCtx.NotificationMQService != nil {
			event := &NotificationEvent{
				EventType:   "employee.leave",
				EmployeeIDs: []string{employee.Id}, // 离职员工ID，消费者会使用
				Type:        3,                     // 离职通知
				Category:    "employee",
				Priority:    3, // 高优先级
				RelatedID:   employee.Id,
				RelatedType: "employee",
			}
			if err := s.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
				logx.Errorf("发布通知事件失败: %v", err)
			}
		}
	}
}

// 自动派发任务
func (s *SchedulerService) AutoDispatchTasks() {
	ctx := context.Background()

	// 获取待派发的任务节点
	taskNodes, err := s.svcCtx.TaskNodeModel.FindByStatus(ctx, 1) // 待开始
	if err != nil {
		logx.Errorf("查询待派发的任务节点失败: %v", err)
		return
	}

	// 自动派发任务
	dispatchService := NewDispatchService(s.svcCtx)
	for _, taskNode := range taskNodes {
		if err := dispatchService.AutoDispatchTask(ctx, taskNode.TaskNodeId); err != nil {
			logx.Errorf("自动派发任务失败: %v", err)
		}
	}
}

// 发送系统通知
func (s *SchedulerService) SendSystemNotification(ctx context.Context, employeeID, title, content string) error {
	notification := &user_auth.Notification{
		Id:         utils.NewCommon().GenerateID(),
		EmployeeId: employeeID,
		Title:      title,
		Content:    content,
		Type:       0, // 系统通知
		Priority:   1, // 低优先级
		IsRead:     0, // 未读
		SenderId:   sql.NullString{String: "system", Valid: true},
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}

	_, err := s.svcCtx.NotificationModel.Insert(ctx, notification)
	return err
}

// SendEmailNotification 发送邮件通知（通过消息队列）
func (s *SchedulerService) SendEmailNotification(ctx context.Context, to []string, subject, content string) error {
	// 通过消息队列发送邮件
	if s.svcCtx.EmailMQService != nil {
		emailEvent := &EmailEvent{
			EventType: "send",
			To:        to,
			Subject:   subject,
			Body:      content,
			IsHTML:    true,
		}
		if err := s.svcCtx.EmailMQService.PublishEmailEvent(ctx, emailEvent); err != nil {
			logx.Errorf("发布邮件事件失败: %v", err)
			return err
		}
		return nil // 消息已发布到队列，异步处理
	}

	// 消息队列未配置，记录警告
	logx.Errorf("EmailMQService not initialized, email not sent: to=%v, subject=%s", to, subject)
	return fmt.Errorf("email service not available")
}

// 发送短信通知
func (s *SchedulerService) SendSMSNotification(ctx context.Context, phone, content string) error {
	return s.svcCtx.SMSMiddleware.SendNotificationSMS(ctx, phone, content)
}

// 任务节点闲置检查定时任务
func (s *SchedulerService) startTaskNodeIdleCheck() {
	ticker := time.NewTicker(30 * time.Minute) // 每30分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.checkTaskNodeIdle()
		}
	}
}

// 检查任务节点闲置状态
func (s *SchedulerService) checkTaskNodeIdle() {
	ctx := context.Background()
	logx.Info("开始检查任务节点闲置状态")

	// 1. 获取所有进行中的任务节点
	taskNodes, err := s.svcCtx.TaskNodeModel.FindByStatus(ctx, 1) // 进行中
	if err != nil {
		logx.Errorf("查询进行中的任务节点失败: %v", err)
		return
	}

	for _, node := range taskNodes {
		// 2. 检查执行人是否有效
		if err := s.validateAndFixTaskNodeExecutor(ctx, node); err != nil {
			logx.Errorf("检查任务节点 %s 执行人失败: %v", node.TaskNodeId, err)
			continue
		}
	}

	logx.Info("任务节点闲置状态检查完成")
}

// CheckTaskNodeIdle 公开方法，用于测试
func (s *SchedulerService) CheckTaskNodeIdle() {
	s.checkTaskNodeIdle()
}

// 验证并修复任务节点执行人
func (s *SchedulerService) validateAndFixTaskNodeExecutor(ctx context.Context, node *task.TaskNode) error {
	// 检查是否有执行人
	if node.ExecutorId == "" {
		logx.Infof("任务节点 %s 没有执行人，尝试自动派发", node.TaskNodeId)
		return s.autoDispatchTaskNode(ctx, node)
	}

	// 分割执行人ID（支持多个执行人）
	executorIDs := strings.Split(node.ExecutorId, ",")
	var validExecutors []string
	var invalidExecutors []string

	// 检查每个执行人的状态
	for _, executorID := range executorIDs {
		executorID = strings.TrimSpace(executorID)
		if executorID == "" {
			continue
		}

		// 查询员工信息
		employee, err := s.svcCtx.EmployeeModel.FindOneByEmployeeId(ctx, executorID)
		if err != nil {
			logx.Errorf("查询执行人 %s 失败: %v", executorID, err)
			invalidExecutors = append(invalidExecutors, executorID)
			continue
		}

		// 检查员工状态
		if employee.Status == 0 { // 离职
			logx.Infof("执行人 %s 已离职，从任务节点 %s 中移除", executorID, node.TaskNodeId)
			invalidExecutors = append(invalidExecutors, executorID)

			// 发送邮件通知离职员工
			s.notifyLeftEmployee(ctx, employee, node)
		} else {
			validExecutors = append(validExecutors, executorID)
		}
	}

	// 如果有无效的执行人，更新任务节点
	if len(invalidExecutors) > 0 {
		newExecutorID := strings.Join(validExecutors, ",")

		// 更新任务节点执行人
		err := s.svcCtx.TaskNodeModel.UpdateExecutor(ctx, node.TaskNodeId, newExecutorID)
		if err != nil {
			logx.Errorf("更新任务节点 %s 执行人失败: %v", node.TaskNodeId, err)
			return err
		}

		logx.Infof("任务节点 %s 执行人已更新: %s -> %s", node.TaskNodeId, node.ExecutorId, newExecutorID)

		// 如果所有执行人都无效，尝试自动派发
		if len(validExecutors) == 0 {
			logx.Infof("任务节点 %s 所有执行人都无效，尝试自动派发", node.TaskNodeId)
			return s.autoDispatchTaskNode(ctx, node)
		}
	}

	return nil
}

// 自动派发任务节点
func (s *SchedulerService) autoDispatchTaskNode(ctx context.Context, node *task.TaskNode) error {
	// 创建派发服务
	dispatchService := NewDispatchService(s.svcCtx)

	// 执行自动派发
	err := dispatchService.AutoDispatchTask(ctx, node.TaskNodeId)
	if err != nil {
		logx.Errorf("自动派发任务节点 %s 失败: %v", node.TaskNodeId, err)

		// 创建交接记录等待手动分配
		handover := &task.TaskHandover{
			HandoverId:     utils.Common.GenId("handover"),
			TaskId:         node.TaskId,
			FromEmployeeId: node.ExecutorId,
			ToEmployeeId:   "", // 待分配
			HandoverReason: sql.NullString{String: "系统检测到任务节点闲置，自动派发失败", Valid: true},
			HandoverNote:   sql.NullString{String: "等待管理者分配接替者", Valid: true},
			HandoverStatus: 1, // 待处理
			CreateTime:     time.Now(),
			UpdateTime:     time.Now(),
		}

		_, err = s.svcCtx.TaskHandoverModel.Insert(ctx, handover)
		if err != nil {
			logx.Errorf("创建交接记录失败: %v", err)
		}

		return err
	}

	logx.Infof("任务节点 %s 自动派发成功", node.TaskNodeId)
	return nil
}

// 通知离职员工（通过消息队列，消费者会查询并发送）
func (s *SchedulerService) notifyLeftEmployee(ctx context.Context, employee *user.Employee, node *task.TaskNode) {
	// 发布邮件事件（消费者会查询并发送）
	if s.svcCtx.EmailMQService != nil {
		emailEvent := &EmailEvent{
			EventType:  "task.node.executor.left",
			NodeID:     node.TaskNodeId,
			EmployeeID: employee.Id, // 负责人ID
		}
		if err := s.svcCtx.EmailMQService.PublishEmailEvent(ctx, emailEvent); err != nil {
			logx.Errorf("发布离职员工通知邮件事件失败: %v", err)
		}
	}

	// 发布通知事件（消费者会查询并创建）
	if s.svcCtx.NotificationMQService != nil {
		event := &NotificationEvent{
			EventType:   "task.node.executor.left",
			NodeID:      node.TaskNodeId,
			EmployeeIDs: []string{employee.Id}, // 负责人ID，消费者会使用
			Type:        2,                     // 任务通知
			Category:    "task",
			Priority:    2, // 中优先级
			RelatedID:   node.TaskNodeId,
			RelatedType: "task",
		}
		if err := s.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
			logx.Errorf("发布通知事件失败: %v", err)
		}
	}
}
