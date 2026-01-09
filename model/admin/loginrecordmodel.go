package admin

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ LoginRecordModel = (*customLoginRecordModel)(nil)

type (
	// LoginRecordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customLoginRecordModel.
	LoginRecordModel interface {
		loginRecordModel
		withSession(session sqlx.Session) LoginRecordModel

		// 登录记录查询操作
		FindByUserId(ctx context.Context, userId string, page, pageSize int) ([]*LoginRecord, int64, error)
		FindByUserType(ctx context.Context, userType string, page, pageSize int) ([]*LoginRecord, int64, error)
		FindByStatus(ctx context.Context, status int, page, pageSize int) ([]*LoginRecord, int64, error)
		FindByTimeRange(ctx context.Context, startTime, endTime string, page, pageSize int) ([]*LoginRecord, int64, error)
		FindByPage(ctx context.Context, page, pageSize int) ([]*LoginRecord, int64, error)
		FindByFilters(ctx context.Context, userId, userType string, status int, startTime, endTime string, page, pageSize int) ([]*LoginRecord, int64, error)
		GetRecentByUserId(ctx context.Context, userId string, limit int) ([]*LoginRecord, error)
		GetLoginRecordCount(ctx context.Context) (int64, error)
		GetFailedLoginCount(ctx context.Context, userId string, startTime string) (int64, error)
	}

	customLoginRecordModel struct {
		*defaultLoginRecordModel
	}
)

// NewLoginRecordModel returns a model for the database table.
func NewLoginRecordModel(conn sqlx.SqlConn) LoginRecordModel {
	return &customLoginRecordModel{
		defaultLoginRecordModel: newLoginRecordModel(conn),
	}
}

func (m *customLoginRecordModel) withSession(session sqlx.Session) LoginRecordModel {
	return NewLoginRecordModel(sqlx.NewSqlConnFromSession(session))
}

// FindByUserId 根据用户ID查找登录记录
func (m *customLoginRecordModel) FindByUserId(ctx context.Context, userId string, page, pageSize int) ([]*LoginRecord, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `user_id` = ?", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, userId)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `user_id` = ? ORDER BY `login_time` DESC LIMIT ? OFFSET ?", loginRecordRows, m.table)
	var resp []*LoginRecord
	err = m.conn.QueryRowsCtx(ctx, &resp, query, userId, pageSize, offset)
	return resp, total, err
}

// FindByUserType 根据用户类型查找登录记录
func (m *customLoginRecordModel) FindByUserType(ctx context.Context, userType string, page, pageSize int) ([]*LoginRecord, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `user_type` = ?", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, userType)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `user_type` = ? ORDER BY `login_time` DESC LIMIT ? OFFSET ?", loginRecordRows, m.table)
	var resp []*LoginRecord
	err = m.conn.QueryRowsCtx(ctx, &resp, query, userType, pageSize, offset)
	return resp, total, err
}

// FindByStatus 根据登录状态查找登录记录
func (m *customLoginRecordModel) FindByStatus(ctx context.Context, status int, page, pageSize int) ([]*LoginRecord, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `login_status` = ?", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, status)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `login_status` = ? ORDER BY `login_time` DESC LIMIT ? OFFSET ?", loginRecordRows, m.table)
	var resp []*LoginRecord
	err = m.conn.QueryRowsCtx(ctx, &resp, query, status, pageSize, offset)
	return resp, total, err
}

// FindByTimeRange 根据时间范围查找登录记录
func (m *customLoginRecordModel) FindByTimeRange(ctx context.Context, startTime, endTime string, page, pageSize int) ([]*LoginRecord, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `login_time` >= ? AND `login_time` <= ?", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, startTime, endTime)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `login_time` >= ? AND `login_time` <= ? ORDER BY `login_time` DESC LIMIT ? OFFSET ?", loginRecordRows, m.table)
	var resp []*LoginRecord
	err = m.conn.QueryRowsCtx(ctx, &resp, query, startTime, endTime, pageSize, offset)
	return resp, total, err
}

// FindByPage 分页查找登录记录
func (m *customLoginRecordModel) FindByPage(ctx context.Context, page, pageSize int) ([]*LoginRecord, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY `login_time` DESC LIMIT ? OFFSET ?", loginRecordRows, m.table)
	var resp []*LoginRecord
	err = m.conn.QueryRowsCtx(ctx, &resp, query, pageSize, offset)
	return resp, total, err
}

// FindByFilters 根据多个条件筛选登录记录
func (m *customLoginRecordModel) FindByFilters(ctx context.Context, userId, userType string, status int, startTime, endTime string, page, pageSize int) ([]*LoginRecord, int64, error) {
	offset := (page - 1) * pageSize

	// 构建动态查询条件
	var conditions []string
	var args []interface{}

	if userId != "" {
		conditions = append(conditions, "`user_id` = ?")
		args = append(args, userId)
	}
	if userType != "" {
		conditions = append(conditions, "`user_type` = ?")
		args = append(args, userType)
	}
	if status >= 0 {
		conditions = append(conditions, "`login_status` = ?")
		args = append(args, status)
	}
	if startTime != "" {
		conditions = append(conditions, "`login_time` >= ?")
		args = append(args, startTime)
	}
	if endTime != "" {
		conditions = append(conditions, "`login_time` <= ?")
		args = append(args, endTime)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", m.table, whereClause)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	args = append(args, pageSize, offset)
	query := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY `login_time` DESC LIMIT ? OFFSET ?", loginRecordRows, m.table, whereClause)
	var resp []*LoginRecord
	err = m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, total, err
}

// GetRecentByUserId 获取用户最近的登录记录
func (m *customLoginRecordModel) GetRecentByUserId(ctx context.Context, userId string, limit int) ([]*LoginRecord, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `user_id` = ? ORDER BY `login_time` DESC LIMIT ?", loginRecordRows, m.table)
	var resp []*LoginRecord
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, limit)
	return resp, err
}

// GetLoginRecordCount 获取登录记录总数
func (m *customLoginRecordModel) GetLoginRecordCount(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

// GetFailedLoginCount 获取指定时间后的失败登录次数
func (m *customLoginRecordModel) GetFailedLoginCount(ctx context.Context, userId string, startTime string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `user_id` = ? AND `login_status` = 0 AND `login_time` >= ?", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, userId, startTime)
	return count, err
}
