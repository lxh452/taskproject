package upload

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Attachment_comment 附件评论标注模型(存储到MongoDB)
// 用于存储附件上的评论、标注信息
type Attachment_comment struct {
	ID              bson.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	CommentID       string          `bson:"commentId" json:"commentId" index:"commentId"`     // 评论ID
	FileID          string          `bson:"fileId" json:"fileId" index:"fileId"`              // 附件文件ID
	TaskID          string          `bson:"taskId" json:"taskId" index:"taskId"`              // 关联任务ID
	TaskNodeID      string          `bson:"taskNodeId" json:"taskNodeId" index:"taskNodeId"`  // 关联任务节点ID
	UserID          string          `bson:"userId" json:"userId" index:"userId"`              // 评论用户ID
	EmployeeID      string          `bson:"employeeId" json:"employeeId"`                     // 员工ID
	EmployeeName    string          `bson:"employeeName" json:"employeeName"`                 // 员工姓名
	Content         string          `bson:"content" json:"content"`                           // 评论内容
	AtEmployeeIDs   []string        `bson:"atEmployeeIds" json:"atEmployeeIds"`               // @的员工ID列表
	AtEmployeeNames []string        `bson:"atEmployeeNames" json:"atEmployeeNames"`           // @的员工姓名列表
	AnnotationType  string          `bson:"annotationType" json:"annotationType"`             // 标注类型: point/rect/highlight/arrow
	AnnotationData  *AnnotationData `bson:"annotationData" json:"annotationData"`             // 标注数据
	PageNumber      int             `bson:"pageNumber" json:"pageNumber"`                     // PDF页码(如果是PDF文件)
	ParentID        string          `bson:"parentId" json:"parentId"`                         // 父评论ID(用于回复)
	ReplyToUserID   string          `bson:"replyToUserId" json:"replyToUserId"`               // 回复的用户ID
	ReplyToName     string          `bson:"replyToName" json:"replyToName"`                   // 回复的用户姓名
	IsResolved      bool            `bson:"isResolved" json:"isResolved"`                     // 是否已解决
	ResolvedBy      string          `bson:"resolvedBy" json:"resolvedBy"`                     // 解决人ID
	ResolvedAt      time.Time       `bson:"resolvedAt,omitempty" json:"resolvedAt,omitempty"` // 解决时间
	IsDeleted       bool            `bson:"isDeleted" json:"isDeleted"`                       // 是否已删除
	DeletedAt       time.Time       `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`   // 删除时间
	UpdateAt        time.Time       `bson:"updateAt,omitempty" json:"updateAt,omitempty"`
	CreateAt        time.Time       `bson:"createAt,omitempty" json:"createAt,omitempty" index:"createAt"`
}

// AnnotationData 标注数据
type AnnotationData struct {
	X      float64 `bson:"x" json:"x"`           // X坐标(相对于图片/文档宽度的百分比)
	Y      float64 `bson:"y" json:"y"`           // Y坐标(相对于图片/文档高度的百分比)
	Width  float64 `bson:"width" json:"width"`   // 宽度(矩形标注)
	Height float64 `bson:"height" json:"height"` // 高度(矩形标注)
	Color  string  `bson:"color" json:"color"`   // 标注颜色
	Text   string  `bson:"text" json:"text"`     // 高亮文本内容
	StartX float64 `bson:"startX" json:"startX"` // 箭头起点X
	StartY float64 `bson:"startY" json:"startY"` // 箭头起点Y
	EndX   float64 `bson:"endX" json:"endX"`     // 箭头终点X
	EndY   float64 `bson:"endY" json:"endY"`     // 箭头终点Y
}
