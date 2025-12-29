package task

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TaskLogModel = (*customTaskLogModel)(nil)

type (
	// TaskLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTaskLogModel.
	TaskLogModel interface {
		taskLogModel
		withSession(session sqlx.Session) TaskLogModel

		// 任务日志CRUD操作
		FindByTaskID(ctx context.Context, taskID string) ([]*TaskLog, error)
		FindByTaskNodeID(ctx context.Context, taskNodeID string) ([]*TaskLog, error)
		FindByOperator(ctx context.Context, operatorID string, page, pageSize int) ([]*TaskLog, int64, error)
		FindByLogType(ctx context.Context, logType string) ([]*TaskLog, error)
		FindByPage(ctx context.Context, page, pageSize int) ([]*TaskLog, int64, error)
		SearchTaskLogs(ctx context.Context, keyword string, page, pageSize int) ([]*TaskLog, int64, error)
		GetTaskLogCount(ctx context.Context) (int64, error)
		GetTaskLogCountByTask(ctx context.Context, taskID string) (int64, error)
		GetTaskLogCountByTaskNode(ctx context.Context, taskNodeID string) (int64, error)
		GetTaskLogCountByOperator(ctx context.Context, operatorID string) (int64, error)
		GetTaskLogCountByLogType(ctx context.Context, logType string) (int64, error)
	}

	customTaskLogModel struct {
		*defaultTaskLogModel
	}
)

// NewTaskLogModel returns a model for the database table.
func NewTaskLogModel(conn sqlx.SqlConn) TaskLogModel {
	return &customTaskLogModel{
		defaultTaskLogModel: newTaskLogModel(conn),
	}
}

func (m *customTaskLogModel) withSession(session sqlx.Session) TaskLogModel {
	return NewTaskLogModel(sqlx.NewSqlConnFromSession(session))
}

// FindByTaskID 根据任务ID查找任务日志
func (m *customTaskLogModel) FindByTaskID(ctx context.Context, taskID string) ([]*TaskLog, error) {
	var taskLogs []*TaskLog
	query := `SELECT * FROM task_log WHERE task_id = ? ORDER BY create_time DESC`
	err := m.conn.QueryRowsCtx(ctx, &taskLogs, query, taskID)
	return taskLogs, err
}

// FindByTaskNodeID 根据任务节点ID查找任务日志
func (m *customTaskLogModel) FindByTaskNodeID(ctx context.Context, taskNodeID string) ([]*TaskLog, error) {
	var taskLogs []*TaskLog
	query := `SELECT * FROM task_log WHERE task_node_id = ? ORDER BY create_time DESC`
	err := m.conn.QueryRowsCtx(ctx, &taskLogs, query, taskNodeID)
	return taskLogs, err
}

// FindByOperator 根据操作人ID查找任务日志
func (m *customTaskLogModel) FindByOperator(ctx context.Context, operatorID string, page, pageSize int) ([]*TaskLog, int64, error) {
	var taskLogs []*TaskLog
	var total int64

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM task_log WHERE operator_id = ?`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, operatorID)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `SELECT * FROM task_log WHERE operator_id = ? ORDER BY create_time DESC LIMIT ? OFFSET ?`
	err = m.conn.QueryRowsCtx(ctx, &taskLogs, query, operatorID, pageSize, offset)
	return taskLogs, total, err
}

// FindByLogType 根据日志类型查找任务日志
func (m *customTaskLogModel) FindByLogType(ctx context.Context, logType string) ([]*TaskLog, error) {
	var taskLogs []*TaskLog
	query := `SELECT * FROM task_log WHERE log_type = ? ORDER BY create_time DESC`
	err := m.conn.QueryRowsCtx(ctx, &taskLogs, query, logType)
	return taskLogs, err
}

// FindByPage 分页查找任务日志
func (m *customTaskLogModel) FindByPage(ctx context.Context, page, pageSize int) ([]*TaskLog, int64, error) {
	var taskLogs []*TaskLog
	var total int64

	// 查询总数
	countQuery := `SELECT COUNT(*) FROM task_log`
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `SELECT * FROM task_log ORDER BY create_time DESC LIMIT ? OFFSET ?`
	err = m.conn.QueryRowsCtx(ctx, &taskLogs, query, pageSize, offset)
	return taskLogs, total, err
}

// SearchTaskLogs 搜索任务日志
func (m *customTaskLogModel) SearchTaskLogs(ctx context.Context, keyword string, page, pageSize int) ([]*TaskLog, int64, error) {
	var taskLogs []*TaskLog
	var total int64

	// 构建搜索条件
	searchCondition := fmt.Sprintf("log_content LIKE '%%%s%%'", keyword)

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM task_log WHERE %s", searchCondition)
	err := m.conn.QueryRowCtx(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("SELECT * FROM task_log WHERE %s ORDER BY create_time DESC LIMIT ? OFFSET ?", searchCondition)
	err = m.conn.QueryRowsCtx(ctx, &taskLogs, query, pageSize, offset)
	return taskLogs, total, err
}

// GetTaskLogCount 获取任务日志总数
func (m *customTaskLogModel) GetTaskLogCount(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_log`
	err := m.conn.QueryRowCtx(ctx, &count, query)
	return count, err
}

// GetTaskLogCountByTask 根据任务获取任务日志数量
func (m *customTaskLogModel) GetTaskLogCountByTask(ctx context.Context, taskID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_log WHERE task_id = ?`
	err := m.conn.QueryRowCtx(ctx, &count, query, taskID)
	return count, err
}

// GetTaskLogCountByTaskNode 根据任务节点获取任务日志数量
func (m *customTaskLogModel) GetTaskLogCountByTaskNode(ctx context.Context, taskNodeID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_log WHERE task_node_id = ?`
	err := m.conn.QueryRowCtx(ctx, &count, query, taskNodeID)
	return count, err
}

// GetTaskLogCountByOperator 根据操作人获取任务日志数量
func (m *customTaskLogModel) GetTaskLogCountByOperator(ctx context.Context, operatorID string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_log WHERE operator_id = ?`
	err := m.conn.QueryRowCtx(ctx, &count, query, operatorID)
	return count, err
}

// GetTaskLogCountByLogType 根据日志类型获取任务日志数量
func (m *customTaskLogModel) GetTaskLogCountByLogType(ctx context.Context, logType string) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM task_log WHERE log_type = ?`
	err := m.conn.QueryRowCtx(ctx, &count, query, logType)
	return count, err
}
