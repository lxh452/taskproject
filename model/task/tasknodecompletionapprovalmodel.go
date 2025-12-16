package task

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// TaskNodeCompletionApproval 任务节点完成审批记录
type TaskNodeCompletionApproval struct {
	Id           int64          `db:"id"`            // 自增ID
	ApprovalId   string         `db:"approval_id"`   // 审批记录ID
	TaskNodeId   string         `db:"task_node_id"`  // 任务节点ID
	ApproverId   string         `db:"approver_id"`   // 审批人ID（项目负责人）
	ApproverName string         `db:"approver_name"` // 审批人姓名
	ApprovalType int64          `db:"approval_type"` // 审批类型 0-待审批 1-同意 2-拒绝
	Comment      sql.NullString `db:"comment"`       // 审批意见
	CreateTime   time.Time      `db:"create_time"`   // 创建时间
	UpdateTime   time.Time      `db:"update_time"`   // 更新时间
}

type (
	TaskNodeCompletionApprovalModel interface {
		Insert(ctx context.Context, data *TaskNodeCompletionApproval) (sql.Result, error)
		FindOne(ctx context.Context, approvalId string) (*TaskNodeCompletionApproval, error)
		FindByTaskNodeId(ctx context.Context, taskNodeId string) ([]*TaskNodeCompletionApproval, error)
		FindLatestByTaskNodeId(ctx context.Context, taskNodeId string) (*TaskNodeCompletionApproval, error)
		Update(ctx context.Context, data *TaskNodeCompletionApproval) error
	}

	defaultTaskNodeCompletionApprovalModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

func NewTaskNodeCompletionApprovalModel(conn sqlx.SqlConn) TaskNodeCompletionApprovalModel {
	return &defaultTaskNodeCompletionApprovalModel{
		conn:  conn,
		table: "`task_node_completion_approval`",
	}
}

func (m *defaultTaskNodeCompletionApprovalModel) Insert(ctx context.Context, data *TaskNodeCompletionApproval) (sql.Result, error) {
	query := fmt.Sprintf("INSERT INTO %s (approval_id, task_node_id, approver_id, approver_name, approval_type, comment, create_time, update_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", m.table)
	ret, err := m.conn.ExecCtx(ctx, query, data.ApprovalId, data.TaskNodeId, data.ApproverId, data.ApproverName, data.ApprovalType, data.Comment, data.CreateTime, data.UpdateTime)
	return ret, err
}

func (m *defaultTaskNodeCompletionApprovalModel) FindOne(ctx context.Context, approvalId string) (*TaskNodeCompletionApproval, error) {
	query := fmt.Sprintf("SELECT id, approval_id, task_node_id, approver_id, approver_name, approval_type, comment, create_time, update_time FROM %s WHERE approval_id = ? LIMIT 1", m.table)
	var resp TaskNodeCompletionApproval
	err := m.conn.QueryRowCtx(ctx, &resp, query, approvalId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultTaskNodeCompletionApprovalModel) FindByTaskNodeId(ctx context.Context, taskNodeId string) ([]*TaskNodeCompletionApproval, error) {
	var approvals []*TaskNodeCompletionApproval
	query := fmt.Sprintf("SELECT id, approval_id, task_node_id, approver_id, approver_name, approval_type, comment, create_time, update_time FROM %s WHERE task_node_id = ? ORDER BY create_time DESC", m.table)
	err := m.conn.QueryRowsCtx(ctx, &approvals, query, taskNodeId)
	return approvals, err
}

func (m *defaultTaskNodeCompletionApprovalModel) FindLatestByTaskNodeId(ctx context.Context, taskNodeId string) (*TaskNodeCompletionApproval, error) {
	query := fmt.Sprintf("SELECT id, approval_id, task_node_id, approver_id, approver_name, approval_type, comment, create_time, update_time FROM %s WHERE task_node_id = ? ORDER BY create_time DESC LIMIT 1", m.table)
	var resp TaskNodeCompletionApproval
	err := m.conn.QueryRowCtx(ctx, &resp, query, taskNodeId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultTaskNodeCompletionApprovalModel) Update(ctx context.Context, data *TaskNodeCompletionApproval) error {
	query := fmt.Sprintf("UPDATE %s SET approver_id = ?, approver_name = ?, approval_type = ?, comment = ?, update_time = ? WHERE approval_id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.ApproverId, data.ApproverName, data.ApprovalType, data.Comment, data.UpdateTime, data.ApprovalId)
	return err
}
