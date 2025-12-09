package upload

import "github.com/zeromicro/go-zero/core/stores/mon"

var _ Attachment_commentModel = (*customAttachment_commentModel)(nil)

type (
	// Attachment_commentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAttachment_commentModel.
	Attachment_commentModel interface {
		attachment_commentModel
	}

	customAttachment_commentModel struct {
		*defaultAttachment_commentModel
	}
)

// NewAttachment_commentModel returns a model for the mongo.
func NewAttachment_commentModel(url, db, collection string) Attachment_commentModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customAttachment_commentModel{
		defaultAttachment_commentModel: newDefaultAttachment_commentModel(conn),
	}
}

