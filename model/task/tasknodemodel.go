package task

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TaskNodeModel = (*customTaskNodeModel)(nil)

type (
	// TaskNodeModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTaskNodeModel.
	TaskNodeModel interface {
		taskNodeModel
		withSession(session sqlx.Session) TaskNodeModel

		// 任务节点CRUD操作
		FindByTaskID(ctx context.Context, taskID string) ([]*TaskNode, error)
		FindByDepartment(ctx context.Context, departmentID string, page, pageSize int) ([]*TaskNode, int64, error)
		FindByExecutor(ctx context.Context, executorID string, page, pageSize int) ([]*TaskNode, int64, error)
		FindByLeader(ctx context.Context, leaderID string, page, pageSize int) ([]*TaskNode, int64, error)
		FindByTaskNodeIDLeader(ctx context.Context, taskNodeID, LeaderId string) (*TaskNode, error)
		FindByStatus(ctx context.Context, status int) ([]*TaskNode, error)
		FindByDeadlineRange(ctx context.Context, startTime, endTime string) ([]*TaskNode, error)
		FindByPage(ctx context.Context, page, pageSize int) ([]*TaskNode, int64, error)
		SearchTaskNodes(ctx context.Context, keyword string, page, pageSize int) ([]*TaskNode, int64, error)
		UpdateStatus(ctx context.Context, id string, status int) error
		UpdateProgress(ctx context.Context, id string, progress int) error
		UpdateActualHours(ctx context.Context, id string, actualHours int) error
		UpdateExecutor(ctx context.Context, id, executorID string) error
		UpdateLeader(ctx context.Context, id, leaderID string) error
		UpdateDeadline(ctx context.Context, id, deadline string) error
		SoftDelete(ctx context.Context, id string) error
		BatchUpdateStatus(ctx context.Context, ids []string, status int) error
		GetTaskNodeCount(ctx context.Context) (int64, error)
		GetTaskNodeCountByStatus(ctx context.Context, status int) (int64, error)
		GetTaskNodeCountByTask(ctx context.Context, taskID string) (int64, error)
		GetTaskNodeCountByDepartment(ctx context.Context, departmentID string) (int64, error)
		GetTaskNodeCountByExecutor(ctx context.Context, executorID string) (int64, error)
		GetTaskNodeCountByLeader(ctx context.Context, leaderID string) (int64, error)
		GetTaskNodeCountByEmployee(ctx context.Context, employeeID string) (int64, error) // 统计员工参与的所有任务节点（去重）
		// FindOneSafe 单条查询（对 ex_node_ids 做 COALESCE 防止 NULL 扫描错误）
		FindOneSafe(ctx context.Context, taskNodeID string) (*TaskNode, error)
		// UpdateExNodeIds 单独更新前置节点集合
		UpdateExNodeIds(ctx context.Context, taskNodeID string, exNodeIds string) error
		// InsertTask
		InsertTask(ctx context.Context, data *TaskNode) (sql.Result, error)
		// UpdateChecklistCount 更新任务节点的清单统计数
		UpdateChecklistCount(ctx context.Context, taskNodeId string, totalCount, completedCount int64) error
		// GetCompletedNodeCountByTask 获取任务下已完成的节点数
		GetCompletedNodeCountByTask(ctx context.Context, taskID string) (int64, error)
	}

	customTaskNodeModel struct {
		*defaultTaskNodeModel
	}
)

// NewTaskNodeModel returns a model for the database table.
func NewTaskNodeModel(conn sqlx.SqlConn) TaskNodeModel {
	return &customTaskNodeModel{
		defaultTaskNodeModel: newTaskNodeModel(conn),
	}
}

func (m *customTaskNodeModel) withSession(session sqlx.Session) TaskNodeModel {
	return NewTaskNodeModel(sqlx.NewSqlConnFromSession(session))
}

// FindByTaskID 根据任务ID查找任务节点
func (m *customTaskNodeModel) FindByTaskID(ctx context.Context, taskID string) ([]*TaskNode, error) {
	var taskNodes []*TaskNode
	// 使用 COALESCE 保证 ex_node_ids 非空，避免扫描到 string 报错
	query := `SELECT task_node_id, task_id, department_id, node_name, node_detail,
        COALESCE(ex_node_ids, '') AS ex_node_ids,
        node_deadline, node_start_time, estimated_days, actual_days,
        node_status, node_finish_time, executor_id, leader_id, progress, node_priority,
        create_time, update_time, delete_time
        FROM task_node WHERE task_id = ? AND delete_time IS NULL ORDER BY create_time ASC`
	err := m.conn.QueryRowsCtx(ctx, &taskNodes, query, taskID)
	return taskNodes, err
}

// FindOne 重写生成的 FindOne 方法，对 ex_node_ids 做 COALESCE 以避免 NULL 扫描错误
func (m *customTaskNodeModel) FindOne(ctx context.Context, taskNodeID string) (*TaskNode, error) {
	var node TaskNode
	query := `SELECT task_node_id, task_id, department_id, node_name, node_detail,
        COALESCE(ex_node_ids, '') AS ex_node_ids,
        node_deadline, node_start_time, estimated_days, actual_days,
        node_status, node_finish_time, executor_id, leader_id, progress, node_priority,
        create_time, update_time, delete_time
        FROM task_node WHERE task_node_id = ? LIMIT 1`
	if err := m.conn.QueryRowCtx(ctx, &node, query, taskNodeID); err != nil {
		return nil, err
	}
	return &node, nil
}

// FindOneSafe 同 FindOne，但对 ex_node_ids 做 COALESCE 以避免 NULL 扫描错误（保留以兼容旧代码）
func (m *customTaskNodeModel) FindOneSafe(ctx context.Context, taskNodeID string) (*TaskNode, error) {
	return m.FindOne(ctx, taskNodeID)
}

// FindByTaskNodeIDLeader 根据任务节点id 和leaderid确定任务是否存在
func (m *customTaskNodeModel) FindByTaskNodeIDLeader(ctx context.Context, taskNodeID, LeaderId string) (*TaskNode, error) {
	var taskNode TaskNode
	// 使用 COALESCE 保证 ex_node_ids 非空，避免扫描到 string 报错
	query := `SELECT task_node_id, task_id, department_id, node_name, node_detail,
        COALESCE(ex_node_ids, '') AS ex_node_ids,
        node_deadline, node_start_time, estimated_days, actual_days,
        node_status, node_finish_time, executor_id, leader_id, progress, node_priority,
        create_time, update_time, delete_time
        FROM task_node WHERE task_node_id = ? AND leader_id = ? AND delete_time IS NULL ORDER BY create_time ASC`
	err := m.conn.QueryRowCtx(ctx, &taskNode, query, taskNodeID, LeaderId)
	return &taskNode, err
}

// FindByDepartment 根据部门ID查找任务节点
func (m *customTaskNodeModel) FindByDepartment(ctx context.Context, departmentID string, page, pageSize int) ([]*TaskNode, int64, error) {
	var taskNodes []*TaskNode
	var total int64

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM task_node WHERE department_id = ? AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, departmentID)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	// 使用 COALESCE 保证 ex_node_ids 非空，避免扫描到 string 报错
	query := `SELECT task_node_id, task_id, department_id, node_name, node_detail,
        COALESCE(ex_node_ids, '') AS ex_node_ids,
        node_deadline, node_start_time, estimated_days, actual_days,
        node_status, node_finish_time, executor_id, leader_id, progress, node_priority,
        create_time, update_time, delete_time
        FROM task_node WHERE department_id = ? AND delete_time IS NULL ORDER BY create_time DESC LIMIT ? OFFSET ?`
	err = m.conn.QueryRowsCtx(ctx, &taskNodes, query, departmentID, pageSize, offset)
	return taskNodes, total, err
}

// FindByExecutor 根据执行人ID查找任务节点
func (m *customTaskNodeModel) FindByExecutor(ctx context.Context, executorID string, page, pageSize int) ([]*TaskNode, int64, error) {
	var taskNodes []*TaskNode
	var total int64

	// 查询总数（支持多执行人，使用 FIND_IN_SET）
	countQuery := `SELECT COUNT(*) FROM task_node WHERE FIND_IN_SET(?, executor_id) AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, executorID)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	// 使用 COALESCE 保证 ex_node_ids 非空，避免扫描到 string 报错
	query := `SELECT task_node_id, task_id, department_id, node_name, node_detail,
        COALESCE(ex_node_ids, '') AS ex_node_ids,
        node_deadline, node_start_time, estimated_days, actual_days,
        node_status, node_finish_time, executor_id, leader_id, progress, node_priority,
        create_time, update_time, delete_time
        FROM task_node WHERE FIND_IN_SET(?, executor_id) AND delete_time IS NULL ORDER BY create_time DESC LIMIT ? OFFSET ? `
	err = m.conn.QueryRowsCtx(ctx, &taskNodes, query, executorID, pageSize, offset)
	fmt.Println(pageSize)
	return taskNodes, total, err
}

// FindByLeader 根据负责人ID查找任务节点
func (m *customTaskNodeModel) FindByLeader(ctx context.Context, leaderID string, page, pageSize int) ([]*TaskNode, int64, error) {
	var taskNodes []*TaskNode
	var total int64

	// 查询总数（支持多负责人存储，使用 FIND_IN_SET）
	countQuery := `SELECT COUNT(*) FROM task_node WHERE FIND_IN_SET(?, leader_id) AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, leaderID)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	fmt.Println(pageSize)
	// 使用 COALESCE 保证 ex_node_ids 非空，避免扫描到 string 报错
	query := `SELECT task_node_id, task_id, department_id, node_name, node_detail,
        COALESCE(ex_node_ids, '') AS ex_node_ids,
        node_deadline, node_start_time, estimated_days, actual_days,
        node_status, node_finish_time, executor_id, leader_id, progress, node_priority,
        create_time, update_time, delete_time
        FROM task_node WHERE FIND_IN_SET(?, leader_id) AND delete_time IS NULL ORDER BY create_time DESC LIMIT ? OFFSET ?`
	err = m.conn.QueryRowsCtx(ctx, &taskNodes, query, leaderID, pageSize, offset)
	return taskNodes, total, err
}

// FindByStatus 根据状态查找任务节点
func (m *customTaskNodeModel) FindByStatus(ctx context.Context, status int) ([]*TaskNode, error) {
	var taskNodes []*TaskNode
	// 使用 COALESCE 保证 ex_node_ids 非空，避免扫描到 string 报错
	query := `SELECT task_node_id, task_id, department_id, node_name, node_detail,
        COALESCE(ex_node_ids, '') AS ex_node_ids,
        node_deadline, node_start_time, estimated_days, actual_days,
        node_status, node_finish_time, executor_id, leader_id, progress, node_priority,
        create_time, update_time, delete_time
        FROM task_node WHERE node_status = ? AND delete_time IS NULL ORDER BY create_time DESC`
	err := m.conn.QueryRowsCtx(ctx, &taskNodes, query, status)
	return taskNodes, err
}

// FindByDeadlineRange 根据截止时间范围查找任务节点
func (m *customTaskNodeModel) FindByDeadlineRange(ctx context.Context, startTime, endTime string) ([]*TaskNode, error) {
	var taskNodes []*TaskNode
	// 使用 COALESCE 保证 ex_node_ids 非空，避免扫描到 string 报错
	query := `SELECT task_node_id, task_id, department_id, node_name, node_detail,
        COALESCE(ex_node_ids, '') AS ex_node_ids,
        node_deadline, node_start_time, estimated_days, actual_days,
        node_status, node_finish_time, executor_id, leader_id, progress, node_priority,
        create_time, update_time, delete_time
        FROM task_node WHERE node_deadline >= ? AND node_deadline <= ? AND delete_time IS NULL ORDER BY node_deadline ASC`
	err := m.conn.QueryRowsCtx(ctx, &taskNodes, query, startTime, endTime)
	return taskNodes, err
}

// FindByPage 分页查找任务节点
func (m *customTaskNodeModel) FindByPage(ctx context.Context, page, pageSize int) ([]*TaskNode, int64, error) {
	var taskNodes []*TaskNode
	var total int64

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM task_node WHERE delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	// 使用 COALESCE 保证 ex_node_ids 非空，避免扫描到 string 报错
	query := `SELECT task_node_id, task_id, department_id, node_name, node_detail,
        COALESCE(ex_node_ids, '') AS ex_node_ids,
        node_deadline, node_start_time, estimated_days, actual_days,
        node_status, node_finish_time, executor_id, leader_id, progress, node_priority,
        create_time, update_time, delete_time
        FROM task_node WHERE delete_time IS NULL ORDER BY create_time DESC LIMIT ? OFFSET ?`
	err = m.conn.QueryRowsCtx(ctx, &taskNodes, query, pageSize, offset)
	return taskNodes, total, err
}

// SearchTaskNodes 搜索任务节点
func (m *customTaskNodeModel) SearchTaskNodes(ctx context.Context, keyword string, page, pageSize int) ([]*TaskNode, int64, error) {
	var taskNodes []*TaskNode
	var total int64

	// 构建搜索条件
	searchCondition := fmt.Sprintf("(node_name LIKE '%%%s%%' OR node_detail LIKE '%%%s%%') AND delete_time IS NULL", keyword, keyword)

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM task_node WHERE %s", searchCondition)
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	// 使用 COALESCE 保证 ex_node_ids 非空，避免扫描到 string 报错
	query := fmt.Sprintf(`SELECT task_node_id, task_id, department_id, node_name, node_detail,
        COALESCE(ex_node_ids, '') AS ex_node_ids,
        node_deadline, node_start_time, estimated_days, actual_days,
        node_status, node_finish_time, executor_id, leader_id, progress, node_priority,
        create_time, update_time, delete_time
        FROM task_node WHERE %s ORDER BY create_time DESC LIMIT ? OFFSET ?`, searchCondition)
	err = m.conn.QueryRowsCtx(ctx, &taskNodes, query, pageSize, offset)
	return taskNodes, total, err
}

// UpdateStatus 更新任务节点状态
func (m *customTaskNodeModel) UpdateStatus(ctx context.Context, id string, status int) error {
	query := `UPDATE task_node SET node_status = ?, update_time = NOW() WHERE task_node_id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, status, id)
	return err
}

// UpdateProgress 更新任务节点进度
func (m *customTaskNodeModel) UpdateProgress(ctx context.Context, id string, progress int) error {
	query := `UPDATE task_node SET progress = ?, update_time = NOW() WHERE task_node_id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, progress, id)
	return err
}

// UpdateActualHours 更新任务节点实际工时
func (m *customTaskNodeModel) UpdateActualHours(ctx context.Context, id string, actualHours int) error {
	query := `UPDATE task_node SET actual_hours = ?, update_time = NOW() WHERE task_node_id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, actualHours, id)
	return err
}

// UpdateExecutor 更新任务节点执行人
func (m *customTaskNodeModel) UpdateExecutor(ctx context.Context, id, executorID string) error {
	query := `UPDATE task_node SET executor_id = ?, update_time = NOW() WHERE task_node_id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, executorID, id)
	return err
}

// UpdateLeader 更新任务节点负责人
func (m *customTaskNodeModel) UpdateLeader(ctx context.Context, id, leaderID string) error {
	query := `UPDATE task_node SET leader_id = ?, update_time = NOW() WHERE task_node_id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, leaderID, id)
	return err
}

// UpdateDeadline 更新任务节点截止时间
func (m *customTaskNodeModel) UpdateDeadline(ctx context.Context, id, deadline string) error {
	query := `UPDATE task_node SET node_deadline = ?, update_time = NOW() WHERE task_node_id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, deadline, id)
	return err
}

// SoftDelete 软删除任务节点
func (m *customTaskNodeModel) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE task_node SET delete_time = NOW() WHERE task_node_id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// BatchUpdateStatus 批量更新任务节点状态
func (m *customTaskNodeModel) BatchUpdateStatus(ctx context.Context, ids []string, status int) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(ids)-1) + "?"
	query := fmt.Sprintf("UPDATE task_node SET status = ?, update_time = NOW() WHERE task_node_id IN (%s) AND delete_time IS NULL", placeholders)

	args := make([]interface{}, 0, len(ids)+1)
	args = append(args, status)
	for _, id := range ids {
		args = append(args, id)
	}

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

// GetTaskNodeCount 获取任务节点总数
func (m *customTaskNodeModel) GetTaskNodeCount(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_node WHERE delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

// GetTaskNodeCountByStatus 根据状态获取任务节点数量
func (m *customTaskNodeModel) GetTaskNodeCountByStatus(ctx context.Context, status int) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_node WHERE status = ? AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query, status)
	return count, err
}

// GetTaskNodeCountByTask 根据任务获取任务节点数量
func (m *customTaskNodeModel) GetTaskNodeCountByTask(ctx context.Context, taskID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_node WHERE task_id = ? AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query, taskID)
	return count, err
}

// GetTaskNodeCountByDepartment 根据部门获取任务节点数量
func (m *customTaskNodeModel) GetTaskNodeCountByDepartment(ctx context.Context, departmentID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_node WHERE department_id = ? AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query, departmentID)
	return count, err
}

// GetTaskNodeCountByExecutor 根据执行人获取任务节点数量
func (m *customTaskNodeModel) GetTaskNodeCountByExecutor(ctx context.Context, executorID string) (int64, error) {
	var count int64
	// 支持多执行人
	query := `SELECT COUNT(*) FROM task_node WHERE FIND_IN_SET(?, executor_id) AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query, executorID)
	return count, err
}

// GetTaskNodeCountByLeader 根据负责人获取任务节点数量
func (m *customTaskNodeModel) GetTaskNodeCountByLeader(ctx context.Context, leaderID string) (int64, error) {
	var count int64
	// 支持多负责人
	query := `SELECT COUNT(*) FROM task_node WHERE FIND_IN_SET(?, leader_id) AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query, leaderID)
	return count, err
}

// GetTaskNodeCountByEmployee 统计员工参与的所有任务节点数量（去重）
// 包括作为执行人和负责人的任务节点，但每个任务节点只计算一次
func (m *customTaskNodeModel) GetTaskNodeCountByEmployee(ctx context.Context, employeeID string) (int64, error) {
	var count int64
	// 使用 DISTINCT 去重，统计该员工作为执行人或负责人的所有任务节点
	query := `SELECT COUNT(DISTINCT task_node_id) FROM task_node 
		WHERE (FIND_IN_SET(?, executor_id) OR FIND_IN_SET(?, leader_id)) 
		AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query, employeeID, employeeID)
	return count, err
}

// UpdateExNodeIds 单独更新前置节点集合
func (m *customTaskNodeModel) UpdateExNodeIds(ctx context.Context, taskNodeID string, exNodeIds string) error {
	query := `UPDATE task_node SET ex_node_ids = ?, update_time = NOW() WHERE task_node_id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, exNodeIds, taskNodeID)
	return err
}

// Insert 覆盖生成的 Insert 方法，正确包含所有字段（包括 ExNodeIds）
// 生成的代码遗漏了 ExNodeIds 字段，导致列数和值数不匹配
func (m *customTaskNodeModel) InsertTask(ctx context.Context, data *TaskNode) (sql.Result, error) {
	// 手动构建插入语句，包含所有17个字段（排除 create_time 和 update_time）
	query := `INSERT INTO task_node 
		(task_node_id, task_id, department_id, node_name, node_detail, ex_node_ids, 
		 node_deadline, node_start_time, estimated_days, actual_days, node_status, 
		 node_finish_time, executor_id, leader_id, progress, node_priority, delete_time) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	ret, err := m.conn.ExecCtx(ctx, query,
		data.TaskNodeId,     // 1
		data.TaskId,         // 2
		data.DepartmentId,   // 3
		data.NodeName,       // 4
		data.NodeDetail,     // 5
		data.ExNodeIds,      // 6 - 这是生成的代码遗漏的字段
		data.NodeDeadline,   // 7
		data.NodeStartTime,  // 8
		data.EstimatedDays,  // 9
		data.ActualDays,     // 10
		data.NodeStatus,     // 11
		data.NodeFinishTime, // 12
		data.ExecutorId,     // 13
		data.LeaderId,       // 14
		data.Progress,       // 15
		data.NodePriority,   // 16
		data.DeleteTime,     // 17
	)
	return ret, err
}

// UpdateChecklistCount 更新任务节点的清单统计数
// 注意：需要先在数据库中添加 total_checklist_count 和 completed_checklist_count 字段
func (m *customTaskNodeModel) UpdateChecklistCount(ctx context.Context, taskNodeId string, totalCount, completedCount int64) error {
	query := `UPDATE task_node SET total_checklist_count = ?, completed_checklist_count = ?, update_time = NOW() WHERE task_node_id = ? AND delete_time IS NULL`
	_, err := m.conn.ExecCtx(ctx, query, totalCount, completedCount, taskNodeId)
	return err
}

// GetCompletedNodeCountByTask 获取任务下已完成的节点数
func (m *customTaskNodeModel) GetCompletedNodeCountByTask(ctx context.Context, taskID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_node WHERE task_id = ? AND node_status = 2 AND delete_time IS NULL`
	err := m.conn.QueryRowCtx(ctx, &count, query, taskID)
	return count, err
}
