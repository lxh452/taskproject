package task

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Task_project_detail struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TaskID   string        `bson:"taskId" json:"taskId" index:"taskId"` //设置索引
	FileID   string        `bson:"fileId" json:"fileId" index:"fileId"` //设置索引
	FileName string        `bson:"fileName" json:"fileName"`
	FileURL  string        `bson:"fileUrl" json:"fileUrl"`
	FileType string        `bson:"fileType" json:"fileType" index:"fileType"` //设置索引
	FileSize int64         `bson:"fileSize" json:"fileSize"`
	UpdateAt time.Time     `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt time.Time     `bson:"createAt,omitempty" json:"createAt,omitempty" index:"createAt"` //设置索引
}
