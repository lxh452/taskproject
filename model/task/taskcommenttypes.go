package task

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Task_comment 任务评论模型(存储到MongoDB)
// 用于存储任务节点的评论、@提及等信息
type Task_comment struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	CommentID       string        `bson:"commentId" json:"commentId" index:"commentId"`    // 评论ID
	TaskID          string        `bson:"taskId" json:"taskId" index:"taskId"`             // 任务ID
	TaskNodeID      string        `bson:"taskNodeId" json:"taskNodeId" index:"taskNodeId"` // 任务节点ID
	UserID          string        `bson:"userId" json:"userId" index:"userId"`             // 评论用户ID
	EmployeeID      string        `bson:"employeeId" json:"employeeId" index:"employeeId"` // 员工ID
	EmployeeName    string        `bson:"employeeName" json:"employeeName"`                // 员工姓名
	Content         string        `bson:"content" json:"content"`                          // 评论内容
	ContentHTML     string        `bson:"contentHtml" json:"contentHtml"`                  // HTML格式内容(支持富文本)
	AtEmployeeIDs   []string      `bson:"atEmployeeIds" json:"atEmployeeIds"`              // @的员工ID列表
	AtEmployeeNames []string      `bson:"atEmployeeNames" json:"atEmployeeNames"`          // @的员工姓名列表
	ParentID        string        `bson:"parentId" json:"parentId"`                        // 父评论ID(用于回复)
	ReplyToUserID   string        `bson:"replyToUserId" json:"replyToUserId"`              // 回复的用户ID
	ReplyToName     string        `bson:"replyToName" json:"replyToName"`                  // 回复的用户姓名
	AttachmentIDs   []string      `bson:"attachmentIds" json:"attachmentIds"`              // 附件ID列表
	AttachmentURLs  []string      `bson:"attachmentUrls" json:"attachmentUrls"`            // 附件URL列表
	LikeCount       int64         `bson:"likeCount" json:"likeCount"`                      // 点赞数
	LikedBy         []string      `bson:"likedBy" json:"likedBy"`                          // 点赞的用户ID列表
	IsDeleted       bool          `bson:"isDeleted" json:"isDeleted"`                      // 是否已删除
	DeletedAt       time.Time     `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`  // 删除时间
	UpdateAt        time.Time     `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt        time.Time     `bson:"createAt,omitempty" json:"createAt,omitempty" index:"createAt"`
}
