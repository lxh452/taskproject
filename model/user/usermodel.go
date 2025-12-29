package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		withSession(session sqlx.Session) UserModel

		// 用户CRUD操作
		FindByUsername(ctx context.Context, username string) (*User, error)
		FindByEmail(ctx context.Context, email string) (*User, error)
		FindByPhone(ctx context.Context, phone string) (*User, error)
		FindByStatus(ctx context.Context, status int) ([]*User, error)
		FindByPage(ctx context.Context, page, pageSize int) ([]*User, int64, error)
		SearchUsers(ctx context.Context, keyword string, page, pageSize int) ([]*User, int64, error)
		UpdateLastLogin(ctx context.Context, id string, lastLoginTime, lastLoginIP string) error
		UpdateLoginFailedCount(ctx context.Context, id string, count int) error
		UpdateLockStatus(ctx context.Context, id string, lockedUntil string) error
		ClearLockStatus(ctx context.Context, id string) error
		UpdatePassword(ctx context.Context, id, passwordHash string) error
		UpdateProfile(ctx context.Context, id, realName, avatar string, gender int, birthday string) error
		UpdateStatus(ctx context.Context, id string, status int) error
		UpdateHasJoinedCompany(ctx context.Context, id string, hasJoined bool) error
		SoftDelete(ctx context.Context, id string) error
		BatchUpdateStatus(ctx context.Context, ids []string, status int) error
		GetUserCount(ctx context.Context) (int64, error)
		GetUserCountByStatus(ctx context.Context, status int) (int64, error)
	}

	customUserModel struct {
		*defaultUserModel
	}
)

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn),
	}
}

func (m *customUserModel) withSession(session sqlx.Session) UserModel {
	return NewUserModel(sqlx.NewSqlConnFromSession(session))
}

// FindByUsername 根据用户名查找用户
func (m *customUserModel) FindByUsername(ctx context.Context, username string) (*User, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `username` = ? AND `delete_time` IS NULL", userRows, m.table)
	var resp User
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

// FindByEmail 根据邮箱查找用户
func (m *customUserModel) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `email` = ? AND `delete_time` IS NULL", userRows, m.table)
	var resp User
	err := m.conn.QueryRowCtx(ctx, &resp, query, email)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByPhone 根据手机号查找用户
func (m *customUserModel) FindByPhone(ctx context.Context, phone string) (*User, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `phone` = ? AND `delete_time` IS NULL", userRows, m.table)
	var resp User
	err := m.conn.QueryRowCtx(ctx, &resp, query, phone)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByStatus 根据状态查找用户
func (m *customUserModel) FindByStatus(ctx context.Context, status int) ([]*User, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `status` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", userRows, m.table)
	var resp []*User
	err := m.conn.QueryRowsCtx(ctx, &resp, query, status)
	return resp, err
}

// FindByPage 分页查找用户
func (m *customUserModel) FindByPage(ctx context.Context, page, pageSize int) ([]*User, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", userRows, m.table)
	var resp []*User
	err = m.conn.QueryRowsCtx(ctx, &resp, query, pageSize, offset)
	return resp, total, err
}

// SearchUsers 搜索用户
func (m *customUserModel) SearchUsers(ctx context.Context, keyword string, page, pageSize int) ([]*User, int64, error) {
	offset := (page - 1) * pageSize
	keyword = "%" + keyword + "%"

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE (`username` LIKE ? OR `real_name` LIKE ? OR `email` LIKE ? OR `phone` LIKE ?) AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, keyword, keyword, keyword, keyword)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE (`username` LIKE ? OR `real_name` LIKE ? OR `email` LIKE ? OR `phone` LIKE ?) AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", userRows, m.table)
	var resp []*User
	err = m.conn.QueryRowsCtx(ctx, &resp, query, keyword, keyword, keyword, keyword, pageSize, offset)
	return resp, total, err
}

// UpdateLastLogin 更新最后登录信息
func (m *customUserModel) UpdateLastLogin(ctx context.Context, id string, lastLoginTime, lastLoginIP string) error {
	query := fmt.Sprintf("UPDATE %s SET `last_login_time` = ?, `last_login_ip` = ?, `login_failed_count` = 0, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, lastLoginTime, lastLoginIP, id)
	return err
}

// UpdateLoginFailedCount 更新登录失败次数
func (m *customUserModel) UpdateLoginFailedCount(ctx context.Context, id string, count int) error {
	query := fmt.Sprintf("UPDATE %s SET `login_failed_count` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, count, id)
	return err
}

// UpdateLockStatus 更新锁定状态
func (m *customUserModel) UpdateLockStatus(ctx context.Context, id string, lockedUntil string) error {
	query := fmt.Sprintf("UPDATE %s SET `locked_until` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, lockedUntil, id)
	return err
}

// ClearLockStatus 清除锁定状态（自动解锁）
func (m *customUserModel) ClearLockStatus(ctx context.Context, id string) error {
	query := fmt.Sprintf("UPDATE %s SET `locked_until` = NULL, `login_failed_count` = 0, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// UpdatePassword 更新密码
func (m *customUserModel) UpdatePassword(ctx context.Context, id, passwordHash string) error {
	query := fmt.Sprintf("UPDATE %s SET `password_hash` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, passwordHash, id)
	return err
}

// UpdateProfile 更新用户资料
func (m *customUserModel) UpdateProfile(ctx context.Context, id, realName, avatar string, gender int, birthday string) error {
	query := fmt.Sprintf("UPDATE %s SET `real_name` = ?, `avatar` = ?, `gender` = ?, `birthday` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, realName, avatar, gender, birthday, id)
	return err
}

// UpdateStatus 更新用户状态
func (m *customUserModel) UpdateStatus(ctx context.Context, id string, status int) error {
	query := fmt.Sprintf("UPDATE %s SET `status` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

// SoftDelete 软删除用户
func (m *customUserModel) SoftDelete(ctx context.Context, id string) error {
	query := fmt.Sprintf("UPDATE %s SET `delete_time` = NOW(), `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// BatchUpdateStatus 批量更新用户状态
func (m *customUserModel) BatchUpdateStatus(ctx context.Context, ids []string, status int) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(ids)-1) + "?"
	query := fmt.Sprintf("UPDATE %s SET `status` = ?, `update_time` = NOW() WHERE `id` IN (%s)", m.table, placeholders)

	args := make([]interface{}, len(ids)+1)
	args[0] = status
	for i, id := range ids {
		args[i+1] = id
	}

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

// GetUserCount 获取用户总数
func (m *customUserModel) GetUserCount(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

// GetUserCountByStatus 根据状态获取用户数量
func (m *customUserModel) GetUserCountByStatus(ctx context.Context, status int) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `status` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, status)
	return count, err
}

// UpdateHasJoinedCompany 更新用户是否已加入公司
func (m *customUserModel) UpdateHasJoinedCompany(ctx context.Context, id string, hasJoined bool) error {
	value := 0
	if hasJoined {
		value = 1
	}
	query := fmt.Sprintf("UPDATE %s SET `has_joined_company` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, value, id)
	return err
}
