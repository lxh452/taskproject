package svc

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// EmailTemplateService 邮件模板服务
type EmailTemplateService struct {
	templates map[string]*template.Template
}

// NewEmailTemplateService 创建邮件模板服务
func NewEmailTemplateService() (*EmailTemplateService, error) {
	service := &EmailTemplateService{
		templates: make(map[string]*template.Template),
	}

	// 获取模板文件目录
	// 在 Docker 容器中，工作目录是 /app，模板在 /app/task/internal/templates/email/
	// 在本地开发时，可能在项目根目录或 task 目录下运行
	possiblePaths := []string{
		"task/internal/templates/email",      // 从项目根目录运行
		"./task/internal/templates/email",    // 从项目根目录运行（显式相对路径）
		"internal/templates/email",           // 从 task 目录运行
		"../templates/email",                 // 从 svc 目录运行
		"/app/task/internal/templates/email", // Docker 容器绝对路径
	}

	var templateDir string
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			templateDir = path
			logx.Infof("[EmailTemplateService] Found template directory: %s", path)
			break
		}
	}

	if templateDir == "" {
		logx.Errorf("[EmailTemplateService] Template directory not found, tried paths: %v", possiblePaths)
		// 获取当前工作目录用于调试
		if cwd, err := os.Getwd(); err == nil {
			logx.Errorf("[EmailTemplateService] Current working directory: %s", cwd)
		}
		return nil, fmt.Errorf("template directory not found")
	}

	// 加载所有模板文件
	templateFiles := map[string]string{
		"task_deadline_reminder":  "task_deadline_reminder.tpl",
		"task_completed":          "task_completed.tpl",
		"task_updated":            "task_updated.tpl",
		"task_deleted":            "task_deleted.tpl",
		"task_node_deleted":       "task_node_deleted.tpl",
		"task_node_created":       "task_node_created.tpl",
		"task_node_executor_left": "task_node_executor_left.tpl",
		"task_slow_progress":      "task_slow_progress.tpl",
		"handover":                "handover.tpl",
		"employee_leave":          "employee_leave.tpl",
		"cross_department":        "cross_department.tpl",
		"login_success":           "login_success.tpl",
		"register_success":        "register_success.tpl",
		"onboarding":              "onboarding.tpl",
		"daily_report_reminder":   "daily_report_reminder.tpl",
	}

	for name, fileName := range templateFiles {
		filePath := filepath.Join(templateDir, fileName)
		tmpl, err := template.ParseFiles(filePath)
		if err != nil {
			logx.Errorf("Failed to parse template %s from %s: %v", name, filePath, err)
			continue
		}
		service.templates[name] = tmpl
		logx.Infof("Loaded email template: %s from %s", name, filePath)
	}

	return service, nil
}

// RenderTemplate 渲染模板
func (s *EmailTemplateService) RenderTemplate(templateName string, data interface{}) (string, error) {
	tmpl, ok := s.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// TaskDeadlineReminderData 任务截止提醒邮件数据
type TaskDeadlineReminderData struct {
	BaseURL  string
	NodeName string
	Deadline string
	Progress int
	Year     int
}

// TaskCompletedData 任务完成邮件数据
type TaskCompletedData struct {
	BaseURL      string
	TaskTitle    string
	NodeName     string
	CompleteTime string
	TaskID       string
	Year         int
}

// HandoverData 交接邮件数据
type HandoverData struct {
	BaseURL      string
	EmployeeName string
	Message      string
	HandoverID   string
	TaskTitle    string
	Year         int
}

// LoginSuccessData 登录成功邮件数据
type LoginSuccessData struct {
	BaseURL    string
	Username   string
	LoginTime  string
	LoginIP    string
	DeviceInfo string
	Year       int
}

// RegisterSuccessData 注册成功邮件数据
type RegisterSuccessData struct {
	BaseURL      string
	Username     string
	Email        string
	RegisterTime string
	Year         int
}

// OnboardingData 入职通知邮件数据
type OnboardingData struct {
	BaseURL        string
	EmployeeName   string
	CompanyName    string
	DepartmentName string
	PositionName   string
	OnboardingTime string
	Year           int
}

// TaskCreatedData 任务创建邮件数据
type TaskCreatedData struct {
	BaseURL   string
	TaskTitle string
	TaskID    string
	Year      int
}

// TaskUpdatedData 任务更新邮件数据
type TaskUpdatedData struct {
	BaseURL    string
	TaskTitle  string
	TaskID     string
	UpdateTime string
	UpdateNote string
	Year       int
}

// TaskSlowProgressData 任务进度缓慢提醒邮件数据
type TaskSlowProgressData struct {
	BaseURL  string
	NodeName string
	Progress int
	Deadline string
	Year     int
}

// DailyReportReminderData 每日汇报提醒邮件数据
type DailyReportReminderData struct {
	BaseURL      string
	EmployeeName string
	Year         int
}

// TaskNodeExecutorLeftData 任务节点执行人离职通知邮件数据
type TaskNodeExecutorLeftData struct {
	BaseURL          string
	TaskTitle        string
	NodeName         string
	NodeDetail       string
	LeftEmployeeName string
	Year             int
}

// TaskDeletedData 任务删除邮件数据
type TaskDeletedData struct {
	BaseURL      string
	TaskTitle    string
	TaskID       string
	OperatorName string
	DeleteTime   string
	Year         int
}

// TaskNodeDeletedData 任务节点删除邮件数据
type TaskNodeDeletedData struct {
	BaseURL      string
	TaskTitle    string
	NodeName     string
	OperatorName string
	DeleteTime   string
	Year         int
}

// GetCurrentYear 获取当前年份（模板辅助函数）
func GetCurrentYear() int {
	return time.Now().Year()
}
