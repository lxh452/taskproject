package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// JoinApplication 加入公司申请
type JoinApplication struct {
	Id          string         `db:"id"`           // 申请ID
	UserId      string         `db:"user_id"`      // 申请用户ID
	CompanyId   string         `db:"company_id"`   // 目标公司ID
	InviteCode  sql.NullString `db:"invite_code"`  // 使用的邀请码
	ApplyReason sql.NullString `db:"apply_reason"` // 申请理由
	Status      int64          `db:"status"`       // 状态 0-待审批 1-已通过 2-已拒绝 3-已取消
	ApproverId  sql.NullString `db:"approver_id"`  // 审批人员工ID
	ApproveTime sql.NullTime   `db:"approve_time"` // 审批时间
	ApproveNote sql.NullString `db:"approve_note"` // 审批备注
	CreateTime  time.Time      `db:"create_time"`  // 申请时间
	UpdateTime  time.Time      `db:"update_time"`  // 更新时间
}

// 申请状态常量
const (
	JoinApplicationStatusPending  = 0 // 待审批
	JoinApplicationStatusApproved = 1 // 已通过
	JoinApplicationStatusRejected = 2 // 已拒绝
	JoinApplicationStatusCanceled = 3 // 已取消
)

type JoinApplicationModel interface {
	Insert(ctx context.Context, data *JoinApplication) (sql.Result, error)
	FindOne(ctx context.Context, id string) (*JoinApplication, error)
	Update(ctx context.Context, data *JoinApplication) error
	Delete(ctx context.Context, id string) error
	FindByUserId(ctx context.Context, userId string) ([]*JoinApplication, error)
	FindByCompanyId(ctx context.Context, companyId string, status *int) ([]*JoinApplication, error)
	FindPendingByUserId(ctx context.Context, userId string) (*JoinApplication, error)
	UpdateStatus(ctx context.Context, id string, status int, approverId, approveNote string) error
}

type defaultJoinApplicationModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewJoinApplicationModel(conn sqlx.SqlConn) JoinApplicationModel {
	return &defaultJoinApplicationModel{
		conn:  conn,
		table: "`join_application`",
	}
}

func (m *defaultJoinApplicationModel) Insert(ctx context.Context, data *JoinApplication) (sql.Result, error) {
	query := fmt.Sprintf("INSERT INTO %s (`id`, `user_id`, `company_id`, `invite_code`, `apply_reason`, `status`, `create_time`, `update_time`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", m.table)
	return m.conn.ExecCtx(ctx, query, data.Id, data.UserId, data.CompanyId, data.InviteCode, data.ApplyReason, data.Status, data.CreateTime, data.UpdateTime)
}

func (m *defaultJoinApplicationModel) FindOne(ctx context.Context, id string) (*JoinApplication, error) {
	query := fmt.Sprintf("SELECT `id`, `user_id`, `company_id`, `invite_code`, `apply_reason`, `status`, `approver_id`, `approve_time`, `approve_note`, `create_time`, `update_time` FROM %s WHERE `id` = ?", m.table)
	var resp JoinApplication
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *defaultJoinApplicationModel) Update(ctx context.Context, data *JoinApplication) error {
	query := fmt.Sprintf("UPDATE %s SET `user_id` = ?, `company_id` = ?, `invite_code` = ?, `apply_reason` = ?, `status` = ?, `approver_id` = ?, `approve_time` = ?, `approve_note` = ?, `update_time` = ? WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.UserId, data.CompanyId, data.InviteCode, data.ApplyReason, data.Status, data.ApproverId, data.ApproveTime, data.ApproveNote, time.Now(), data.Id)
	return err
}

func (m *defaultJoinApplicationModel) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// FindByUserId 根据用户ID查询申请列表
func (m *defaultJoinApplicationModel) FindByUserId(ctx context.Context, userId string) ([]*JoinApplication, error) {
	query := fmt.Sprintf("SELECT `id`, `user_id`, `company_id`, `invite_code`, `apply_reason`, `status`, `approver_id`, `approve_time`, `approve_note`, `create_time`, `update_time` FROM %s WHERE `user_id` = ? ORDER BY `create_time` DESC", m.table)
	var resp []*JoinApplication
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId)
	return resp, err
}

// FindByCompanyId 根据公司ID查询申请列表（可选按状态筛选）
func (m *defaultJoinApplicationModel) FindByCompanyId(ctx context.Context, companyId string, status *int) ([]*JoinApplication, error) {
	var resp []*JoinApplication
	var err error

	if status != nil {
		query := fmt.Sprintf("SELECT `id`, `user_id`, `company_id`, `invite_code`, `apply_reason`, `status`, `approver_id`, `approve_time`, `approve_note`, `create_time`, `update_time` FROM %s WHERE `company_id` = ? AND `status` = ? ORDER BY `create_time` DESC", m.table)
		err = m.conn.QueryRowsCtx(ctx, &resp, query, companyId, *status)
	} else {
		query := fmt.Sprintf("SELECT `id`, `user_id`, `company_id`, `invite_code`, `apply_reason`, `status`, `approver_id`, `approve_time`, `approve_note`, `create_time`, `update_time` FROM %s WHERE `company_id` = ? ORDER BY `create_time` DESC", m.table)
		err = m.conn.QueryRowsCtx(ctx, &resp, query, companyId)
	}

	return resp, err
}

// FindPendingByUserId 查询用户是否有待审批的申请
func (m *defaultJoinApplicationModel) FindPendingByUserId(ctx context.Context, userId string) (*JoinApplication, error) {
	query := fmt.Sprintf("SELECT `id`, `user_id`, `company_id`, `invite_code`, `apply_reason`, `status`, `approver_id`, `approve_time`, `approve_note`, `create_time`, `update_time` FROM %s WHERE `user_id` = ? AND `status` = 0 LIMIT 1", m.table)
	var resp JoinApplication
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateStatus 更新申请状态
func (m *defaultJoinApplicationModel) UpdateStatus(ctx context.Context, id string, status int, approverId, approveNote string) error {
	query := fmt.Sprintf("UPDATE %s SET `status` = ?, `approver_id` = ?, `approve_time` = ?, `approve_note` = ?, `update_time` = ? WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, status, approverId, time.Now(), approveNote, time.Now(), id)
	return err
}
