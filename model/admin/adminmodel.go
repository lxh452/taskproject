package admin

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ AdminModel = (*customAdminModel)(nil)

type (
	// AdminModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAdminModel.
	AdminModel interface {
		adminModel
		withSession(session sqlx.Session) AdminModel

		// 管理员CRUD操作
		FindByUsername(ctx context.Context, username string) (*Admin, error)
		FindByStatus(ctx context.Context, status int) ([]*Admin, error)
		FindByRole(ctx context.Context, role string) ([]*Admin, error)
		FindByPage(ctx context.Context, page, pageSize int) ([]*Admin, int64, error)
		UpdateLastLogin(ctx context.Context, id string, lastLoginTime, lastLoginIP string) error
		UpdateStatus(ctx context.Context, id string, status int) error
		UpdatePassword(ctx context.Context, id, passwordHash string) error
		SoftDelete(ctx context.Context, id string) error
		GetAdminCount(ctx context.Context) (int64, error)
	}

	customAdminModel struct {
		*defaultAdminModel
	}
)

// NewAdminModel returns a model for the database table.
func NewAdminModel(conn sqlx.SqlConn) AdminModel {
	return &customAdminModel{
		defaultAdminModel: newAdminModel(conn),
	}
}

func (m *customAdminModel) withSession(session sqlx.Session) AdminModel {
	return NewAdminModel(sqlx.NewSqlConnFromSession(session))
}

// FindByUsername 根据用户名查找管理员
func (m *customAdminModel) FindByUsername(ctx context.Context, username string) (*Admin, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `username` = ? AND `delete_time` IS NULL", adminRows, m.table)
	var resp Admin
	err := m.conn.QueryRowCtx(ctx, &resp, query, username)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByStatus 根据状态查找管理员
func (m *customAdminModel) FindByStatus(ctx context.Context, status int) ([]*Admin, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `status` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", adminRows, m.table)
	var resp []*Admin
	err := m.conn.QueryRowsCtx(ctx, &resp, query, status)
	return resp, err
}

// FindByRole 根据角色查找管理员
func (m *customAdminModel) FindByRole(ctx context.Context, role string) ([]*Admin, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `role` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", adminRows, m.table)
	var resp []*Admin
	err := m.conn.QueryRowsCtx(ctx, &resp, query, role)
	return resp, err
}

// FindByPage 分页查找管理员
func (m *customAdminModel) FindByPage(ctx context.Context, page, pageSize int) ([]*Admin, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", adminRows, m.table)
	var resp []*Admin
	err = m.conn.QueryRowsCtx(ctx, &resp, query, pageSize, offset)
	return resp, total, err
}

// UpdateLastLogin 更新最后登录信息
func (m *customAdminModel) UpdateLastLogin(ctx context.Context, id string, lastLoginTime, lastLoginIP string) error {
	query := fmt.Sprintf("UPDATE %s SET `last_login_time` = ?, `last_login_ip` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, lastLoginTime, lastLoginIP, id)
	return err
}

// UpdateStatus 更新管理员状态
func (m *customAdminModel) UpdateStatus(ctx context.Context, id string, status int) error {
	query := fmt.Sprintf("UPDATE %s SET `status` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

// UpdatePassword 更新密码
func (m *customAdminModel) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	query := fmt.Sprintf("UPDATE %s SET `password_hash` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, passwordHash, id)
	return err
}

// SoftDelete 软删除管理员
func (m *customAdminModel) SoftDelete(ctx context.Context, id string) error {
	query := fmt.Sprintf("UPDATE %s SET `delete_time` = NOW(), `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// GetAdminCount 获取管理员总数
func (m *customAdminModel) GetAdminCount(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}
