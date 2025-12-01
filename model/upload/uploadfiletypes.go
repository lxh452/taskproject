package upload

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Upload_file struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FileID      string        `bson:"fileId" json:"fileId"`
	FileName    string        `bson:"fileName" json:"fileName"`
	FilePath    string        `bson:"filePath" json:"filePath"`
	FileURL     string        `bson:"fileUrl" json:"fileUrl"`
	FileType    string        `bson:"fileType" json:"fileType"`
	FileSize    int64         `bson:"fileSize" json:"fileSize"`
	Module      string        `bson:"module" json:"module"`
	Category    string        `bson:"category" json:"category"`
	RelatedID   string        `bson:"relatedId" json:"relatedId"`
	Description string        `bson:"description" json:"description"`
	Tags        string        `bson:"tags" json:"tags"`
	UpdateAt    time.Time     `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt    time.Time     `bson:"createAt,omitempty" json:"createAt,omitempty"`
}
