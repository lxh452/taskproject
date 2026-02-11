package admin

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// SystemLog 系统日志结构体
type SystemLog struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	LogID      string        `bson:"logId" json:"logId"`           // 日志ID
	Level      string        `bson:"level" json:"level"`           // 日志级别: info, warn, error
	Module     string        `bson:"module" json:"module"`         // 模块名称
	Action     string        `bson:"action" json:"action"`         // 操作类型
	Message    string        `bson:"message" json:"message"`       // 日志消息
	Detail     string        `bson:"detail" json:"detail"`         // 详细信息
	UserID     string        `bson:"userId" json:"userId"`         // 用户ID
	UserType   string        `bson:"userType" json:"userType"`     // 用户类型: user, admin
	IP         string        `bson:"ip" json:"ip"`                 // IP地址
	UserAgent  string        `bson:"userAgent" json:"userAgent"`   // 浏览器UA
	RequestID  string        `bson:"requestId" json:"requestId"`   // 请求ID
	TraceID    string        `bson:"traceId" json:"traceId"`       // 追踪ID
	StackTrace string        `bson:"stackTrace" json:"stackTrace"` // 堆栈追踪
	Extra      bson.M        `bson:"extra" json:"extra"`           // 额外信息
	CreateAt   time.Time     `bson:"createAt" json:"createAt"`     // 创建时间
}

type systemLogModel interface {
	Insert(ctx context.Context, data *SystemLog) error
	FindOne(ctx context.Context, id string) (*SystemLog, error)
	FindByLogID(ctx context.Context, logID string) (*SystemLog, error)
	FindList(ctx context.Context, filter SystemLogFilter, page, pageSize int64) ([]*SystemLog, int64, error)
	Delete(ctx context.Context, id string) (int64, error)
	DeleteByTime(ctx context.Context, before time.Time) (int64, error)
}

// SystemLogFilter 系统日志查询过滤条件
type SystemLogFilter struct {
	Level     string    // 日志级别
	Module    string    // 模块名称
	Keyword   string    // 关键词搜索
	UserID    string    // 用户ID
	UserType  string    // 用户类型: user, admin
	StartTime time.Time // 开始时间
	EndTime   time.Time // 结束时间
}

type defaultSystemLogModel struct {
	conn *mon.Model
}

func newDefaultSystemLogModel(conn *mon.Model) *defaultSystemLogModel {
	return &defaultSystemLogModel{conn: conn}
}

func (m *defaultSystemLogModel) Insert(ctx context.Context, data *SystemLog) error {
	if data.ID.IsZero() {
		data.ID = bson.NewObjectID()
		data.CreateAt = time.Now()
	}
	_, err := m.conn.InsertOne(ctx, data)
	return err
}

func (m *defaultSystemLogModel) FindOne(ctx context.Context, id string) (*SystemLog, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var data SystemLog
	err = m.conn.FindOne(ctx, &data, bson.M{"_id": oid})
	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultSystemLogModel) FindByLogID(ctx context.Context, logID string) (*SystemLog, error) {
	var data SystemLog
	err := m.conn.FindOne(ctx, &data, bson.M{"logId": logID})
	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultSystemLogModel) FindList(ctx context.Context, filter SystemLogFilter, page, pageSize int64) ([]*SystemLog, int64, error) {
	query := bson.M{}

	// 日志级别筛选
	if filter.Level != "" {
		query["level"] = filter.Level
	}

	// 模块筛选
	if filter.Module != "" {
		query["module"] = filter.Module
	}

	// 用户ID筛选
	if filter.UserID != "" {
		query["userId"] = filter.UserID
	}

	// 用户类型筛选
	if filter.UserType != "" {
		query["userType"] = filter.UserType
	}

	// 关键词搜索（搜索message和detail字段）
	if filter.Keyword != "" {
		query["$or"] = []bson.M{
			{"message": bson.M{"$regex": filter.Keyword, "$options": "i"}},
			{"detail": bson.M{"$regex": filter.Keyword, "$options": "i"}},
			{"module": bson.M{"$regex": filter.Keyword, "$options": "i"}},
			{"action": bson.M{"$regex": filter.Keyword, "$options": "i"}},
		}
	}

	// 时间范围筛选
	if !filter.StartTime.IsZero() || !filter.EndTime.IsZero() {
		timeQuery := bson.M{}
		if !filter.StartTime.IsZero() {
			timeQuery["$gte"] = filter.StartTime
		}
		if !filter.EndTime.IsZero() {
			timeQuery["$lte"] = filter.EndTime
		}
		query["createAt"] = timeQuery
	}

	// 计算总数
	total, err := m.conn.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询，按时间倒序
	skip := (page - 1) * pageSize
	opts := options.Find().SetSort(bson.M{"createAt": -1}).SetSkip(skip).SetLimit(pageSize)

	var data []*SystemLog
	err = m.conn.Find(ctx, &data, query, opts)
	if err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (m *defaultSystemLogModel) Delete(ctx context.Context, id string) (int64, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return 0, err
	}

	res, err := m.conn.DeleteOne(ctx, bson.M{"_id": oid})
	return res, err
}

func (m *defaultSystemLogModel) DeleteByTime(ctx context.Context, before time.Time) (int64, error) {
	res, err := m.conn.DeleteMany(ctx, bson.M{"createAt": bson.M{"$lt": before}})
	return res, err
}
