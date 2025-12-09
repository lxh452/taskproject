package task

import "github.com/zeromicro/go-zero/core/stores/mon"

var _ Task_commentModel = (*customTask_commentModel)(nil)

type (
	// Task_commentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTask_commentModel.
	Task_commentModel interface {
		task_commentModel
	}

	customTask_commentModel struct {
		*defaultTask_commentModel
	}
)

// NewTask_commentModel returns a model for the mongo.
func NewTask_commentModel(url, db, collection string) Task_commentModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customTask_commentModel{
		defaultTask_commentModel: newDefaultTask_commentModel(conn),
	}
}

