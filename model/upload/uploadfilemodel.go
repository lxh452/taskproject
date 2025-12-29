package upload

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var _ Upload_fileModel = (*customUpload_fileModel)(nil)

type (
	// Upload_fileModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUpload_fileModel.
	Upload_fileModel interface {
		upload_fileModel
		FindByTaskNodeID(ctx context.Context, taskNodeID string) ([]*Upload_file, error)
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

// FindByTaskNodeID 根据任务节点ID查询附件
func (m *customUpload_fileModel) FindByTaskNodeID(ctx context.Context, taskNodeID string) ([]*Upload_file, error) {
	filter := bson.M{"taskNodeId": taskNodeID}
	opts := options.Find().SetSort(bson.M{"createAt": -1})

	var data []*Upload_file
	err := m.conn.Find(ctx, &data, filter, opts)
	if err != nil {
		return nil, err
	}

	return data, nil
}
