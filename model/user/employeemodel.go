package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ EmployeeModel = (*customEmployeeModel)(nil)

type (
	// EmployeeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customEmployeeModel.
	EmployeeModel interface {
		employeeModel
		withSession(session sqlx.Session) EmployeeModel

		// 员工CRUD操作
		FindByUserID(ctx context.Context, userID string) (*Employee, error)
		FindByCompanyID(ctx context.Context, companyID string) ([]*Employee, error)
		FindByDepartmentID(ctx context.Context, departmentID string) ([]*Employee, error)
		FindByPositionID(ctx context.Context, positionID string) ([]*Employee, error)
		FindByEmployeeID(ctx context.Context, employeeID string) (*Employee, error)
		FindByStatus(ctx context.Context, status int) ([]*Employee, error)
		FindByPage(ctx context.Context, page, pageSize int) ([]*Employee, int64, error)
		FindByCompanyPage(ctx context.Context, companyID string, page, pageSize int) ([]*Employee, int64, error)
		FindByDepartmentPage(ctx context.Context, departmentID string, page, pageSize int) ([]*Employee, int64, error)
		SearchEmployees(ctx context.Context, keyword string, page, pageSize int) ([]*Employee, int64, error)
		SearchEmployeesByCompany(ctx context.Context, companyID, keyword string, page, pageSize int) ([]*Employee, int64, error)
		UpdateDepartment(ctx context.Context, id, departmentID string) error
		UpdatePosition(ctx context.Context, id, positionID string) error
		UpdateSkills(ctx context.Context, id, skills string) error
		UpdateRoleTags(ctx context.Context, id, roleTags string) error
		UpdateWorkContact(ctx context.Context, id, workEmail, workPhone string) error
		UpdateLeaveDate(ctx context.Context, id, leaveDate string) error
		UpdateStatus(ctx context.Context, id string, status int) error
		SoftDelete(ctx context.Context, id string) error
		BatchUpdateStatus(ctx context.Context, ids []string, status int) error
		BatchUpdateDepartment(ctx context.Context, ids []string, departmentID string) error
		GetEmployeeCount(ctx context.Context) (int64, error)
		GetEmployeeCountByCompany(ctx context.Context, companyID string) (int64, error)
		GetEmployeeCountByDepartment(ctx context.Context, departmentID string) (int64, error)
		GetEmployeeCountByStatus(ctx context.Context, status int) (int64, error)
		GetEmployeesByRoleTags(ctx context.Context, roleTags string) ([]*Employee, error)
		GetEmployeesBySkills(ctx context.Context, skills string) ([]*Employee, error)
	}

	customEmployeeModel struct {
		*defaultEmployeeModel
	}
)

// NewEmployeeModel returns a model for the database table.
func NewEmployeeModel(conn sqlx.SqlConn) EmployeeModel {
	return &customEmployeeModel{
		defaultEmployeeModel: newEmployeeModel(conn),
	}
}

func (m *customEmployeeModel) withSession(session sqlx.Session) EmployeeModel {
	return NewEmployeeModel(sqlx.NewSqlConnFromSession(session))
}

// FindByUserID 根据用户ID查找员工
func (m *customEmployeeModel) FindByUserID(ctx context.Context, userID string) (*Employee, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `user_id` = ? AND `delete_time` IS NULL", employeeRows, m.table)
	var resp Employee
	err := m.conn.QueryRowCtx(ctx, &resp, query, userID)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByCompanyID 根据公司ID查找员工
func (m *customEmployeeModel) FindByCompanyID(ctx context.Context, companyID string) ([]*Employee, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `company_id` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", employeeRows, m.table)
	var resp []*Employee
	err := m.conn.QueryRowsCtx(ctx, &resp, query, companyID)
	return resp, err
}

// FindByDepartmentID 根据部门ID查找员工
func (m *customEmployeeModel) FindByDepartmentID(ctx context.Context, departmentID string) ([]*Employee, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `department_id` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", employeeRows, m.table)
	var resp []*Employee
	err := m.conn.QueryRowsCtx(ctx, &resp, query, departmentID)
	return resp, err
}

// FindByPositionID 根据职位ID查找员工
func (m *customEmployeeModel) FindByPositionID(ctx context.Context, positionID string) ([]*Employee, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `position_id` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", employeeRows, m.table)
	var resp []*Employee
	err := m.conn.QueryRowsCtx(ctx, &resp, query, positionID)
	return resp, err
}

// FindByEmployeeID 根据员工编号查找员工
func (m *customEmployeeModel) FindByEmployeeID(ctx context.Context, employeeID string) (*Employee, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `employee_id` = ? AND `delete_time` IS NULL", employeeRows, m.table)
	var resp Employee
	err := m.conn.QueryRowCtx(ctx, &resp, query, employeeID)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByStatus 根据状态查找员工
func (m *customEmployeeModel) FindByStatus(ctx context.Context, status int) ([]*Employee, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `status` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", employeeRows, m.table)
	var resp []*Employee
	err := m.conn.QueryRowsCtx(ctx, &resp, query, status)
	return resp, err
}

// FindByPage 分页查找员工
func (m *customEmployeeModel) FindByPage(ctx context.Context, page, pageSize int) ([]*Employee, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", employeeRows, m.table)
	var resp []*Employee
	err = m.conn.QueryRowsCtx(ctx, &resp, query, pageSize, offset)
	return resp, total, err
}

// FindByCompanyPage 根据公司分页查找员工
func (m *customEmployeeModel) FindByCompanyPage(ctx context.Context, companyID string, page, pageSize int) ([]*Employee, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `company_id` = ? AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, companyID)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `company_id` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", employeeRows, m.table)
	var resp []*Employee
	err = m.conn.QueryRowsCtx(ctx, &resp, query, companyID, pageSize, offset)
	return resp, total, err
}

// FindByDepartmentPage 根据部门分页查找员工
func (m *customEmployeeModel) FindByDepartmentPage(ctx context.Context, departmentID string, page, pageSize int) ([]*Employee, int64, error) {
	offset := (page - 1) * pageSize

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `department_id` = ? AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, departmentID)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `department_id` = ? AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", employeeRows, m.table)
	var resp []*Employee
	err = m.conn.QueryRowsCtx(ctx, &resp, query, departmentID, pageSize, offset)
	return resp, total, err
}

// SearchEmployees 搜索员工
func (m *customEmployeeModel) SearchEmployees(ctx context.Context, keyword string, page, pageSize int) ([]*Employee, int64, error) {
	offset := (page - 1) * pageSize
	keyword = "%" + keyword + "%"

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE (`real_name` LIKE ? OR `employee_id` LIKE ? OR `work_email` LIKE ? OR `work_phone` LIKE ?) AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, keyword, keyword, keyword, keyword)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE (`real_name` LIKE ? OR `employee_id` LIKE ? OR `work_email` LIKE ? OR `work_phone` LIKE ?) AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", employeeRows, m.table)
	var resp []*Employee
	err = m.conn.QueryRowsCtx(ctx, &resp, query, keyword, keyword, keyword, keyword, pageSize, offset)
	return resp, total, err
}

// SearchEmployeesByCompany 根据公司搜索员工
func (m *customEmployeeModel) SearchEmployeesByCompany(ctx context.Context, companyID, keyword string, page, pageSize int) ([]*Employee, int64, error) {
	offset := (page - 1) * pageSize
	keyword = "%" + keyword + "%"

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `company_id` = ? AND (`real_name` LIKE ? OR `employee_id` LIKE ? OR `work_email` LIKE ? OR `work_phone` LIKE ?) AND `delete_time` IS NULL", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, companyID, keyword, keyword, keyword, keyword)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `company_id` = ? AND (`real_name` LIKE ? OR `employee_id` LIKE ? OR `work_email` LIKE ? OR `work_phone` LIKE ?) AND `delete_time` IS NULL ORDER BY `create_time` DESC LIMIT ? OFFSET ?", employeeRows, m.table)
	var resp []*Employee
	err = m.conn.QueryRowsCtx(ctx, &resp, query, companyID, keyword, keyword, keyword, keyword, pageSize, offset)
	return resp, total, err
}

// UpdateDepartment 更新部门
func (m *customEmployeeModel) UpdateDepartment(ctx context.Context, id, departmentID string) error {
	query := fmt.Sprintf("UPDATE %s SET `department_id` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, departmentID, id)
	return err
}

// UpdatePosition 更新职位
func (m *customEmployeeModel) UpdatePosition(ctx context.Context, id, positionID string) error {
	query := fmt.Sprintf("UPDATE %s SET `position_id` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, positionID, id)
	return err
}

// UpdateSkills 更新技能
func (m *customEmployeeModel) UpdateSkills(ctx context.Context, id, skills string) error {
	query := fmt.Sprintf("UPDATE %s SET `skills` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, skills, id)
	return err
}

// UpdateRoleTags 更新角色标签
func (m *customEmployeeModel) UpdateRoleTags(ctx context.Context, id, roleTags string) error {
	query := fmt.Sprintf("UPDATE %s SET `role_tags` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, roleTags, id)
	return err
}

// UpdateWorkContact 更新工作联系方式
func (m *customEmployeeModel) UpdateWorkContact(ctx context.Context, id, workEmail, workPhone string) error {
	query := fmt.Sprintf("UPDATE %s SET `work_email` = ?, `work_phone` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, workEmail, workPhone, id)
	return err
}

// UpdateLeaveDate 更新离职日期
func (m *customEmployeeModel) UpdateLeaveDate(ctx context.Context, id, leaveDate string) error {
	query := fmt.Sprintf("UPDATE %s SET `leave_date` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, leaveDate, id)
	return err
}

// UpdateStatus 更新员工状态
func (m *customEmployeeModel) UpdateStatus(ctx context.Context, id string, status int) error {
	query := fmt.Sprintf("UPDATE %s SET `status` = ?, `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

// SoftDelete 软删除员工
func (m *customEmployeeModel) SoftDelete(ctx context.Context, id string) error {
	query := fmt.Sprintf("UPDATE %s SET `delete_time` = NOW(), `update_time` = NOW() WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// BatchUpdateStatus 批量更新员工状态
func (m *customEmployeeModel) BatchUpdateStatus(ctx context.Context, ids []string, status int) error {
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

// BatchUpdateDepartment 批量更新员工部门
func (m *customEmployeeModel) BatchUpdateDepartment(ctx context.Context, ids []string, departmentID string) error {
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

// GetEmployeeCount 获取员工总数
func (m *customEmployeeModel) GetEmployeeCount(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

// GetEmployeeCountByCompany 根据公司获取员工数量
func (m *customEmployeeModel) GetEmployeeCountByCompany(ctx context.Context, companyID string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `company_id` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, companyID)
	return count, err
}

// GetEmployeeCountByDepartment 根据部门获取员工数量
func (m *customEmployeeModel) GetEmployeeCountByDepartment(ctx context.Context, departmentID string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `department_id` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, departmentID)
	return count, err
}

// GetEmployeeCountByStatus 根据状态获取员工数量
func (m *customEmployeeModel) GetEmployeeCountByStatus(ctx context.Context, status int) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE `status` = ? AND `delete_time` IS NULL", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, status)
	return count, err
}

// GetEmployeesByRoleTags 根据角色标签查找员工
func (m *customEmployeeModel) GetEmployeesByRoleTags(ctx context.Context, roleTags string) ([]*Employee, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `role_tags` LIKE ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", employeeRows, m.table)
	var resp []*Employee
	err := m.conn.QueryRowsCtx(ctx, &resp, query, "%"+roleTags+"%")
	return resp, err
}

// GetEmployeesBySkills 根据技能查找员工
func (m *customEmployeeModel) GetEmployeesBySkills(ctx context.Context, skills string) ([]*Employee, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `skills` LIKE ? AND `delete_time` IS NULL ORDER BY `create_time` DESC", employeeRows, m.table)
	var resp []*Employee
	err := m.conn.QueryRowsCtx(ctx, &resp, query, "%"+skills+"%")
	return resp, err
}
