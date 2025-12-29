package task

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TaskHandoverModel = (*customTaskHandoverModel)(nil)

type (
	// TaskHandoverModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTaskHandoverModel.
	TaskHandoverModel interface {
		taskHandoverModel
		withSession(session sqlx.Session) TaskHandoverModel

		// 任务交接CRUD操作
		FindByTaskID(ctx context.Context, taskID string) ([]*TaskHandover, error)
		FindByTaskNodeID(ctx context.Context, taskNodeID string) ([]*TaskHandover, error)
		FindByFromEmployee(ctx context.Context, fromEmployeeID string, page, pageSize int) ([]*TaskHandover, int64, error)
		FindByToEmployee(ctx context.Context, toEmployeeID string, page, pageSize int) ([]*TaskHandover, int64, error)
		FindByEmployeeInvolved(ctx context.Context, employeeID string, page, pageSize int) ([]*TaskHandover, int64, error)
		FindByStatus(ctx context.Context, status int) ([]*TaskHandover, error)
		FindByPage(ctx context.Context, page, pageSize int) ([]*TaskHandover, int64, error)
		SearchTaskHandovers(ctx context.Context, keyword string, page, pageSize int) ([]*TaskHandover, int64, error)
		UpdateStatus(ctx context.Context, id string, status int) error
		UpdateApproval(ctx context.Context, id string, approved int, approvalNote string) error
		SoftDelete(ctx context.Context, id string) error
		BatchUpdateStatus(ctx context.Context, ids []string, status int) error
		GetTaskHandoverCount(ctx context.Context) (int64, error)
		GetTaskHandoverCountByStatus(ctx context.Context, status int) (int64, error)
		GetTaskHandoverCountByTask(ctx context.Context, taskID string) (int64, error)
		GetTaskHandoverCountByTaskNode(ctx context.Context, taskNodeID string) (int64, error)
		GetTaskHandoverCountByFromEmployee(ctx context.Context, fromEmployeeID string) (int64, error)
		GetTaskHandoverCountByToEmployee(ctx context.Context, toEmployeeID string) (int64, error)
	}

	customTaskHandoverModel struct {
		*defaultTaskHandoverModel
	}
)

// NewTaskHandoverModel returns a model for the database table.
func NewTaskHandoverModel(conn sqlx.SqlConn) TaskHandoverModel {
	return &customTaskHandoverModel{
		defaultTaskHandoverModel: newTaskHandoverModel(conn),
	}
}

func (m *customTaskHandoverModel) withSession(session sqlx.Session) TaskHandoverModel {
	return NewTaskHandoverModel(sqlx.NewSqlConnFromSession(session))
}

// FindByTaskID 根据任务ID查找任务交接
func (m *customTaskHandoverModel) FindByTaskID(ctx context.Context, taskID string) ([]*TaskHandover, error) {
	var taskHandovers []*TaskHandover
	query := `SELECT * FROM task_handover WHERE task_id = ? ORDER BY create_time DESC`
	err := m.conn.QueryRowsCtx(ctx, &taskHandovers, query, taskID)
	return taskHandovers, err
}

// FindByTaskNodeID 根据任务节点ID查找任务交接
func (m *customTaskHandoverModel) FindByTaskNodeID(ctx context.Context, taskNodeID string) ([]*TaskHandover, error) {
	var taskHandovers []*TaskHandover
	query := `SELECT * FROM task_handover WHERE task_node_id = ? ORDER BY create_time DESC`
	err := m.conn.QueryRowsCtx(ctx, &taskHandovers, query, taskNodeID)
	return taskHandovers, err
}

// FindByFromEmployee 根据交接人ID查找任务交接
func (m *customTaskHandoverModel) FindByFromEmployee(ctx context.Context, fromEmployeeID string, page, pageSize int) ([]*TaskHandover, int64, error) {
	var taskHandovers []*TaskHandover
	var total int64

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM task_handover WHERE from_employee_id = ?`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, fromEmployeeID)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `SELECT * FROM task_handover WHERE from_employee_id = ? ORDER BY create_time DESC LIMIT ? OFFSET ?`
	err = m.conn.QueryRowsCtx(ctx, &taskHandovers, query, fromEmployeeID, pageSize, offset)
	return taskHandovers, total, err
}

// FindByToEmployee 根据接收人ID查找任务交接
func (m *customTaskHandoverModel) FindByToEmployee(ctx context.Context, toEmployeeID string, page, pageSize int) ([]*TaskHandover, int64, error) {
	var taskHandovers []*TaskHandover
	var total int64

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM task_handover WHERE to_employee_id = ?`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, toEmployeeID)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `SELECT * FROM task_handover WHERE to_employee_id = ? ORDER BY create_time DESC LIMIT ? OFFSET ?`
	err = m.conn.QueryRowsCtx(ctx, &taskHandovers, query, toEmployeeID, pageSize, offset)
	return taskHandovers, total, err
}

// FindByStatus 根据状态查找任务交接
func (m *customTaskHandoverModel) FindByStatus(ctx context.Context, status int) ([]*TaskHandover, error) {
	var taskHandovers []*TaskHandover
	query := `SELECT * FROM task_handover WHERE handover_status = ? ORDER BY create_time DESC`
	err := m.conn.QueryRowsCtx(ctx, &taskHandovers, query, status)
	return taskHandovers, err
}

// FindByPage 分页查找任务交接
func (m *customTaskHandoverModel) FindByPage(ctx context.Context, page, pageSize int) ([]*TaskHandover, int64, error) {
	var taskHandovers []*TaskHandover
	var total int64

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM task_handover`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `SELECT * FROM task_handover ORDER BY create_time DESC LIMIT ? OFFSET ?`
	err = m.conn.QueryRowsCtx(ctx, &taskHandovers, query, pageSize, offset)
	return taskHandovers, total, err
}

// SearchTaskHandovers 搜索任务交接
func (m *customTaskHandoverModel) SearchTaskHandovers(ctx context.Context, keyword string, page, pageSize int) ([]*TaskHandover, int64, error) {
	var taskHandovers []*TaskHandover
	var total int64

	// 构建搜索条件
	searchCondition := fmt.Sprintf("(handover_reason LIKE '%%%s%%' OR handover_note LIKE '%%%s%%')", keyword, keyword)

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM task_handover WHERE %s", searchCondition)
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM task_handover WHERE %s ORDER BY create_time DESC LIMIT ? OFFSET ?", searchCondition)
	err = m.conn.QueryRowsCtx(ctx, &taskHandovers, query, pageSize, offset)
	return taskHandovers, total, err
}

// UpdateStatus 更新任务交接状态
func (m *customTaskHandoverModel) UpdateStatus(ctx context.Context, id string, status int) error {
	query := `UPDATE task_handover SET handover_status = ?, update_time = NOW() WHERE handover_id = ?`
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

// UpdateApproval 更新任务交接审批
func (m *customTaskHandoverModel) UpdateApproval(ctx context.Context, id string, approved int, approvalNote string) error {
	query := `UPDATE task_handover SET handover_status = ?, handover_note = ?, update_time = NOW() WHERE handover_id = ?`
	_, err := m.conn.ExecCtx(ctx, query, approved, approvalNote, id)
	return err
}

// SoftDelete 软删除任务交接（由于表没有delete_time字段，直接删除）
func (m *customTaskHandoverModel) SoftDelete(ctx context.Context, id string) error {
	query := `DELETE FROM task_handover WHERE handover_id = ?`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// BatchUpdateStatus 批量更新任务交接状态
func (m *customTaskHandoverModel) BatchUpdateStatus(ctx context.Context, ids []string, status int) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(ids)-1) + "?"
	query := fmt.Sprintf("UPDATE task_handover SET handover_status = ?, update_time = NOW() WHERE handover_id IN (%s)", placeholders)

	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, status)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

// GetTaskHandoverCount 获取任务交接总数
func (m *customTaskHandoverModel) GetTaskHandoverCount(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_handover`
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

// GetTaskHandoverCountByStatus 根据状态获取任务交接数量
func (m *customTaskHandoverModel) GetTaskHandoverCountByStatus(ctx context.Context, status int) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_handover WHERE handover_status = ?`
	err := m.conn.QueryRowCtx(ctx, &count, query, status)
	return count, err
}

// GetTaskHandoverCountByTask 根据任务获取任务交接数量
func (m *customTaskHandoverModel) GetTaskHandoverCountByTask(ctx context.Context, taskID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_handover WHERE task_id = ?`
	err := m.conn.QueryRowCtx(ctx, &count, query, taskID)
	return count, err
}

// GetTaskHandoverCountByTaskNode 根据任务节点获取任务交接数量
func (m *customTaskHandoverModel) GetTaskHandoverCountByTaskNode(ctx context.Context, taskNodeID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_handover WHERE task_node_id = ?`
	err := m.conn.QueryRowCtx(ctx, &count, query, taskNodeID)
	return count, err
}

// GetTaskHandoverCountByFromEmployee 根据交接人获取任务交接数量
func (m *customTaskHandoverModel) GetTaskHandoverCountByFromEmployee(ctx context.Context, fromEmployeeID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_handover WHERE from_employee_id = ?`
	err := m.conn.QueryRowCtx(ctx, &count, query, fromEmployeeID)
	return count, err
}

// GetTaskHandoverCountByToEmployee 根据接收人获取任务交接数量
func (m *customTaskHandoverModel) GetTaskHandoverCountByToEmployee(ctx context.Context, toEmployeeID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_handover WHERE to_employee_id = ?`
	err := m.conn.QueryRowCtx(ctx, &count, query, toEmployeeID)
	return count, err
}

// FindByEmployeeInvolved 查询与员工相关的所有交接（作为发起人或接收人或审批人）
func (m *customTaskHandoverModel) FindByEmployeeInvolved(ctx context.Context, employeeID string, page, pageSize int) ([]*TaskHandover, int64, error) {
	var taskHandovers []*TaskHandover
	var total int64

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM task_handover WHERE from_employee_id = ? OR to_employee_id = ? OR approver_id = ?`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, employeeID, employeeID, employeeID)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `SELECT * FROM task_handover WHERE from_employee_id = ? OR to_employee_id = ? OR approver_id = ? ORDER BY create_time DESC LIMIT ? OFFSET ?`
	err = m.conn.QueryRowsCtx(ctx, &taskHandovers, query, employeeID, employeeID, employeeID, pageSize, offset)
	return taskHandovers, total, err
}
