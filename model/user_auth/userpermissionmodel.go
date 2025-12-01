package user_auth

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserPermissionModel = (*customUserPermissionModel)(nil)

type (
	// UserPermissionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserPermissionModel.
	UserPermissionModel interface {
		userPermissionModel
		withSession(session sqlx.Session) UserPermissionModel
		DeleteByUserIdAndGrantType(ctx context.Context, userId string, grantType int64) error
	}

	customUserPermissionModel struct {
		*defaultUserPermissionModel
	}
)

// NewUserPermissionModel returns a model for the database table.
func NewUserPermissionModel(conn sqlx.SqlConn) UserPermissionModel {
	return &customUserPermissionModel{
		defaultUserPermissionModel: newUserPermissionModel(conn),
	}
}

func (m *customUserPermissionModel) withSession(session sqlx.Session) UserPermissionModel {
	return NewUserPermissionModel(sqlx.NewSqlConnFromSession(session))
}

// DeleteByUserIdAndGrantType 根据用户ID和授权类型删除权限
func (m *customUserPermissionModel) DeleteByUserIdAndGrantType(ctx context.Context, userId string, grantType int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE `user_id` = ? AND `grant_type` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, userId, grantType)
	return err
}
