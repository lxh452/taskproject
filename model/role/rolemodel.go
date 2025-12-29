package role

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ RoleModel = (*customRoleModel)(nil)

type (
	// RoleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRoleModel.
	RoleModel interface {
		roleModel
		withSession(session sqlx.Session) RoleModel

		// 角色CRUD操作
		FindByCompanyID(ctx context.Context, companyID string) ([]*Role, error)
		FindByStatus(ctx context.Context, status int) ([]*Role, error)
		FindBySystem(ctx context.Context, isSystem int) ([]*Role, error)
		FindByRoleCode(ctx context.Context, companyID, roleCode string) (*Role, error)
		FindByPage(ctx context.Context, page, pageSize int) ([]*Role, int64, error)
		FindByCompanyPage(ctx context.Context, companyID string, page, pageSize int) ([]*Role, int64, error)
		SearchRoles(ctx context.Context, keyword string, page, pageSize int) ([]*Role, int64, error)
		SearchRolesByCompany(ctx context.Context, companyID, keyword string, page, pageSize int) ([]*Role, int64, error)
		UpdateBasicInfo(ctx context.Context, id, roleName, roleCode, roleDescription string) error
		UpdatePermissions(ctx context.Context, id, permissions string) error
		UpdateStatus(ctx context.Context, id string, status int) error
		SoftDelete(ctx context.Context, id string) error
		BatchUpdateStatus(ctx context.Context, ids []string, status int) error
		GetRoleCount(ctx context.Context) (int64, error)
		GetRoleCountByCompany(ctx context.Context, companyID string) (int64, error)
		GetRoleCountByStatus(ctx context.Context, status int) (int64, error)
		GetRoleCountBySystem(ctx context.Context, isSystem int) (int64, error)
		GetRolesByPermissions(ctx context.Context, permissions string) ([]*Role, error)
	}

	customRoleModel struct {
		*defaultRoleModel
	}
)

// NewRoleModel returns a model for the database table.
func NewRoleModel(conn sqlx.SqlConn) RoleModel {
	return &customRoleModel{
		defaultRoleModel: newRoleModel(conn),
	}
}

func (m *customRoleModel) withSession(session sqlx.Session) RoleModel {
	return NewRoleModel(sqlx.NewSqlConnFromSession(session))
}

// FindByCompanyID 根据公司ID查找角色
func (m *customRoleModel) FindByCompanyID(ctx context.Context, companyID string) ([]*Role, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `company_id` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", roleRows, m.table)
	var resp []*Role
	err := m.conn.QueryRowsCtx(ctx, &resp, query, companyID)
	return resp, err
}

// FindByStatus 根据状态查找角色
func (m *customRoleModel) FindByStatus(ctx context.Context, status int) ([]*Role, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `status` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", roleRows, m.table)
	var resp []*Role
	err := m.conn.QueryRowsCtx(ctx, &resp, query, status)
	return resp, err
}

// FindBySystem 根据是否系统角色查找
func (m *customRoleModel) FindBySystem(ctx context.Context, isSystem int) ([]*Role, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `is_system` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", roleRows, m.table)
	var resp []*Role
	err := m.conn.QueryRowsCtx(ctx, &resp, query, isSystem)
	return resp, err
}

// FindByRoleCode 根据角色编码查找角色
func (m *customRoleModel) FindByRoleCode(ctx context.Context, companyID, roleCode string) (*Role, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `company_id` = ? AND `role_code` = ? AND `delete_time` IS NULL", roleRows, m.table)
	var resp Role
	err := m.conn.QueryRowCtx(ctx, &resp, query, companyID, roleCode)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByPage 分页查找角色
func (m *customRoleModel) FindByPage(ctx context.Context, page, pageSize int) ([]*Role, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", roleRows, m.table)
	var resp []*Role
	err = m.conn.QueryRowsCtx(ctx, &resp, query, pageSize, offset)
	return resp, total, err
}

// FindByCompanyPage 根据公司分页查找角色
func (m *customRoleModel) FindByCompanyPage(ctx context.Context, companyID string, page, pageSize int) ([]*Role, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `company_id` = ? AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, companyID)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `company_id` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", roleRows, m.table)
	var resp []*Role
	err = m.conn.QueryRowsCtx(ctx, &resp, query, companyID, pageSize, offset)
	return resp, total, err
}

// SearchRoles 搜索角色
func (m *customRoleModel) SearchRoles(ctx context.Context, keyword string, page, pageSize int) ([]*Role, int64, error) {
	offset := (page - 1) * pageSize
	keyword = "%" + keyword + "%"

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE (`role_name` LIKE ? OR `role_code` LIKE ? OR `role_description` LIKE ?) AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, keyword, keyword, keyword)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE (`role_name` LIKE ? OR `role_code` LIKE ? OR `role_description` LIKE ?) AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", roleRows, m.table)
	var resp []*Role
	err = m.conn.QueryRowsCtx(ctx, &resp, query, keyword, keyword, keyword, pageSize, offset)
	return resp, total, err
}

// SearchRolesByCompany 根据公司搜索角色
func (m *customRoleModel) SearchRolesByCompany(ctx context.Context, companyID, keyword string, page, pageSize int) ([]*Role, int64, error) {
	offset := (page - 1) * pageSize
	keyword = "%" + keyword + "%"

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `company_id` = ? AND (`role_name` LIKE ? OR `role_code` LIKE ? OR `role_description` LIKE ?) AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, companyID, keyword, keyword, keyword)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `company_id` = ? AND (`role_name` LIKE ? OR `role_code` LIKE ? OR `role_description` LIKE ?) AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", roleRows, m.table)
	var resp []*Role
	err = m.conn.QueryRowsCtx(ctx, &resp, query, companyID, keyword, keyword, keyword, pageSize, offset)
	return resp, total, err
}

// UpdateBasicInfo 更新角色基本信息
func (m *customRoleModel) UpdateBasicInfo(ctx context.Context, id, roleName, roleCode, roleDescription string) error {
	query := fmt.Sprintf("UPDATE %s SET `role_name` = ?, `role_code` = ?, `role_description` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, roleName, roleCode, roleDescription, id)
	return err
}

// UpdatePermissions 更新角色权限
func (m *customRoleModel) UpdatePermissions(ctx context.Context, id, permissions string) error {
	query := fmt.Sprintf("UPDATE %s SET `permissions` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, permissions, id)
	return err
}

// UpdateStatus 更新角色状态
func (m *customRoleModel) UpdateStatus(ctx context.Context, id string, status int) error {
	query := fmt.Sprintf("UPDATE %s SET `status` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

// SoftDelete 软删除角色
func (m *customRoleModel) SoftDelete(ctx context.Context, id string) error {
	query := fmt.Sprintf("UPDATE %s SET `delete_time` = NOW(), `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// BatchUpdateStatus 批量更新角色状态
func (m *customRoleModel) BatchUpdateStatus(ctx context.Context, ids []string, status int) error {
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

// GetRoleCount 获取角色总数
func (m *customRoleModel) GetRoleCount(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

// GetRoleCountByCompany 根据公司获取角色数量
func (m *customRoleModel) GetRoleCountByCompany(ctx context.Context, companyID string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `company_id` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, companyID)
	return count, err
}

// GetRoleCountByStatus 根据状态获取角色数量
func (m *customRoleModel) GetRoleCountByStatus(ctx context.Context, status int) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `status` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, status)
	return count, err
}

// GetRoleCountBySystem 根据是否系统角色获取数量
func (m *customRoleModel) GetRoleCountBySystem(ctx context.Context, isSystem int) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `is_system` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, isSystem)
	return count, err
}

// GetRolesByPermissions 根据权限查找角色
func (m *customRoleModel) GetRolesByPermissions(ctx context.Context, permissions string) ([]*Role, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `permissions` LIKE ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", roleRows, m.table)
	var resp []*Role
	err := m.conn.QueryRowsCtx(ctx, &resp, query, "%"+permissions+"%")
	return resp, err
}
