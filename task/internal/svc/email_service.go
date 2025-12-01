package svc

import (
	"context"
	"fmt"
	"time"

	"task_Project/task/internal/middleware"

	"github.com/zeromicro/go-zero/core/logx"
)

// EmailService 邮件服务（统一处理邮件发送，使用模板和消息队列）
type EmailService struct {
	templateService *EmailTemplateService
	emailMQService  *EmailMQService
	emailMiddleware *middleware.EmailMiddleware
	baseURL         string
}

// NewEmailService 创建邮件服务
func NewEmailService(templateService *EmailTemplateService, emailMQService *EmailMQService, emailMiddleware *middleware.EmailMiddleware, baseURL string) *EmailService {
	return &EmailService{
		templateService: templateService,
		emailMQService:  emailMQService,
		emailMiddleware: emailMiddleware,
		baseURL:         baseURL,
	}
}

// SendTaskDispatchEmail 发送任务派发邮件
func (s *EmailService) SendTaskDispatchEmail(ctx context.Context, employeeEmail, employeeName, taskTitle, nodeName, nodeDetail, deadline, taskId string) error {
	data := TaskDispatchData{
		BaseURL:      s.baseURL,
		EmployeeName: employeeName,
		TaskTitle:    taskTitle,
		NodeName:     nodeName,
		NodeDetail:   nodeDetail,
		Deadline:     deadline,
		TaskId:       taskId,
		Year:         time.Now().Year(),
	}

	body, err := s.templateService.RenderTemplate("task_dispatch", data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sendEmail(ctx, []string{employeeEmail}, "任务派发通知", body)
}

// SendTaskDeadlineReminderEmail 发送任务截止提醒邮件
func (s *EmailService) SendTaskDeadlineReminderEmail(ctx context.Context, employeeEmail, nodeName, deadline string, progress int) error {
	data := TaskDeadlineReminderData{
		BaseURL:  s.baseURL,
		NodeName: nodeName,
		Deadline: deadline,
		Progress: progress,
		Year:     time.Now().Year(),
	}

	body, err := s.templateService.RenderTemplate("task_deadline_reminder", data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sendEmail(ctx, []string{employeeEmail}, "任务截止时间提醒", body)
}

// SendTaskCompletedEmail 发送任务完成邮件
func (s *EmailService) SendTaskCompletedEmail(ctx context.Context, employeeEmail, taskTitle, nodeName, completeTime string) error {
	data := TaskCompletedData{
		BaseURL:      s.baseURL,
		TaskTitle:    taskTitle,
		NodeName:     nodeName,
		CompleteTime: completeTime,
		// 完成通知里也可以带任务ID，方便跳转
		TaskID: "",
		Year:   time.Now().Year(),
	}

	body, err := s.templateService.RenderTemplate("task_completed", data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sendEmail(ctx, []string{employeeEmail}, "任务完成通知", body)
}

// SendHandoverEmail 发送交接邮件
func (s *EmailService) SendHandoverEmail(ctx context.Context, employeeEmail, employeeName, message, handoverID, taskTitle string) error {
	data := HandoverData{
		BaseURL:      s.baseURL,
		EmployeeName: employeeName,
		Message:      message,
		HandoverID:   handoverID,
		TaskTitle:    taskTitle,
		Year:         time.Now().Year(),
	}

	body, err := s.templateService.RenderTemplate("handover", data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sendEmail(ctx, []string{employeeEmail}, "任务交接通知", body)
}

// SendEmployeeLeaveEmail 发送员工离职邮件
func (s *EmailService) SendEmployeeLeaveEmail(ctx context.Context, recipientEmail, employeeName string, taskNodes []string) error {
	data := struct {
		BaseURL      string
		EmployeeName string
		TaskNodes    []string
		Year         int
	}{
		BaseURL:      s.baseURL,
		EmployeeName: employeeName,
		TaskNodes:    taskNodes,
		Year:         time.Now().Year(),
	}

	body, err := s.templateService.RenderTemplate("employee_leave", data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sendEmail(ctx, []string{recipientEmail}, "员工离职任务交接通知", body)
}

// SendLoginSuccessEmail 发送登录成功邮件
func (s *EmailService) SendLoginSuccessEmail(ctx context.Context, employeeEmail, username, loginTime, loginIP string) error {
	data := LoginSuccessData{
		BaseURL:   s.baseURL,
		Username:  username,
		LoginTime: loginTime,
		LoginIP:   loginIP,
		Year:      time.Now().Year(),
	}

	body, err := s.templateService.RenderTemplate("login_success", data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sendEmail(ctx, []string{employeeEmail}, "登录成功通知", body)
}

// SendRegisterSuccessEmail 发送注册成功邮件
func (s *EmailService) SendRegisterSuccessEmail(ctx context.Context, email, username, registerTime string) error {
	data := RegisterSuccessData{
		BaseURL:      s.baseURL,
		Username:     username,
		Email:        email,
		RegisterTime: registerTime,
		Year:         time.Now().Year(),
	}

	body, err := s.templateService.RenderTemplate("register_success", data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sendEmail(ctx, []string{email}, "注册成功通知", body)
}

// SendOnboardingEmail 发送入职通知邮件
func (s *EmailService) SendOnboardingEmail(ctx context.Context, employeeEmail, employeeName, companyName, departmentName, positionName, onboardingTime string) error {
	data := OnboardingData{
		BaseURL:        s.baseURL,
		EmployeeName:   employeeName,
		CompanyName:    companyName,
		DepartmentName: departmentName,
		PositionName:   positionName,
		OnboardingTime: onboardingTime,
		Year:           time.Now().Year(),
	}

	body, err := s.templateService.RenderTemplate("onboarding", data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sendEmail(ctx, []string{employeeEmail}, "入职通知", body)
}

// SendTaskUpdatedEmail 发送任务更新邮件
func (s *EmailService) SendTaskUpdatedEmail(ctx context.Context, employeeEmail, taskTitle, taskID, updateTime, updateNote string) error {
	data := TaskUpdatedData{
		BaseURL:    s.baseURL,
		TaskTitle:  taskTitle,
		TaskID:     taskID,
		UpdateTime: updateTime,
		UpdateNote: updateNote,
		Year:       time.Now().Year(),
	}

	body, err := s.templateService.RenderTemplate("task_updated", data)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sendEmail(ctx, []string{employeeEmail}, "任务更新通知", body)
}

// sendEmail 统一发送邮件方法（通过消息队列）
func (s *EmailService) sendEmail(ctx context.Context, to []string, subject, body string) error {
	logx.WithContext(ctx).Infof("[EmailService] sendEmail called: to=%v, subject=%s, bodyLength=%d",
		to, subject, len(body))

	// 优先使用消息队列
	if s.emailMQService != nil {
		logx.WithContext(ctx).Infof("[EmailService] Using message queue to send email")
		emailEvent := &EmailEvent{
			EventType: "send",
			To:        to,
			Subject:   subject,
			Body:      body,
			IsHTML:    true,
		}
		if err := s.emailMQService.PublishEmailEvent(ctx, emailEvent); err != nil {
			logx.WithContext(ctx).Errorf("[EmailService] Failed to publish email event: error=%v, to=%v, subject=%s",
				err, to, subject)
			return err
		}
		logx.WithContext(ctx).Infof("[EmailService] Email event published successfully: to=%v, subject=%s", to, subject)
		return nil
	}

	// 降级：直接发送（如果消息队列未配置）
	if s.emailMiddleware != nil {
		logx.WithContext(ctx).Infof("[EmailService] Message queue not available, sending email directly")
		emailMsg := middleware.EmailMessage{
			To:      to,
			Subject: subject,
			Body:    body,
			IsHTML:  true,
		}
		err := s.emailMiddleware.SendEmail(ctx, emailMsg)
		if err != nil {
			logx.WithContext(ctx).Errorf("[EmailService] Failed to send email directly: error=%v, to=%v, subject=%s",
				err, to, subject)
		} else {
			logx.WithContext(ctx).Infof("[EmailService] Email sent directly successfully: to=%v, subject=%s", to, subject)
		}
		return err
	}

	logx.WithContext(ctx).Errorf("[EmailService] Email service not available: emailMQService=%v, emailMiddleware=%v",
		s.emailMQService != nil, s.emailMiddleware != nil)
	return fmt.Errorf("email service not available")
}
