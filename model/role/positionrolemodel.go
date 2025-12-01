package role

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PositionRoleModel = (*customPositionRoleModel)(nil)

type (
	// PositionRoleModel adds custom query helpers on top of the generated model interface.
	PositionRoleModel interface {
		positionRoleModel
		withSession(session sqlx.Session) PositionRoleModel
		ListByPositionId(ctx context.Context, positionId string) ([]*PositionRole, error)
		ListRolesByPositionId(ctx context.Context, positionId string) ([]*Role, error)
		ListRolesByEmployeeId(ctx context.Context, employeeId string) ([]*Role, error)
	}

	customPositionRoleModel struct {
		*defaultPositionRoleModel
	}
)

// NewPositionRoleModel returns a model for the database table.
func NewPositionRoleModel(conn sqlx.SqlConn) PositionRoleModel {
	return &customPositionRoleModel{
		defaultPositionRoleModel: newPositionRoleModel(conn),
	}
}

func (m *customPositionRoleModel) withSession(session sqlx.Session) PositionRoleModel {
	return NewPositionRoleModel(sqlx.NewSqlConnFromSession(session))
}

// ListByPositionId returns all position-role relations for a position.
func (m *customPositionRoleModel) ListByPositionId(ctx context.Context, positionId string) ([]*PositionRole, error) {
	var list []*PositionRole
	query := fmt.Sprintf("select %s from %s where `position_id` = ? and `status` = 1", positionRoleRows, m.tableName())
	if err := m.conn.QueryRowsCtx(ctx, &list, query, positionId); err != nil {
		return nil, err
	}
	return list, nil
}

// ListRolesByPositionId returns all active role records linked to the given position.
func (m *customPositionRoleModel) ListRolesByPositionId(ctx context.Context, positionId string) ([]*Role, error) {
	var roles []*Role
	roleTable := "`role`"
	prTable := m.tableName()
	query := fmt.Sprintf(`
        select r.id, r.company_id, r.role_name, r.role_code, r.role_description, r.is_system, r.permissions, r.status, r.create_time, r.update_time, r.delete_time
        from %s pr
        join %s r on pr.role_id = r.id
        where pr.position_id = ? and pr.status = 1 and (pr.expire_time is null or pr.expire_time > now()) and r.delete_time is null and r.status = 1`, prTable, roleTable)
	if err := m.conn.QueryRowsCtx(ctx, &roles, query, positionId); err != nil {
		return nil, err
	}
	return roles, nil
}

// ListRolesByEmployeeId returns the roles for the employee via their position.
func (m *customPositionRoleModel) ListRolesByEmployeeId(ctx context.Context, employeeId string) ([]*Role, error) {
	var roles []*Role
	roleTable := "`role`"
	prTable := m.tableName()
	employeeTable := "`employee`"
	query := fmt.Sprintf(`
        select distinct r.id, r.company_id, r.role_name, r.role_code, r.role_description, r.is_system, r.permissions, r.status, r.create_time, r.update_time, r.delete_time
        from %s e
        join %s pr on e.position_id = pr.position_id
        join %s r on pr.role_id = r.id
        where e.id = ? and e.delete_time is null and pr.status = 1 and (pr.expire_time is null or pr.expire_time > now()) and r.delete_time is null and r.status = 1`, employeeTable, prTable, roleTable)
	if err := m.conn.QueryRowsCtx(ctx, &roles, query, employeeId); err != nil {
		return nil, err
	}
	return roles, nil
}
