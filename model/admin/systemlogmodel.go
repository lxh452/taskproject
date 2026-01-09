package admin

import "github.com/zeromicro/go-zero/core/stores/mon"

var _ SystemLogModel = (*customSystemLogModel)(nil)

type (
	// SystemLogModel is an interface to be customized
	SystemLogModel interface {
		systemLogModel
	}

	customSystemLogModel struct {
		*defaultSystemLogModel
	}
)

// NewSystemLogModel returns a model for the mongo.
func NewSystemLogModel(url, db, collection string) SystemLogModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customSystemLogModel{
		defaultSystemLogModel: newDefaultSystemLogModel(conn),
	}
}
