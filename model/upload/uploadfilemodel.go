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
		FindByUploaderID(ctx context.Context, uploaderID string, page, pageSize int64, fileType, module string) ([]*Upload_file, int64, error)
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

// FindByUploaderID 根据上传者ID查询附件（支持分页和筛选）
func (m *customUpload_fileModel) FindByUploaderID(ctx context.Context, uploaderID string, page, pageSize int64, fileType, module string) ([]*Upload_file, int64, error) {
	filter := bson.M{"uploaderId": uploaderID}

	// 添加文件类型筛选
	if fileType != "" {
		filter["fileType"] = fileType
	}

	// 添加模块筛选
	if module != "" {
		filter["module"] = module
	}

	// 统计总数
	total, err := m.conn.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	skip := (page - 1) * pageSize
	opts := options.Find().SetSort(bson.M{"createAt": -1}).SetSkip(skip).SetLimit(pageSize)

	var data []*Upload_file
	err = m.conn.Find(ctx, &data, filter, opts)
	if err != nil {
		return nil, 0, err
	}

	return data, total, nil
}
