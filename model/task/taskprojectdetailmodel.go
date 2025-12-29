package task

import "github.com/zeromicro/go-zero/core/stores/mon"

var _ Task_project_detailModel = (*customTask_project_detailModel)(nil)

type (
	// Task_project_detailModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTask_project_detailModel.
	Task_project_detailModel interface {
		task_project_detailModel
	}

	customTask_project_detailModel struct {
		*defaultTask_project_detailModel
	}
)

// NewTask_project_detailModel returns a model for the mongo.
func NewTask_project_detailModel(url, db, collection string) Task_project_detailModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customTask_project_detailModel{
		defaultTask_project_detailModel: newDefaultTask_project_detailModel(conn),
	}
}
