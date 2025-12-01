package user_auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ NotificationModel = (*customNotificationModel)(nil)

type (
	// NotificationModel extends the generated interface with custom queries.
	NotificationModel interface {
		notificationModel
		withSession(session sqlx.Session) NotificationModel
		FindByEmployee(ctx context.Context, employeeID string, isRead *int, category *string, page, pageSize int) ([]*Notification, int64, error)
		UpdateReadStatus(ctx context.Context, id string, isRead int64) error
	}

	customNotificationModel struct {
		*defaultNotificationModel
	}
)

// NewNotificationModel returns a model for the database table.
func NewNotificationModel(conn sqlx.SqlConn) NotificationModel {
	return &customNotificationModel{
		defaultNotificationModel: newNotificationModel(conn),
	}
}

func (m *customNotificationModel) withSession(session sqlx.Session) NotificationModel {
	return NewNotificationModel(sqlx.NewSqlConnFromSession(session))
}

// FindByEmployee returns notifications for a given employee with optional filters.
func (m *customNotificationModel) FindByEmployee(ctx context.Context, employeeID string, isRead *int, category *string, page, pageSize int) ([]*Notification, int64, error) {
	offset := (page - 1) * pageSize

	whereConditions := []string{"`employee_id` = ?"}
	args := []interface{}{employeeID}

	if isRead != nil {
		whereConditions = append(whereConditions, "`is_read` = ?")
		args = append(args, *isRead)
	}

	if category != nil && *category != "" {
		whereConditions = append(whereConditions, "`category` = ?")
		args = append(args, *category)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", m.table, whereClause)
	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s ORDER BY `create_time` DESC LIMIT ? OFFSET ?", notificationRows, m.table, whereClause)
	queryArgs := append(args, pageSize, offset)
	var resp []*Notification
	if err := m.conn.QueryRowsCtx(ctx, &resp, query, queryArgs...); err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}

// UpdateReadStatus updates the read status and timestamp of a notification.
func (m *customNotificationModel) UpdateReadStatus(ctx context.Context, id string, isRead int64) error {
	query := fmt.Sprintf("UPDATE %s SET `is_read` = ?, `read_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, isRead, id)
	return err
}
