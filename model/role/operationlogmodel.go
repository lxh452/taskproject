package role

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ OperationLogModel = (*customOperationLogModel)(nil)

type (
	// OperationLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customOperationLogModel.
	OperationLogModel interface {
		operationLogModel
		withSession(session sqlx.Session) OperationLogModel
	}

	customOperationLogModel struct {
		*defaultOperationLogModel
	}
)

// NewOperationLogModel returns a model for the database table.
func NewOperationLogModel(conn sqlx.SqlConn) OperationLogModel {
	return &customOperationLogModel{
		defaultOperationLogModel: newOperationLogModel(conn),
	}
}

func (m *customOperationLogModel) withSession(session sqlx.Session) OperationLogModel {
	return NewOperationLogModel(sqlx.NewSqlConnFromSession(session))
}
