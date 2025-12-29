package task

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TaskChecklistModel = (*customTaskChecklistModel)(nil)

type (
	// TaskChecklistModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTaskChecklistModel.
	TaskChecklistModel interface {
		taskChecklistModel
		// 根据任务节点ID查询所有清单（不含已删除）
		FindByTaskNodeId(ctx context.Context, taskNodeId string) ([]*TaskChecklist, error)
		// 根据任务节点ID和创建者ID查询清单（不含已删除）
		FindByTaskNodeIdAndCreator(ctx context.Context, taskNodeId, creatorId string) ([]*TaskChecklist, error)
		// 统计任务节点的清单数量
		CountByTaskNodeId(ctx context.Context, taskNodeId string) (total int64, completed int64, err error)
		// 软删除清单
		SoftDelete(ctx context.Context, checklistId string) error
		// 根据任务节点ID分页查询清单
		FindByTaskNodeIdWithPage(ctx context.Context, taskNodeId string, page, pageSize int) ([]*TaskChecklist, int64, error)
		// 批量更新清单完成状态
		BatchUpdateCompleteStatus(ctx context.Context, checklistIds []string, isCompleted int64) error
		// 根据创建者ID分页查询清单（我的清单）
		FindByCreatorIdWithPage(ctx context.Context, creatorId string, isCompleted int64, page, pageSize int) ([]*TaskChecklist, int64, error)
		// 统计创建者的未完成清单数量
		CountUncompletedByCreatorId(ctx context.Context, creatorId string) (int64, error)
	}

	customTaskChecklistModel struct {
		*defaultTaskChecklistModel
	}
)

// NewTaskChecklistModel returns a model for the database table.
func NewTaskChecklistModel(conn sqlx.SqlConn) TaskChecklistModel {
	return &customTaskChecklistModel{
		defaultTaskChecklistModel: newTaskChecklistModel(conn),
	}
}

// FindByTaskNodeId 根据任务节点ID查询所有清单（不含已删除）
func (m *customTaskChecklistModel) FindByTaskNodeId(ctx context.Context, taskNodeId string) ([]*TaskChecklist, error) {
	query := fmt.Sprintf("select %s from %s where `task_node_id` = ? and `delete_time` is null order by `sort_order` asc, `create_time` asc", taskChecklistRows, m.table)
	var resp []*TaskChecklist
	err := m.conn.QueryRowsCtx(ctx, &resp, query, taskNodeId)
	return resp, err
}

// FindByTaskNodeIdAndCreator 根据任务节点ID和创建者ID查询清单（不含已删除）
func (m *customTaskChecklistModel) FindByTaskNodeIdAndCreator(ctx context.Context, taskNodeId, creatorId string) ([]*TaskChecklist, error) {
	query := fmt.Sprintf("select %s from %s where `task_node_id` = ? and `creator_id` = ? and `delete_time` is null order by `sort_order` asc, `create_time` asc", taskChecklistRows, m.table)
	var resp []*TaskChecklist
	err := m.conn.QueryRowsCtx(ctx, &resp, query, taskNodeId, creatorId)
	return resp, err
}

// CountByTaskNodeId 统计任务节点的清单数量
func (m *customTaskChecklistModel) CountByTaskNodeId(ctx context.Context, taskNodeId string) (total int64, completed int64, err error) {
	// 查询总数
	totalQuery := fmt.Sprintf("select count(*) from %s where `task_node_id` = ? and `delete_time` is null", m.table)
	err = m.conn.QueryRowCtx(ctx, &total, totalQuery, taskNodeId)
	if err != nil {
		return 0, 0, err
	}

	// 查询已完成数
	completedQuery := fmt.Sprintf("select count(*) from %s where `task_node_id` = ? and `is_completed` = 1 and `delete_time` is null", m.table)
	err = m.conn.QueryRowCtx(ctx, &completed, completedQuery, taskNodeId)
	if err != nil {
		return 0, 0, err
	}

	return total, completed, nil
}

// SoftDelete 软删除清单
func (m *customTaskChecklistModel) SoftDelete(ctx context.Context, checklistId string) error {
	query := fmt.Sprintf("update %s set `delete_time` = now() where `checklist_id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, checklistId)
	return err
}

// FindByTaskNodeIdWithPage 根据任务节点ID分页查询清单
func (m *customTaskChecklistModel) FindByTaskNodeIdWithPage(ctx context.Context, taskNodeId string, page, pageSize int) ([]*TaskChecklist, int64, error) {
	// 查询总数
	countQuery := fmt.Sprintf("select count(*) from %s where `task_node_id` = ? and `delete_time` is null", m.table)
	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, taskNodeId)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where `task_node_id` = ? and `delete_time` is null order by `sort_order` asc, `create_time` asc limit ?, ?", taskChecklistRows, m.table)
	var resp []*TaskChecklist
	err = m.conn.QueryRowsCtx(ctx, &resp, query, taskNodeId, offset, pageSize)
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	return resp, total, nil
}

// BatchUpdateCompleteStatus 批量更新清单完成状态
func (m *customTaskChecklistModel) BatchUpdateCompleteStatus(ctx context.Context, checklistIds []string, isCompleted int64) error {
	if len(checklistIds) == 0 {
		return nil
	}

	// 构建IN子句的占位符
	placeholders := make([]string, len(checklistIds))
	args := make([]interface{}, len(checklistIds)+1)
	args[0] = isCompleted
	for i, id := range checklistIds {
		placeholders[i] = "?"
		args[i+1] = id
	}

	var query string
	if isCompleted == 1 {
		query = fmt.Sprintf("update %s set `is_completed` = ?, `complete_time` = now() where `checklist_id` in (%s)", m.table, joinStrings(placeholders, ","))
	} else {
		query = fmt.Sprintf("update %s set `is_completed` = ?, `complete_time` = null where `checklist_id` in (%s)", m.table, joinStrings(placeholders, ","))
	}

	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

// joinStrings 连接字符串切片
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// FindByCreatorIdWithPage 根据创建者ID分页查询清单（我的清单）
// isCompleted: -1或其他-全部, 0-未完成, 1-已完成
func (m *customTaskChecklistModel) FindByCreatorIdWithPage(ctx context.Context, creatorId string, isCompleted int64, page, pageSize int) ([]*TaskChecklist, int64, error) {
	var countQuery, dataQuery string
	var args []interface{}

	if isCompleted == 0 || isCompleted == 1 {
		// 筛选特定状态
		countQuery = fmt.Sprintf("select count(*) from %s where `creator_id` = ? and `is_completed` = ? and `delete_time` is null", m.table)
		args = []interface{}{creatorId, isCompleted}
	} else {
		// 全部状态
		countQuery = fmt.Sprintf("select count(*) from %s where `creator_id` = ? and `delete_time` is null", m.table)
		args = []interface{}{creatorId}
	}

	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if isCompleted == 0 || isCompleted == 1 {
		dataQuery = fmt.Sprintf("select %s from %s where `creator_id` = ? and `is_completed` = ? and `delete_time` is null order by `create_time` desc limit ?, ?", taskChecklistRows, m.table)
		args = append(args, offset, pageSize)
	} else {
		dataQuery = fmt.Sprintf("select %s from %s where `creator_id` = ? and `delete_time` is null order by `create_time` desc limit ?, ?", taskChecklistRows, m.table)
		args = append(args, offset, pageSize)
	}

	var resp []*TaskChecklist
	err = m.conn.QueryRowsCtx(ctx, &resp, dataQuery, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}

	return resp, total, nil
}

// CountUncompletedByCreatorId 统计创建者的未完成清单数量
func (m *customTaskChecklistModel) CountUncompletedByCreatorId(ctx context.Context, creatorId string) (int64, error) {
	query := fmt.Sprintf("select count(*) from %s where `creator_id` = ? and `is_completed` = 0 and `delete_time` is null", m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, creatorId)
	return count, err
}
