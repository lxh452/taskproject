package svc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"task_Project/model/task"
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

	// 启动任务节点闲置检查定时任务
	go s.startTaskNodeIdleCheck()
}

// 任务截止提醒定时任务
func (s *SchedulerService) startDeadlineReminder() {
	ticker := time.NewTicker(4 * time.Hour) // 每4小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			// 检查是否在工作时间（9:00-18:00）
			hour := time.Now().Hour()
			if hour >= 9 && hour < 18 {
				s.checkTaskDeadlines()
			}
		}
	}
}

// 检查任务截止时间
func (s *SchedulerService) checkTaskDeadlines() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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
	ticker := time.NewTicker(4 * time.Hour) // 每4小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			now := time.Now()
			hour := now.Hour()
			// 检查是否在工作时间（9:00-18:00）且在下午5点到6点之间
			if hour >= 9 && hour < 18 && hour == 17 {
				s.sendDailyReportReminders()
			}
		}
	}
}

// 发送每日汇报提醒
func (s *SchedulerService) sendDailyReportReminders() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

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
	ticker := time.NewTicker(4 * time.Hour) // 每4小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			// 检查是否在工作时间（9:00-18:00）
			hour := time.Now().Hour()
			if hour >= 9 && hour < 18 {
				s.checkSlowProgress()
			}
		}
	}
}

// 检查进度缓慢的任务
func (s *SchedulerService) checkSlowProgress() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

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
	ticker := time.NewTicker(4 * time.Hour) // 每4小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			// 检查是否在工作时间（9:00-18:00）
			hour := time.Now().Hour()
			if hour >= 9 && hour < 18 {
				s.checkTaskNodeIdle()
			}
		}
	}
}

// 检查任务节点闲置状态
func (s *SchedulerService) checkTaskNodeIdle() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	logx.Info("开始检查任务节点闲置状态")

	// 1. 获取所有进行中的任务节点
	taskNodes, err := s.svcCtx.TaskNodeModel.FindByStatus(ctx, 1) // 进行中
	if err != nil {
		logx.Errorf("查询进行中的任务节点失败: %v", err)
		return
	}

	for _, node := range taskNodes {
		// 2. 检查执行人是否有效
		if err := s.validateTaskNodeExecutor(ctx, node); err != nil {
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

// 验证任务节点执行人（仅检查是否有执行人，不检查离职状态）
func (s *SchedulerService) validateTaskNodeExecutor(ctx context.Context, node *task.TaskNode) error {
	// 检查是否有执行人
	if node.ExecutorId == "" {
		logx.Infof("任务节点 %s 没有执行人，需要手动分配", node.TaskNodeId)
		return nil
	}

	// 已有执行人，无需处理
	return nil
}
