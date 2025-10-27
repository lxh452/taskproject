package company

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ DepartmentModel = (*customDepartmentModel)(nil)

type (
	// DepartmentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customDepartmentModel.
	DepartmentModel interface {
		departmentModel
		withSession(session sqlx.Session) DepartmentModel

		// 部门CRUD操作
		FindByCompanyID(ctx context.Context, companyID string) ([]*Department, error)
		FindByParentID(ctx context.Context, parentID string) ([]*Department, error)
		FindByManagerID(ctx context.Context, managerID string) ([]*Department, error)
		FindByStatus(ctx context.Context, status int) ([]*Department, error)
		FindByPage(ctx context.Context, page, pageSize int) ([]*Department, int64, error)
		FindByCompanyPage(ctx context.Context, companyID string, page, pageSize int) ([]*Department, int64, error)
		SearchDepartments(ctx context.Context, keyword string, page, pageSize int) ([]*Department, int64, error)
		SearchDepartmentsByCompany(ctx context.Context, companyID, keyword string, page, pageSize int) ([]*Department, int64, error)
		UpdateManager(ctx context.Context, id, managerID string) error
		UpdateParent(ctx context.Context, id, parentID string) error
		UpdateBasicInfo(ctx context.Context, id, departmentName, departmentCode, description string) error
		UpdateStatus(ctx context.Context, id string, status int) error
		SoftDelete(ctx context.Context, id string) error
		BatchUpdateStatus(ctx context.Context, ids []string, status int) error
		BatchUpdateManager(ctx context.Context, ids []string, managerID string) error
		GetDepartmentCount(ctx context.Context) (int64, error)
		GetDepartmentCountByCompany(ctx context.Context, companyID string) (int64, error)
		GetDepartmentCountByStatus(ctx context.Context, status int) (int64, error)
		GetDepartmentCountByManager(ctx context.Context, managerID string) (int64, error)
		GetDepartmentTree(ctx context.Context, companyID string) ([]*Department, error)
		GetSubDepartments(ctx context.Context, parentID string) ([]*Department, error)
	}

	customDepartmentModel struct {
		*defaultDepartmentModel
	}
)

// NewDepartmentModel returns a model for the database table.
func NewDepartmentModel(conn sqlx.SqlConn) DepartmentModel {
	return &customDepartmentModel{
		defaultDepartmentModel: newDepartmentModel(conn),
	}
}

func (m *customDepartmentModel) withSession(session sqlx.Session) DepartmentModel {
	return NewDepartmentModel(sqlx.NewSqlConnFromSession(session))
}

// FindByCompanyID 根据公司ID查找部门
func (m *customDepartmentModel) FindByCompanyID(ctx context.Context, companyID string) ([]*Department, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `company_id` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", departmentRows, m.table)
	var resp []*Department
	err := m.conn.QueryRowsCtx(ctx, &resp, query, companyID)
	return resp, err
}

// FindByParentID 根据父部门ID查找部门
func (m *customDepartmentModel) FindByParentID(ctx context.Context, parentID string) ([]*Department, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `parent_id` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", departmentRows, m.table)
	var resp []*Department
	err := m.conn.QueryRowsCtx(ctx, &resp, query, parentID)
	return resp, err
}

// FindByManagerID 根据管理者ID查找部门
func (m *customDepartmentModel) FindByManagerID(ctx context.Context, managerID string) ([]*Department, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `manager_id` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", departmentRows, m.table)
	var resp []*Department
	err := m.conn.QueryRowsCtx(ctx, &resp, query, managerID)
	return resp, err
}

// FindByStatus 根据状态查找部门
func (m *customDepartmentModel) FindByStatus(ctx context.Context, status int) ([]*Department, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `status` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", departmentRows, m.table)
	var resp []*Department
	err := m.conn.QueryRowsCtx(ctx, &resp, query, status)
	return resp, err
}

// FindByPage 分页查找部门
func (m *customDepartmentModel) FindByPage(ctx context.Context, page, pageSize int) ([]*Department, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", departmentRows, m.table)
	var resp []*Department
	err = m.conn.QueryRowsCtx(ctx, &resp, query, pageSize, offset)
	return resp, total, err
}

// FindByCompanyPage 根据公司分页查找部门
func (m *customDepartmentModel) FindByCompanyPage(ctx context.Context, companyID string, page, pageSize int) ([]*Department, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `company_id` = ? AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, companyID)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `company_id` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", departmentRows, m.table)
	var resp []*Department
	err = m.conn.QueryRowsCtx(ctx, &resp, query, companyID, pageSize, offset)
	return resp, total, err
}

// SearchDepartments 搜索部门
func (m *customDepartmentModel) SearchDepartments(ctx context.Context, keyword string, page, pageSize int) ([]*Department, int64, error) {
	offset := (page - 1) * pageSize
	keyword = "%" + keyword + "%"

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE (`department_name` LIKE ? OR `department_code` LIKE ? OR `description` LIKE ?) AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, keyword, keyword, keyword)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE (`department_name` LIKE ? OR `department_code` LIKE ? OR `description` LIKE ?) AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", departmentRows, m.table)
	var resp []*Department
	err = m.conn.QueryRowsCtx(ctx, &resp, query, keyword, keyword, keyword, pageSize, offset)
	return resp, total, err
}

// SearchDepartmentsByCompany 根据公司搜索部门
func (m *customDepartmentModel) SearchDepartmentsByCompany(ctx context.Context, companyID, keyword string, page, pageSize int) ([]*Department, int64, error) {
	offset := (page - 1) * pageSize
	keyword = "%" + keyword + "%"

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `company_id` = ? AND (`department_name` LIKE ? OR `department_code` LIKE ? OR `description` LIKE ?) AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, companyID, keyword, keyword, keyword)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `company_id` = ? AND (`department_name` LIKE ? OR `department_code` LIKE ? OR `description` LIKE ?) AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", departmentRows, m.table)
	var resp []*Department
	err = m.conn.QueryRowsCtx(ctx, &resp, query, companyID, keyword, keyword, keyword, pageSize, offset)
	return resp, total, err
}

// UpdateManager 更新部门管理者
func (m *customDepartmentModel) UpdateManager(ctx context.Context, id, managerID string) error {
	query := fmt.Sprintf("UPDATE %s SET `manager_id` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, managerID, id)
	return err
}

// UpdateParent 更新父部门
func (m *customDepartmentModel) UpdateParent(ctx context.Context, id, parentID string) error {
	query := fmt.Sprintf("UPDATE %s SET `parent_id` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, parentID, id)
	return err
}

// UpdateBasicInfo 更新部门基本信息
func (m *customDepartmentModel) UpdateBasicInfo(ctx context.Context, id, departmentName, departmentCode, description string) error {
	query := fmt.Sprintf("UPDATE %s SET `department_name` = ?, `department_code` = ?, `description` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, departmentName, departmentCode, description, id)
	return err
}

// UpdateStatus 更新部门状态
func (m *customDepartmentModel) UpdateStatus(ctx context.Context, id string, status int) error {
	query := fmt.Sprintf("UPDATE %s SET `status` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

// SoftDelete 软删除部门
func (m *customDepartmentModel) SoftDelete(ctx context.Context, id string) error {
	query := fmt.Sprintf("UPDATE %s SET `delete_time` = NOW(), `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// BatchUpdateStatus 批量更新部门状态
func (m *customDepartmentModel) BatchUpdateStatus(ctx context.Context, ids []string, status int) error {
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

// BatchUpdateManager 批量更新部门管理者
func (m *customDepartmentModel) BatchUpdateManager(ctx context.Context, ids []string, managerID string) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(ids)-1) + "?"
	query := fmt.Sprintf("UPDATE %s SET `manager_id` = ?, `update_time` = NOW() WHERE `id` IN (%s)", m.table, placeholders)

	args := make([]interface{}, len(ids)+1)
	args[0] = managerID
	for i, id := range ids {
		args[i+1] = id
	}

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

// GetDepartmentCount 获取部门总数
func (m *customDepartmentModel) GetDepartmentCount(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

// GetDepartmentCountByCompany 根据公司获取部门数量
func (m *customDepartmentModel) GetDepartmentCountByCompany(ctx context.Context, companyID string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `company_id` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, companyID)
	return count, err
}

// GetDepartmentCountByStatus 根据状态获取部门数量
func (m *customDepartmentModel) GetDepartmentCountByStatus(ctx context.Context, status int) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `status` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, status)
	return count, err
}

// GetDepartmentCountByManager 根据管理者获取部门数量
func (m *customDepartmentModel) GetDepartmentCountByManager(ctx context.Context, managerID string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `manager_id` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, managerID)
	return count, err
}

// GetDepartmentTree 获取部门树形结构
func (m *customDepartmentModel) GetDepartmentTree(ctx context.Context, companyID string) ([]*Department, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `company_id` = ? AND `delete_time` IS NULL ORDER BY `parent_id`, `create_time`", departmentRows, m.table)
	var resp []*Department
	err := m.conn.QueryRowsCtx(ctx, &resp, query, companyID)
	return resp, err
}

// GetSubDepartments 获取子部门
func (m *customDepartmentModel) GetSubDepartments(ctx context.Context, parentID string) ([]*Department, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `parent_id` = ? AND `delete_time` IS NULL ORDER BY `create_time`", departmentRows, m.table)
	var resp []*Department
	err := m.conn.QueryRowsCtx(ctx, &resp, query, parentID)
	return resp, err
}
