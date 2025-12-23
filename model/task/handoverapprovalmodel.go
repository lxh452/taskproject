package task

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// HandoverApproval 交接审批记录（同时支持交接审批和任务节点完成审批）
type HandoverApproval struct {
	Id           int64          `db:"id"`            // 自增ID
	ApprovalId   string         `db:"approval_id"`   // 审批记录ID
	HandoverId   string         `db:"handover_id"`   // 交接ID（交接审批时使用）
	TaskNodeId   sql.NullString `db:"task_node_id"`  // 任务节点ID（任务节点完成审批时使用）
	ApprovalStep int64          `db:"approval_step"` // 审批步骤 1-接收人确认 2-上级审批 3-任务节点完成审批
	ApproverId   string         `db:"approver_id"`   // 审批人ID
	ApproverName string         `db:"approver_name"` // 审批人姓名
	ApprovalType int64          `db:"approval_type"` // 审批类型 0-待审批 1-同意 2-拒绝
	Comment      sql.NullString `db:"comment"`       // 审批意见
	CreateTime   time.Time      `db:"create_time"`   // 创建时间
	UpdateTime   sql.NullTime   `db:"update_time"`   // 更新时间
}

type (
	HandoverApprovalModel interface {
		Insert(ctx context.Context, data *HandoverApproval) (sql.Result, error)
		FindOne(ctx context.Context, approvalId string) (*HandoverApproval, error)
		FindByHandoverId(ctx context.Context, handoverId string) ([]*HandoverApproval, error)
		FindLatestByHandoverId(ctx context.Context, handoverId string) (*HandoverApproval, error)
		// 任务节点完成审批相关方法
		FindByTaskNodeId(ctx context.Context, taskNodeId string) ([]*HandoverApproval, error)
		FindLatestByTaskNodeId(ctx context.Context, taskNodeId string) (*HandoverApproval, error)
		Update(ctx context.Context, data *HandoverApproval) error
	}

	defaultHandoverApprovalModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

func NewHandoverApprovalModel(conn sqlx.SqlConn) HandoverApprovalModel {
	return &defaultHandoverApprovalModel{
		conn:  conn,
		table: "`handover_approval`",
	}
}

func (m *defaultHandoverApprovalModel) Insert(ctx context.Context, data *HandoverApproval) (sql.Result, error) {
	query := fmt.Sprintf("INSERT INTO %s (approval_id, handover_id, task_node_id, approval_step, approver_id, approver_name, approval_type, comment, create_time, update_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table)
	ret, err := m.conn.ExecCtx(ctx, query, data.ApprovalId, data.HandoverId, data.TaskNodeId, data.ApprovalStep, data.ApproverId, data.ApproverName, data.ApprovalType, data.Comment, data.CreateTime, data.UpdateTime)
	return ret, err
}

func (m *defaultHandoverApprovalModel) FindOne(ctx context.Context, approvalId string) (*HandoverApproval, error) {
	query := fmt.Sprintf("SELECT id, approval_id, handover_id, COALESCE(task_node_id, '') as task_node_id, approval_step, approver_id, approver_name, approval_type, comment, create_time, update_time FROM %s WHERE approval_id = ? LIMIT 1", m.table)
	var resp HandoverApproval
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

func (m *defaultHandoverApprovalModel) FindByHandoverId(ctx context.Context, handoverId string) ([]*HandoverApproval, error) {
	var approvals []*HandoverApproval
	query := fmt.Sprintf("SELECT id, approval_id, handover_id, COALESCE(task_node_id, '') as task_node_id, approval_step, approver_id, approver_name, approval_type, comment, create_time, update_time FROM %s WHERE handover_id = ? ORDER BY approval_step ASC, create_time ASC", m.table)
	err := m.conn.QueryRowsCtx(ctx, &approvals, query, handoverId)
	return approvals, err
}

func (m *defaultHandoverApprovalModel) FindLatestByHandoverId(ctx context.Context, handoverId string) (*HandoverApproval, error) {
	query := fmt.Sprintf("SELECT id, approval_id, handover_id, COALESCE(task_node_id, '') as task_node_id, approval_step, approver_id, approver_name, approval_type, comment, create_time, update_time FROM %s WHERE handover_id = ? ORDER BY create_time DESC LIMIT 1", m.table)
	var resp HandoverApproval
	err := m.conn.QueryRowCtx(ctx, &resp, query, handoverId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByTaskNodeId 根据任务节点ID查找审批记录
func (m *defaultHandoverApprovalModel) FindByTaskNodeId(ctx context.Context, taskNodeId string) ([]*HandoverApproval, error) {
	var approvals []*HandoverApproval
	query := fmt.Sprintf("SELECT id, approval_id, handover_id, COALESCE(task_node_id, '') as task_node_id, approval_step, approver_id, approver_name, approval_type, comment, create_time, update_time FROM %s WHERE task_node_id = ? ORDER BY create_time DESC", m.table)
	err := m.conn.QueryRowsCtx(ctx, &approvals, query, taskNodeId)
	return approvals, err
}

// FindLatestByTaskNodeId 根据任务节点ID查找最新的审批记录
func (m *defaultHandoverApprovalModel) FindLatestByTaskNodeId(ctx context.Context, taskNodeId string) (*HandoverApproval, error) {
	query := fmt.Sprintf("SELECT id, approval_id, handover_id, COALESCE(task_node_id, '') as task_node_id, approval_step, approver_id, approver_name, approval_type, comment, create_time, update_time FROM %s WHERE task_node_id = ? ORDER BY create_time DESC LIMIT 1", m.table)
	var resp HandoverApproval
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

// Update 更新审批记录
func (m *defaultHandoverApprovalModel) Update(ctx context.Context, data *HandoverApproval) error {
	query := fmt.Sprintf("UPDATE %s SET approver_id = ?, approver_name = ?, approval_type = ?, comment = ?, update_time = ? WHERE approval_id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.ApproverId, data.ApproverName, data.ApprovalType, data.Comment, data.UpdateTime, data.ApprovalId)
	return err
}
