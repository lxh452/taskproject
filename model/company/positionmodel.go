package company

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PositionModel = (*customPositionModel)(nil)

type (
	// PositionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPositionModel.
	PositionModel interface {
		positionModel
		withSession(session sqlx.Session) PositionModel

		// 职位CRUD操作
		FindByDepartmentID(ctx context.Context, departmentID string) ([]*Position, error)
		FindByStatus(ctx context.Context, status int) ([]*Position, error)
		FindByManagement(ctx context.Context, isManagement int) ([]*Position, error)
		FindByLevel(ctx context.Context, level int) ([]*Position, error)
		FindByPage(ctx context.Context, page, pageSize int) ([]*Position, int64, error)
		FindByDepartmentPage(ctx context.Context, departmentID string, page, pageSize int) ([]*Position, int64, error)
		SearchPositions(ctx context.Context, keyword string, page, pageSize int) ([]*Position, int64, error)
		SearchPositionsByDepartment(ctx context.Context, departmentID, keyword string, page, pageSize int) ([]*Position, int64, error)
		UpdateBasicInfo(ctx context.Context, id, positionName, positionCode, jobDescription, responsibilities, requirements string) error
		UpdateLevel(ctx context.Context, id string, level int) error
		UpdateSalaryRange(ctx context.Context, id string, minSalary, maxSalary int) error
		UpdateManagement(ctx context.Context, id string, isManagement int) error
		UpdateMaxEmployees(ctx context.Context, id string, maxEmployees int) error
		UpdateCurrentEmployees(ctx context.Context, id string, currentEmployees int) error
		UpdateStatus(ctx context.Context, id string, status int) error
		SoftDelete(ctx context.Context, id string) error
		BatchUpdateStatus(ctx context.Context, ids []string, status int) error
		BatchUpdateDepartment(ctx context.Context, ids []string, departmentID string) error
		GetPositionCount(ctx context.Context) (int64, error)
		GetPositionCountByDepartment(ctx context.Context, departmentID string) (int64, error)
		GetPositionCountByStatus(ctx context.Context, status int) (int64, error)
		GetPositionCountByManagement(ctx context.Context, isManagement int) (int64, error)
		GetPositionsBySalaryRange(ctx context.Context, minSalary, maxSalary int) ([]*Position, error)
		GetPositionsBySkills(ctx context.Context, skills string) ([]*Position, error)
	}

	customPositionModel struct {
		*defaultPositionModel
	}
)

// NewPositionModel returns a model for the database table.
func NewPositionModel(conn sqlx.SqlConn) PositionModel {
	return &customPositionModel{
		defaultPositionModel: newPositionModel(conn),
	}
}

func (m *customPositionModel) withSession(session sqlx.Session) PositionModel {
	return NewPositionModel(sqlx.NewSqlConnFromSession(session))
}

// FindByDepartmentID 根据部门ID查找职位
func (m *customPositionModel) FindByDepartmentID(ctx context.Context, departmentID string) ([]*Position, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `department_id` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", positionRows, m.table)
	var resp []*Position
	err := m.conn.QueryRowsCtx(ctx, &resp, query, departmentID)
	return resp, err
}

// FindByStatus 根据状态查找职位
func (m *customPositionModel) FindByStatus(ctx context.Context, status int) ([]*Position, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `status` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", positionRows, m.table)
	var resp []*Position
	err := m.conn.QueryRowsCtx(ctx, &resp, query, status)
	return resp, err
}

// FindByManagement 根据是否管理职位查找
func (m *customPositionModel) FindByManagement(ctx context.Context, isManagement int) ([]*Position, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `is_management` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", positionRows, m.table)
	var resp []*Position
	err := m.conn.QueryRowsCtx(ctx, &resp, query, isManagement)
	return resp, err
}

// FindByLevel 根据职位级别查找
func (m *customPositionModel) FindByLevel(ctx context.Context, level int) ([]*Position, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `position_level` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", positionRows, m.table)
	var resp []*Position
	err := m.conn.QueryRowsCtx(ctx, &resp, query, level)
	return resp, err
}

// FindByPage 分页查找职位
func (m *customPositionModel) FindByPage(ctx context.Context, page, pageSize int) ([]*Position, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", positionRows, m.table)
	var resp []*Position
	err = m.conn.QueryRowsCtx(ctx, &resp, query, pageSize, offset)
	return resp, total, err
}

// FindByDepartmentPage 根据部门分页查找职位
func (m *customPositionModel) FindByDepartmentPage(ctx context.Context, departmentID string, page, pageSize int) ([]*Position, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `department_id` = ? AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, departmentID)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `department_id` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", positionRows, m.table)
	var resp []*Position
	err = m.conn.QueryRowsCtx(ctx, &resp, query, departmentID, pageSize, offset)
	return resp, total, err
}

// SearchPositions 搜索职位
func (m *customPositionModel) SearchPositions(ctx context.Context, keyword string, page, pageSize int) ([]*Position, int64, error) {
	offset := (page - 1) * pageSize
	keyword = "%" + keyword + "%"

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE (`position_name` LIKE ? OR `position_code` LIKE ? OR `job_description` LIKE ? OR `responsibilities` LIKE ? OR `requirements` LIKE ?) AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, keyword, keyword, keyword, keyword, keyword)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE (`position_name` LIKE ? OR `position_code` LIKE ? OR `job_description` LIKE ? OR `responsibilities` LIKE ? OR `requirements` LIKE ?) AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", positionRows, m.table)
	var resp []*Position
	err = m.conn.QueryRowsCtx(ctx, &resp, query, keyword, keyword, keyword, keyword, keyword, pageSize, offset)
	return resp, total, err
}

// SearchPositionsByDepartment 根据部门搜索职位
func (m *customPositionModel) SearchPositionsByDepartment(ctx context.Context, departmentID, keyword string, page, pageSize int) ([]*Position, int64, error) {
	offset := (page - 1) * pageSize
	keyword = "%" + keyword + "%"

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `department_id` = ? AND (`position_name` LIKE ? OR `position_code` LIKE ? OR `job_description` LIKE ? OR `responsibilities` LIKE ? OR `requirements` LIKE ?) AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, departmentID, keyword, keyword, keyword, keyword, keyword)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `department_id` = ? AND (`position_name` LIKE ? OR `position_code` LIKE ? OR `job_description` LIKE ? OR `responsibilities` LIKE ? OR `requirements` LIKE ?) AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", positionRows, m.table)
	var resp []*Position
	err = m.conn.QueryRowsCtx(ctx, &resp, query, departmentID, keyword, keyword, keyword, keyword, keyword, pageSize, offset)
	return resp, total, err
}

// UpdateBasicInfo 更新职位基本信息
func (m *customPositionModel) UpdateBasicInfo(ctx context.Context, id, positionName, positionCode, jobDescription, responsibilities, requirements string) error {
	query := fmt.Sprintf("UPDATE %s SET `position_name` = ?, `position_code` = ?, `job_description` = ?, `responsibilities` = ?, `requirements` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, positionName, positionCode, jobDescription, responsibilities, requirements, id)
	return err
}

// UpdateLevel 更新职位级别
func (m *customPositionModel) UpdateLevel(ctx context.Context, id string, level int) error {
	query := fmt.Sprintf("UPDATE %s SET `position_level` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, level, id)
	return err
}

// UpdateSalaryRange 更新薪资范围
func (m *customPositionModel) UpdateSalaryRange(ctx context.Context, id string, minSalary, maxSalary int) error {
	query := fmt.Sprintf("UPDATE %s SET `salary_range_min` = ?, `salary_range_max` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, minSalary, maxSalary, id)
	return err
}

// UpdateManagement 更新是否管理职位
func (m *customPositionModel) UpdateManagement(ctx context.Context, id string, isManagement int) error {
	query := fmt.Sprintf("UPDATE %s SET `is_management` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, isManagement, id)
	return err
}

// UpdateMaxEmployees 更新最大员工数
func (m *customPositionModel) UpdateMaxEmployees(ctx context.Context, id string, maxEmployees int) error {
	query := fmt.Sprintf("UPDATE %s SET `max_employees` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, maxEmployees, id)
	return err
}

// UpdateCurrentEmployees 更新当前员工数
func (m *customPositionModel) UpdateCurrentEmployees(ctx context.Context, id string, currentEmployees int) error {
	query := fmt.Sprintf("UPDATE %s SET `current_employees` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, currentEmployees, id)
	return err
}

// UpdateStatus 更新职位状态
func (m *customPositionModel) UpdateStatus(ctx context.Context, id string, status int) error {
	query := fmt.Sprintf("UPDATE %s SET `status` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

// SoftDelete 软删除职位
func (m *customPositionModel) SoftDelete(ctx context.Context, id string) error {
	query := fmt.Sprintf("UPDATE %s SET `delete_time` = NOW(), `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// BatchUpdateStatus 批量更新职位状态
func (m *customPositionModel) BatchUpdateStatus(ctx context.Context, ids []string, status int) error {
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

// BatchUpdateDepartment 批量更新职位部门
func (m *customPositionModel) BatchUpdateDepartment(ctx context.Context, ids []string, departmentID string) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(ids)-1) + "?"
	query := fmt.Sprintf("UPDATE %s SET `department_id` = ?, `update_time` = NOW() WHERE `id` IN (%s)", m.table, placeholders)

	args := make([]interface{}, len(ids)+1)
	args[0] = departmentID
	for i, id := range ids {
		args[i+1] = id
	}

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

// GetPositionCount 获取职位总数
func (m *customPositionModel) GetPositionCount(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

// GetPositionCountByDepartment 根据部门获取职位数量
func (m *customPositionModel) GetPositionCountByDepartment(ctx context.Context, departmentID string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `department_id` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, departmentID)
	return count, err
}

// GetPositionCountByStatus 根据状态获取职位数量
func (m *customPositionModel) GetPositionCountByStatus(ctx context.Context, status int) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `status` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, status)
	return count, err
}

// GetPositionCountByManagement 根据是否管理职位获取数量
func (m *customPositionModel) GetPositionCountByManagement(ctx context.Context, isManagement int) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `is_management` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, isManagement)
	return count, err
}

// GetPositionsBySalaryRange 根据薪资范围查找职位
func (m *customPositionModel) GetPositionsBySalaryRange(ctx context.Context, minSalary, maxSalary int) ([]*Position, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `salary_range_min` >= ? AND `salary_range_max` <= ? AND `delete_time` IS NULL ORDER BY `salary_range_min` ASC", positionRows, m.table)
	var resp []*Position
	err := m.conn.QueryRowsCtx(ctx, &resp, query, minSalary, maxSalary)
	return resp, err
}

// GetPositionsBySkills 根据技能查找职位
func (m *customPositionModel) GetPositionsBySkills(ctx context.Context, skills string) ([]*Position, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `required_skills` LIKE ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", positionRows, m.table)
	var resp []*Position
	err := m.conn.QueryRowsCtx(ctx, &resp, query, "%"+skills+"%")
	return resp, err
}
