package svc

import (
	"context"
	"time"

	adminModel "task_Project/model/admin"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// SystemLogService 系统日志服务
type SystemLogService struct {
	model adminModel.SystemLogModel
}

// NewSystemLogService 创建系统日志服务
func NewSystemLogService(model adminModel.SystemLogModel) *SystemLogService {
	return &SystemLogService{model: model}
}

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// LogEntry 日志条目
type LogEntry struct {
	Level      LogLevel
	Module     string
	Action     string
	Message    string
	Detail     string
	UserID     string
	UserType   string
	IP         string
	UserAgent  string
	RequestID  string
	TraceID    string
	StackTrace string
	Extra      map[string]interface{}
}

// Log 记录日志
func (s *SystemLogService) Log(ctx context.Context, entry LogEntry) {
	if s.model == nil {
		logx.Errorf("SystemLogModel is nil, cannot log: %+v", entry)
		return
	}

	go func() {
		log := &adminModel.SystemLog{
			ID:         bson.NewObjectID(),
			LogID:      utils.Common.GenId("log"),
			Level:      string(entry.Level),
			Module:     entry.Module,
			Action:     entry.Action,
			Message:    entry.Message,
			Detail:     entry.Detail,
			UserID:     entry.UserID,
			UserType:   entry.UserType,
			IP:         entry.IP,
			UserAgent:  entry.UserAgent,
			RequestID:  entry.RequestID,
			TraceID:    entry.TraceID,
			StackTrace: entry.StackTrace,
			Extra:      bson.M(entry.Extra),
			CreateAt:   time.Now(),
		}

		if err := s.model.Insert(context.Background(), log); err != nil {
			logx.Errorf("Failed to insert system log: %v", err)
		}
	}()
}

// Info 记录信息日志
func (s *SystemLogService) Info(ctx context.Context, module, action, message string, userID, userType string) {
	s.Log(ctx, LogEntry{
		Level:    LogLevelInfo,
		Module:   module,
		Action:   action,
		Message:  message,
		UserID:   userID,
		UserType: userType,
	})
}

// Warn 记录警告日志
func (s *SystemLogService) Warn(ctx context.Context, module, action, message string, userID, userType string) {
	s.Log(ctx, LogEntry{
		Level:    LogLevelWarn,
		Module:   module,
		Action:   action,
		Message:  message,
		UserID:   userID,
		UserType: userType,
	})
}

// Error 记录错误日志
func (s *SystemLogService) Error(ctx context.Context, module, action, message, detail string, userID, userType string) {
	s.Log(ctx, LogEntry{
		Level:    LogLevelError,
		Module:   module,
		Action:   action,
		Message:  message,
		Detail:   detail,
		UserID:   userID,
		UserType: userType,
	})
}

// ErrorWithStack 记录带堆栈的错误日志
func (s *SystemLogService) ErrorWithStack(ctx context.Context, module, action, message, detail, stackTrace string, userID, userType string) {
	s.Log(ctx, LogEntry{
		Level:      LogLevelError,
		Module:     module,
		Action:     action,
		Message:    message,
		Detail:     detail,
		StackTrace: stackTrace,
		UserID:     userID,
		UserType:   userType,
	})
}

// UserAction 记录用户操作日志
func (s *SystemLogService) UserAction(ctx context.Context, module, action, message string, userID string, ip, userAgent string) {
	s.Log(ctx, LogEntry{
		Level:     LogLevelInfo,
		Module:    module,
		Action:    action,
		Message:   message,
		UserID:    userID,
		UserType:  "user",
		IP:        ip,
		UserAgent: userAgent,
	})
}

// AdminAction 记录管理员操作日志
func (s *SystemLogService) AdminAction(ctx context.Context, module, action, message string, adminID string, ip, userAgent string) {
	s.Log(ctx, LogEntry{
		Level:     LogLevelInfo,
		Module:    module,
		Action:    action,
		Message:   message,
		UserID:    adminID,
		UserType:  "admin",
		IP:        ip,
		UserAgent: userAgent,
	})
}
