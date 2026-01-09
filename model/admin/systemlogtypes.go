package admin

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// SystemLog 系统日志文档结构
type SystemLog struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	LogID      string        `bson:"logId" json:"logId"`           // 日志ID
	Level      string        `bson:"level" json:"level"`           // 日志级别: debug/info/warn/error/fatal
	Module     string        `bson:"module" json:"module"`         // 模块名称
	Action     string        `bson:"action" json:"action"`         // 操作类型
	Message    string        `bson:"message" json:"message"`       // 日志消息
	Detail     string        `bson:"detail" json:"detail"`         // 详细信息
	UserID     string        `bson:"userId" json:"userId"`         // 操作用户ID
	UserType   string        `bson:"userType" json:"userType"`     // 用户类型: user/admin
	IP         string        `bson:"ip" json:"ip"`                 // 请求IP
	UserAgent  string        `bson:"userAgent" json:"userAgent"`   // 浏览器UA
	RequestID  string        `bson:"requestId" json:"requestId"`   // 请求ID
	TraceID    string        `bson:"traceId" json:"traceId"`       // 追踪ID
	StackTrace string        `bson:"stackTrace" json:"stackTrace"` // 堆栈信息
	Extra      bson.M        `bson:"extra" json:"extra"`           // 额外信息
	CreateAt   time.Time     `bson:"createAt" json:"createAt"`
}
