package svc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"task_Project/model/user_auth"
	"task_Project/task/internal/middleware"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

// SchedulerService 定时任务服务
type SchedulerService struct {
	svcCtx *ServiceContext
}

// NewSchedulerService 创建定时任务服务
func NewSchedulerService(svcCtx *ServiceContext) *SchedulerService {
	return &SchedulerService{
		svcCtx: svcCtx,
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
}

// 任务截止提醒定时任务
func (s *SchedulerService) startDeadlineReminder() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时检查一次
	defer ticker.Stop()

	for {
		select {
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

	// 发送截止提醒
	notificationService := NewNotificationService(s.svcCtx)
	for _, taskNode := range taskNodes {
		if taskNode.NodeStatus == 2 { // 进行中
			if err := notificationService.SendTaskDeadlineReminder(ctx, taskNode); err != nil {
				logx.Errorf("发送任务截止提醒失败: %v", err)
			}
		}
	}
}

// 每日汇报提醒定时任务
func (s *SchedulerService) startDailyReportReminder() {
	// 每天下午5点30分提醒
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			if now.Hour() == 17 && now.Minute() >= 30 && now.Minute() < 31 {
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

	// 发送提醒
	notificationService := NewNotificationService(s.svcCtx)
	for _, employee := range employees {
		if err := notificationService.SendDailyReportReminder(ctx, employee.Id); err != nil {
			logx.Errorf("发送每日汇报提醒失败: %v", err)
		}
	}
}

// 进度缓慢检测定时任务
func (s *SchedulerService) startSlowProgressDetection() {
	ticker := time.NewTicker(2 * time.Hour) // 每2小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkSlowProgress()
		}
	}
}

// 检查进度缓慢的任务
func (s *SchedulerService) checkSlowProgress() {
	ctx := context.Background()

	// 获取进行中的任务节点
	taskNodes, err := s.svcCtx.TaskNodeModel.FindByStatus(ctx, 2) // 进行中
	if err != nil {
		logx.Errorf("查询进行中的任务节点失败: %v", err)
		return
	}

	// 检查进度缓慢的任务
	notificationService := NewNotificationService(s.svcCtx)
	for _, taskNode := range taskNodes {
		// 计算任务开始时间
		startTime, err := time.Parse("2006-01-02 15:04:05", taskNode.CreateTime.Format("2006-01-02 15:04:05"))
		if err != nil {
			continue
		}

		// 计算预期进度
		deadline := taskNode.NodeDeadline

		totalDuration := deadline.Sub(startTime)
		elapsed := time.Now().Sub(startTime)

		if totalDuration > 0 {
			expectedProgress := float64(elapsed) / float64(totalDuration)
			if expectedProgress > 0.5 && taskNode.Progress < int64(expectedProgress*0.5) {
				// 进度缓慢，发送提醒
				if err := notificationService.SendSlowProgressNotification(ctx, taskNode); err != nil {
					logx.Errorf("发送进度缓慢提醒失败: %v", err)
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

	// 发送离职通知
	notificationService := NewNotificationService(s.svcCtx)
	for _, employee := range employees {
		// 检查是否已经发送过通知
		// TODO: 实现检查逻辑
		// lastNotification, err := s.svcCtx.NotificationModel.FindByEmployeeAndType(ctx, employee.Id, "employee_leave")
		// if err == nil && lastNotification != nil {
		// 	// 已经发送过通知，跳过
		// 	continue
		// }

		// 发送离职通知
		if err := notificationService.SendEmployeeLeaveNotification(ctx, employee.Id); err != nil {
			logx.Errorf("发送员工离职通知失败: %v", err)
		}

		// 记录通知
		notification := &user_auth.Notification{
			Id:         utils.NewCommon().GenerateID(),
			EmployeeId: employee.Id,
			Title:      "员工离职通知",
			Content:    fmt.Sprintf("员工 %s 已离职", employee.RealName),
			Type:       3, // 离职通知
			Priority:   3, // 高优先级
			IsRead:     0, // 未读
			SenderId:   sql.NullString{String: "system", Valid: true},
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		}

		_, err = s.svcCtx.NotificationModel.Insert(ctx, notification)
		if err != nil {
			logx.Errorf("记录员工离职通知失败: %v", err)
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

// 发送邮件通知
func (s *SchedulerService) SendEmailNotification(ctx context.Context, to []string, subject, content string) error {
	emailMsg := middleware.EmailMessage{
		To:      to,
		Subject: subject,
		Body:    content,
		IsHTML:  true,
	}

	return s.svcCtx.EmailMiddleware.SendEmail(ctx, emailMsg)
}

// 发送短信通知
func (s *SchedulerService) SendSMSNotification(ctx context.Context, phone, content string) error {
	return s.svcCtx.SMSMiddleware.SendNotificationSMS(ctx, phone, content)
}
