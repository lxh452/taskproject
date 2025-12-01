package upload

import "github.com/zeromicro/go-zero/core/stores/mon"

var _ Upload_fileModel = (*customUpload_fileModel)(nil)

type (
	// Upload_fileModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUpload_fileModel.
	Upload_fileModel interface {
		upload_fileModel
	}

	customUpload_fileModel struct {
		*defaultUpload_fileModel
	}
)

// NewUpload_fileModel returns a model for the mongo.
func NewUpload_fileModel(url, db, collection string) Upload_fileModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customUpload_fileModel{
		defaultUpload_fileModel: newDefaultUpload_fileModel(conn),
	}
}
