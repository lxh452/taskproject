package svc

import (
	"context"
	"fmt"
	"time"

	"task_Project/task/internal/middleware"

	"github.com/zeromicro/go-zero/core/logx"
)

// UnifiedEmailService 统一邮件服务（替换原有的EmailService、EmailTemplateService、EmailMQService）
type UnifiedEmailService struct {
	templateEngine  *TemplateEngine
	emailMiddleware *middleware.EmailMiddleware
	baseURL         string
}

// NewUnifiedEmailService 创建统一邮件服务
func NewUnifiedEmailService(templateEngine *TemplateEngine, emailMiddleware *middleware.EmailMiddleware, baseURL string) *UnifiedEmailService {
	return &UnifiedEmailService{
		templateEngine:  templateEngine,
		emailMiddleware: emailMiddleware,
		baseURL:         baseURL,
	}
}

// EmailRequest 邮件请求结构
type EmailRequest struct {
	TemplateName string                 `json:"templateName"` // 模板名称
	To           []string               `json:"to"`           // 收件人列表
	Subject      string                 `json:"subject"`      // 邮件主题
	Data         map[string]interface{} `json:"data"`         // 模板数据
	Style        TemplateStyle          `json:"style"`        // 样式配置
	Actions      []TemplateAction       `json:"actions"`      // 操作按钮
	Priority     int                    `json:"priority"`     // 优先级
}

// SendEmail 发送邮件（统一接口）
func (s *UnifiedEmailService) SendEmail(ctx context.Context, req EmailRequest) error {
	// 构建模板数据
	templateData := TemplateData{
		Type:     TemplateTypeEmail,
		Category: "task", // 默认分类，可根据模板名称动态调整
		Title:    req.Subject,
		Message:  req.Subject,
		Data:     req.Data,
		Style:    req.Style,
		Actions:  req.Actions,
		Metadata: TemplateMetadata{
			Timestamp: time.Now(),
			Source:    "task_system",
			Priority:  req.Priority,
		},
	}

	// 渲染模板
	body, err := s.templateEngine.Render(req.TemplateName, templateData)
	if err != nil {
		logx.WithContext(ctx).Errorf("Failed to render email template %s: %v", req.TemplateName, err)
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// 发送邮件
	if s.emailMiddleware != nil {
		if err := s.emailMiddleware.SendEmail(ctx, middleware.EmailMessage{
			To:      req.To,
			Subject: req.Subject,
			Body:    body,
			IsHTML:  true,
		}); err != nil {
			logx.WithContext(ctx).Errorf("Failed to send email: %v", err)
			return fmt.Errorf("failed to send email: %w", err)
		}
	}

	logx.WithContext(ctx).Infof("Email sent successfully: template=%s, recipients=%d", req.TemplateName, len(req.To))
	return nil
}

// SendTaskNotification 发送任务通知邮件（封装常用场景）
func (s *UnifiedEmailService) SendTaskNotification(ctx context.Context, eventType string, to []string, taskData map[string]interface{}) error {
	// 根据事件类型确定模板和样式
	templateName, style := s.getTemplateConfig(eventType)

	req := EmailRequest{
		TemplateName: templateName,
		To:           to,
		Subject:      s.getSubjectByEventType(eventType, taskData),
		Data:         taskData,
		Style:        style,
		Actions:      s.getActionsByEventType(eventType, taskData),
		Priority:     s.getPriorityByEventType(eventType),
	}

	return s.SendEmail(ctx, req)
}

// getTemplateConfig 根据事件类型获取模板配置
func (s *UnifiedEmailService) getTemplateConfig(eventType string) (string, TemplateStyle) {
	switch eventType {
	case "task_created", "task_updated":
		return "email/task/notification", TemplateStyle{Theme: "info"}
	case "task_completed":
		return "email/task/notification", TemplateStyle{Theme: "success"}
	case "task_deadline_reminder":
		return "email/task/notification", TemplateStyle{Theme: "warning"}
	case "task_deleted":
		return "email/task/notification", TemplateStyle{Theme: "danger"}
	case "login_success", "register_success":
		return "email/system/notification", TemplateStyle{Theme: "success"}
	default:
		return "email/task/notification", TemplateStyle{Theme: "info"}
	}
}

// getSubjectByEventType 根据事件类型生成邮件主题
func (s *UnifiedEmailService) getSubjectByEventType(eventType string, data map[string]interface{}) string {
	switch eventType {
	case "task_created":
		if title, ok := data["taskTitle"].(string); ok {
			return fmt.Sprintf("新任务创建 - %s", title)
		}
		return "新任务创建通知"
	case "task_updated":
		if title, ok := data["taskTitle"].(string); ok {
			return fmt.Sprintf("任务更新 - %s", title)
		}
		return "任务更新通知"
	case "task_completed":
		if title, ok := data["taskTitle"].(string); ok {
			return fmt.Sprintf("任务完成 - %s", title)
		}
		return "任务完成通知"
	case "task_deadline_reminder":
		return "任务截止时间提醒"
	case "task_deleted":
		return "任务删除通知"
	case "login_success":
		return "登录成功通知"
	case "register_success":
		return "注册成功通知"
	default:
		return "系统通知"
	}
}

// getActionsByEventType 根据事件类型生成操作按钮
func (s *UnifiedEmailService) getActionsByEventType(eventType string, data map[string]interface{}) []TemplateAction {
	var actions []TemplateAction

	// 基础操作：查看详情
	if taskID, ok := data["taskId"].(string); ok && taskID != "" {
		actions = append(actions, TemplateAction{
			Text: "查看详情",
			URL:  fmt.Sprintf("%s/task/%s", s.baseURL, taskID),
			Type: "primary",
		})
	}

	// 特定事件的操作
	switch eventType {
	case "task_created", "task_updated":
		actions = append(actions, TemplateAction{
			Text: "开始处理",
			URL:  fmt.Sprintf("%s/task/%s/start", s.baseURL, data["taskId"]),
			Type: "secondary",
		})
	case "task_deadline_reminder":
		actions = append(actions, TemplateAction{
			Text: "更新进度",
			URL:  fmt.Sprintf("%s/task/%s/update", s.baseURL, data["taskId"]),
			Type: "warning",
		})
	}

	return actions
}

// getPriorityByEventType 根据事件类型确定优先级
func (s *UnifiedEmailService) getPriorityByEventType(eventType string) int {
	switch eventType {
	case "task_deadline_reminder", "task_deleted":
		return 1 // 高优先级
	case "task_created", "task_updated":
		return 2 // 中优先级
	default:
		return 3 // 低优先级
	}
}

// SendBatchEmails 批量发送邮件
func (s *UnifiedEmailService) SendBatchEmails(ctx context.Context, requests []EmailRequest) error {
	for i, req := range requests {
		if err := s.SendEmail(ctx, req); err != nil {
			logx.WithContext(ctx).Errorf("Failed to send batch email %d/%d: %v", i+1, len(requests), err)
			// 继续发送其他邮件，不中断批量操作
		}
	}
	return nil
}

// TestTemplate 测试模板渲染
func (s *UnifiedEmailService) TestTemplate(ctx context.Context, templateName string, data map[string]interface{}) (string, error) {
	templateData := TemplateData{
		Type:     TemplateTypeEmail,
		Category: "test",
		Title:    "测试邮件",
		Message:  "这是一封测试邮件",
		Data:     data,
		Style:    TemplateStyle{Theme: "info"},
		Metadata: TemplateMetadata{
			Timestamp: time.Now(),
			Source:    "test",
			Priority:  3,
		},
	}

	return s.templateEngine.Render(templateName, templateData)
}
