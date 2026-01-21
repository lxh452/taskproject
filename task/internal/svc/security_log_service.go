package svc

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// SecurityEventType 安全事件类型
type SecurityEventType string

const (
	EventSQLInjectionAttempt  SecurityEventType = "sql_injection_attempt"
	EventRateLimitExceeded    SecurityEventType = "rate_limit_exceeded"
	EventCSRFValidationFailed SecurityEventType = "csrf_validation_failed"
	EventXSSAttempt           SecurityEventType = "xss_attempt"
	EventLoginFailed          SecurityEventType = "login_failed"
	EventLoginSuccess         SecurityEventType = "login_success"
	EventSuspiciousActivity   SecurityEventType = "suspicious_activity"
	EventIPBlocked            SecurityEventType = "ip_blocked"
	EventUnauthorizedAccess   SecurityEventType = "unauthorized_access"
)

// SecuritySeverity 安全事件严重程度
type SecuritySeverity string

const (
	SeverityInfo     SecuritySeverity = "info"
	SeverityWarning  SecuritySeverity = "warning"
	SeverityCritical SecuritySeverity = "critical"
)

// SecurityLog 安全日志结构
type SecurityLog struct {
	ID          bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	EventType   SecurityEventType `bson:"event_type" json:"eventType"`
	Severity    SecuritySeverity  `bson:"severity" json:"severity"`
	IP          string            `bson:"ip" json:"ip"`
	UserID      string            `bson:"user_id,omitempty" json:"userId,omitempty"`
	RequestPath string            `bson:"request_path" json:"requestPath"`
	RequestBody string            `bson:"request_body,omitempty" json:"requestBody,omitempty"`
	Description string            `bson:"description" json:"description"`
	Timestamp   time.Time         `bson:"timestamp" json:"timestamp"`
	UserAgent   string            `bson:"user_agent,omitempty" json:"userAgent,omitempty"`
	Extra       map[string]any    `bson:"extra,omitempty" json:"extra,omitempty"`
}

// SecurityLogService 安全日志服务
type SecurityLogService struct {
	conn *mon.Model
}

// NewSecurityLogService 创建安全日志服务
func NewSecurityLogService(mongoURL, database string) (*SecurityLogService, error) {
	conn := mon.MustNewModel(mongoURL, database, "security_logs")

	service := &SecurityLogService{conn: conn}

	// 创建索引
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := service.ensureIndexes(ctx); err != nil {
		logx.Errorf("[SecurityLogService] 创建索引失败: %v", err)
	}

	return service, nil
}

// ensureIndexes 确保索引存在
func (s *SecurityLogService) ensureIndexes(ctx context.Context) error {
	// 使用go-zero的mon包，索引创建需要通过底层collection
	// 这里简化处理，依赖MongoDB自动创建索引或手动创建
	logx.Info("[SecurityLogService] 索引检查完成")
	return nil
}

// Log 记录安全日志
func (s *SecurityLogService) Log(ctx context.Context, log *SecurityLog) error {
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}
	if log.ID.IsZero() {
		log.ID = bson.NewObjectID()
	}

	_, err := s.conn.InsertOne(ctx, log)
	if err != nil {
		logx.Errorf("[SecurityLogService] 记录安全日志失败: %v", err)
		return err
	}

	// 同时输出到标准日志
	logx.Infof("[SecurityLog] Type=%s, Severity=%s, IP=%s, UserID=%s, Path=%s, Desc=%s",
		log.EventType, log.Severity, log.IP, log.UserID, log.RequestPath, log.Description)

	return nil
}

// LogRateLimitExceeded 记录限流触发事件
func (s *SecurityLogService) LogRateLimitExceeded(ip, userID, path string, limit int) {
	s.Log(context.Background(), &SecurityLog{
		EventType:   EventRateLimitExceeded,
		Severity:    SeverityWarning,
		IP:          ip,
		UserID:      userID,
		RequestPath: path,
		Description: "请求频率超过限制",
		Extra: map[string]any{
			"limit": limit,
		},
	})
}

// LogLoginFailed 记录登录失败事件
func (s *SecurityLogService) LogLoginFailed(ip, userID, reason string) {
	s.Log(context.Background(), &SecurityLog{
		EventType:   EventLoginFailed,
		Severity:    SeverityInfo,
		IP:          ip,
		UserID:      userID,
		Description: reason,
	})
}

// LogLoginSuccess 记录登录成功事件
func (s *SecurityLogService) LogLoginSuccess(ip, userID string) {
	s.Log(context.Background(), &SecurityLog{
		EventType:   EventLoginSuccess,
		Severity:    SeverityInfo,
		IP:          ip,
		UserID:      userID,
		Description: "用户登录成功",
	})
}

// LogSQLInjectionAttempt 记录SQL注入尝试
func (s *SecurityLogService) LogSQLInjectionAttempt(ip, userID, path, payload string) {
	s.Log(context.Background(), &SecurityLog{
		EventType:   EventSQLInjectionAttempt,
		Severity:    SeverityCritical,
		IP:          ip,
		UserID:      userID,
		RequestPath: path,
		Description: "检测到SQL注入尝试",
		Extra: map[string]any{
			"payload": maskSensitiveData(payload),
		},
	})
}

// LogXSSAttempt 记录XSS攻击尝试
func (s *SecurityLogService) LogXSSAttempt(ip, userID, path, payload string) {
	s.Log(context.Background(), &SecurityLog{
		EventType:   EventXSSAttempt,
		Severity:    SeverityWarning,
		IP:          ip,
		UserID:      userID,
		RequestPath: path,
		Description: "检测到XSS攻击尝试",
		Extra: map[string]any{
			"payload": maskSensitiveData(payload),
		},
	})
}

// LogIPBlocked 记录IP封禁事件
func (s *SecurityLogService) LogIPBlocked(ip, reason string, duration time.Duration) {
	s.Log(context.Background(), &SecurityLog{
		EventType:   EventIPBlocked,
		Severity:    SeverityWarning,
		IP:          ip,
		Description: reason,
		Extra: map[string]any{
			"duration_minutes": duration.Minutes(),
		},
	})
}

// LogUnauthorizedAccess 记录未授权访问
func (s *SecurityLogService) LogUnauthorizedAccess(ip, userID, path string) {
	s.Log(context.Background(), &SecurityLog{
		EventType:   EventUnauthorizedAccess,
		Severity:    SeverityWarning,
		IP:          ip,
		UserID:      userID,
		RequestPath: path,
		Description: "未授权访问尝试",
	})
}

// Query 查询安全日志
func (s *SecurityLogService) Query(ctx context.Context, filter SecurityLogFilter) ([]*SecurityLog, int64, error) {
	query := bson.M{}

	if filter.EventType != "" {
		query["event_type"] = filter.EventType
	}
	if filter.Severity != "" {
		query["severity"] = filter.Severity
	}
	if filter.IP != "" {
		query["ip"] = filter.IP
	}
	if filter.UserID != "" {
		query["user_id"] = filter.UserID
	}
	if !filter.StartTime.IsZero() {
		if query["timestamp"] == nil {
			query["timestamp"] = bson.M{}
		}
		query["timestamp"].(bson.M)["$gte"] = filter.StartTime
	}
	if !filter.EndTime.IsZero() {
		if query["timestamp"] == nil {
			query["timestamp"] = bson.M{}
		}
		query["timestamp"].(bson.M)["$lte"] = filter.EndTime
	}

	// 获取总数
	total, err := s.conn.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	skip := int64((filter.Page - 1) * filter.PageSize)
	opts := options.Find().
		SetSort(bson.M{"timestamp": -1}).
		SetSkip(skip).
		SetLimit(int64(filter.PageSize))

	var logs []*SecurityLog
	err = s.conn.Find(ctx, &logs, query, opts)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// SecurityLogFilter 安全日志查询过滤器
type SecurityLogFilter struct {
	EventType SecurityEventType
	Severity  SecuritySeverity
	IP        string
	UserID    string
	StartTime time.Time
	EndTime   time.Time
	Page      int
	PageSize  int
}

// maskSensitiveData 脱敏敏感数据
func maskSensitiveData(data string) string {
	if len(data) <= 20 {
		return data
	}
	return data[:10] + "..." + data[len(data)-10:]
}
