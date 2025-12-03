package task

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TaskModel = (*customTaskModel)(nil)

type (
	// TaskModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTaskModel.
	TaskModel interface {
		taskModel
		withSession(session sqlx.Session) TaskModel

		// 任务CRUD操作
		FindByCompany(ctx context.Context, companyID string, page, pageSize int) ([]*Task, int64, error)
		FindByDepartment(ctx context.Context, departmentID string, page, pageSize int) ([]*Task, int64, error)
		FindByCreator(ctx context.Context, creatorID string, page, pageSize int) ([]*Task, int64, error)
		FindByStatus(ctx context.Context, status int) ([]*Task, error)
		FindByPriority(ctx context.Context, priority int) ([]*Task, error)
		FindByType(ctx context.Context, taskType string) ([]*Task, error)
		FindByPage(ctx context.Context, page, pageSize int) ([]*Task, int64, error)
		// 用户参与的任务（创建者/负责人/节点执行人）
		FindByInvolved(ctx context.Context, employeeID string, page, pageSize int) ([]*Task, int64, error)
		SearchTasks(ctx context.Context, keyword string, page, pageSize int) ([]*Task, int64, error)
		UpdateStatus(ctx context.Context, id string, status int) error
		UpdateProgress(ctx context.Context, id string, progress int) error
		UpdateActualHours(ctx context.Context, id string, actualHours int) error
		UpdateBasicInfo(ctx context.Context, id, title, description string) error
		UpdateDeadline(ctx context.Context, id string, deadline string) error
		SoftDelete(ctx context.Context, id string) error
		BatchUpdateStatus(ctx context.Context, ids []string, status int) error
		GetTaskCount(ctx context.Context) (int64, error)
		GetTaskCountByStatus(ctx context.Context, status int) (int64, error)
		GetTaskCountByCompany(ctx context.Context, companyID string) (int64, error)
		GetTaskCountByDepartment(ctx context.Context, departmentID string) (int64, error)
		GetTaskCountByCreator(ctx context.Context, creatorID string) (int64, error)
		// UpdateNodeCount 更新任务的节点统计数
		UpdateNodeCount(ctx context.Context, taskId string, totalCount, completedCount int64) error
	}

	customTaskModel struct {
		*defaultTaskModel
	}
)

// NewTaskModel returns a model for the database table.
func NewTaskModel(conn sqlx.SqlConn) TaskModel {
	return &customTaskModel{
		defaultTaskModel: newTaskModel(conn),
	}
}

func (m *customTaskModel) withSession(session sqlx.Session) TaskModel {
	return NewTaskModel(sqlx.NewSqlConnFromSession(session))
}

// FindByCompany 根据公司ID查找任务
func (m *customTaskModel) FindByCompany(ctx context.Context, companyID string, page, pageSize int) ([]*Task, int64, error) {
	var tasks []*Task
	var total int64

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM task WHERE company_id = ? AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, companyID)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `SELECT * FROM task WHERE company_id = ? AND delete_time IS NULL ORDER BY create_time DESC LIMIT ? OFFSET ?`
	err = m.conn.QueryRowsCtx(ctx, &tasks, query, companyID, pageSize, offset)
	return tasks, total, err
}

// FindByDepartment 根据部门ID查找任务
func (m *customTaskModel) FindByDepartment(ctx context.Context, departmentID string, page, pageSize int) ([]*Task, int64, error) {
	var tasks []*Task
	var total int64

	// 查询总数
	// 支持多部门存储（department_ids 为逗号分隔）
	countQuery := `SELECT COUNT(*) FROM task WHERE FIND_IN_SET(?, department_ids) AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, departmentID)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `SELECT * FROM task WHERE FIND_IN_SET(?, department_ids) AND delete_time IS NULL ORDER BY create_time DESC LIMIT ? OFFSET ?`
	err = m.conn.QueryRowsCtx(ctx, &tasks, query, departmentID, pageSize, offset)
	return tasks, total, err
}

// FindByInvolved 根据员工是否参与（创建者/负责人/节点执行人）查找任务
func (m *customTaskModel) FindByInvolved(ctx context.Context, employeeID string, page, pageSize int) ([]*Task, int64, error) {
	var tasks []*Task
	var total int64

	// 统计总数
	countQuery := `SELECT COUNT(*) FROM task 
        WHERE (task_creator = ? OR FIND_IN_SET(?, responsible_employee_ids) OR FIND_IN_SET(?, node_employee_ids))
        AND delete_time IS NULL`
	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, employeeID, employeeID, employeeID); err != nil {
		return nil, 0, err
	}

	// 分页
	offset := (page - 1) * pageSize
	query := `SELECT * FROM task 
        WHERE (task_creator = ? OR FIND_IN_SET(?, responsible_employee_ids) OR FIND_IN_SET(?, node_employee_ids))
        AND delete_time IS NULL ORDER BY create_time DESC LIMIT ? OFFSET ?`
	if err := m.conn.QueryRowsCtx(ctx, &tasks, query, employeeID, employeeID, employeeID, pageSize, offset); err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

// FindByCreator 根据创建者ID查找任务
func (m *customTaskModel) FindByCreator(ctx context.Context, creatorID string, page, pageSize int) ([]*Task, int64, error) {
	var tasks []*Task
	var total int64

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM task WHERE creator_id = ? AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, creatorID)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `SELECT * FROM task WHERE creator_id = ? AND delete_time IS NULL ORDER BY create_time DESC LIMIT ? OFFSET ?`
	err = m.conn.QueryRowsCtx(ctx, &tasks, query, creatorID, pageSize, offset)
	return tasks, total, err
}

// FindByStatus 根据状态查找任务
func (m *customTaskModel) FindByStatus(ctx context.Context, status int) ([]*Task, error) {
	var tasks []*Task
	query := `SELECT * FROM task WHERE status = ? AND delete_time IS NULL ORDER BY create_time DESC`
	err := m.conn.QueryRowsCtx(ctx, &tasks, query, status)
	return tasks, err
}

// FindByPriority 根据优先级查找任务
func (m *customTaskModel) FindByPriority(ctx context.Context, priority int) ([]*Task, error) {
	var tasks []*Task
	query := `SELECT * FROM task WHERE priority = ? AND delete_time IS NULL ORDER BY create_time DESC`
	err := m.conn.QueryRowsCtx(ctx, &tasks, query, priority)
	return tasks, err
}

// FindByType 根据任务类型查找任务
func (m *customTaskModel) FindByType(ctx context.Context, taskType string) ([]*Task, error) {
	var tasks []*Task
	query := `SELECT * FROM task WHERE task_type = ? AND delete_time IS NULL ORDER BY create_time DESC`
	err := m.conn.QueryRowsCtx(ctx, &tasks, query, taskType)
	return tasks, err
}

// FindByPage 分页查找任务
func (m *customTaskModel) FindByPage(ctx context.Context, page, pageSize int) ([]*Task, int64, error) {
	var tasks []*Task
	var total int64

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM task WHERE delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `SELECT * FROM task WHERE delete_time IS NULL ORDER BY create_time DESC LIMIT ? OFFSET ?`
	err = m.conn.QueryRowsCtx(ctx, &tasks, query, pageSize, offset)
	return tasks, total, err
}

// SearchTasks 搜索任务
func (m *customTaskModel) SearchTasks(ctx context.Context, keyword string, page, pageSize int) ([]*Task, int64, error) {
	var tasks []*Task
	var total int64

	// 构建搜索条件
	searchCondition := fmt.Sprintf("(task_title LIKE '%%%s%%' OR task_description LIKE '%%%s%%') AND delete_time IS NULL", keyword, keyword)

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM task WHERE %s", searchCondition)
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM task WHERE %s ORDER BY create_time DESC LIMIT ? OFFSET ?", searchCondition)
	err = m.conn.QueryRowsCtx(ctx, &tasks, query, pageSize, offset)
	return tasks, total, err
}

// UpdateStatus 更新任务状态
func (m *customTaskModel) UpdateStatus(ctx context.Context, id string, status int) error {
	query := `UPDATE task SET status = ?, update_time = NOW() WHERE id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

// UpdateProgress 更新任务进度
func (m *customTaskModel) UpdateProgress(ctx context.Context, id string, progress int) error {
	query := `UPDATE task SET progress = ?, update_time = NOW() WHERE id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, progress, id)
	return err
}

// UpdateActualHours 更新实际工时
func (m *customTaskModel) UpdateActualHours(ctx context.Context, id string, actualHours int) error {
	query := `UPDATE task SET actual_hours = ?, update_time = NOW() WHERE id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, actualHours, id)
	return err
}

// UpdateBasicInfo 更新任务基本信息
func (m *customTaskModel) UpdateBasicInfo(ctx context.Context, id, title, description string) error {
	query := `UPDATE task SET task_title = ?, task_description = ?, update_time = NOW() WHERE id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, title, description, id)
	return err
}

// UpdateDeadline 更新任务截止时间
func (m *customTaskModel) UpdateDeadline(ctx context.Context, id string, deadline string) error {
	query := `UPDATE task SET deadline = ?, update_time = NOW() WHERE id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, deadline, id)
	return err
}

// SoftDelete 软删除任务
func (m *customTaskModel) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE task SET delete_time = NOW() WHERE id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// BatchUpdateStatus 批量更新任务状态
func (m *customTaskModel) BatchUpdateStatus(ctx context.Context, ids []string, status int) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(ids)-1) + "?"
	query := fmt.Sprintf("UPDATE task SET status = ?, update_time = NOW() WHERE id IN (%s) AND delete_time IS NULL", placeholders)

	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, status)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

// GetTaskCount 获取任务总数
func (m *customTaskModel) GetTaskCount(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task WHERE delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

// GetTaskCountByStatus 根据状态获取任务数量
func (m *customTaskModel) GetTaskCountByStatus(ctx context.Context, status int) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task WHERE status = ? AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query, status)
	return count, err
}

// GetTaskCountByCompany 根据公司获取任务数量
func (m *customTaskModel) GetTaskCountByCompany(ctx context.Context, companyID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task WHERE company_id = ? AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query, companyID)
	return count, err
}

// GetTaskCountByDepartment 根据部门获取任务数量
func (m *customTaskModel) GetTaskCountByDepartment(ctx context.Context, departmentID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task WHERE department_id = ? AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query, departmentID)
	return count, err
}

// GetTaskCountByCreator 根据创建者获取任务数量
func (m *customTaskModel) GetTaskCountByCreator(ctx context.Context, creatorID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task WHERE creator_id = ? AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query, creatorID)
	return count, err
}

// UpdateNodeCount 更新任务的节点统计数
// 注意：需要先在数据库中添加 total_node_count 和 completed_node_count 字段
func (m *customTaskModel) UpdateNodeCount(ctx context.Context, taskId string, totalCount, completedCount int64) error {
	query := `UPDATE task SET total_node_count = ?, completed_node_count = ?, update_time = NOW() WHERE task_id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, totalCount, completedCount, taskId)
	return err
}
