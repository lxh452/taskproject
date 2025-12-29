package role

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SystemConfigModel = (*customSystemConfigModel)(nil)

type (
	// SystemConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSystemConfigModel.
	SystemConfigModel interface {
		systemConfigModel
		withSession(session sqlx.Session) SystemConfigModel
	}

	customSystemConfigModel struct {
		*defaultSystemConfigModel
	}
)

// NewSystemConfigModel returns a model for the database table.
func NewSystemConfigModel(conn sqlx.SqlConn) SystemConfigModel {
	return &customSystemConfigModel{
		defaultSystemConfigModel: newSystemConfigModel(conn),
	}
}

func (m *customSystemConfigModel) withSession(session sqlx.Session) SystemConfigModel {
	return NewSystemConfigModel(sqlx.NewSqlConnFromSession(session))
}
