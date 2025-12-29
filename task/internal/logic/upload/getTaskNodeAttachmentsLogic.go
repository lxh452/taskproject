// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package upload

import (
	"context"
	"task_Project/task/internal/utils"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskNodeAttachmentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取任务节点附件列表
func NewGetTaskNodeAttachmentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskNodeAttachmentsLogic {
	return &GetTaskNodeAttachmentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTaskNodeAttachmentsLogic) GetTaskNodeAttachments(req *types.GetTaskNodeAttachmentsRequest) (resp *types.BaseResponse, err error) {
	if req.TaskNodeID == "" {
		return utils.Response.ValidationError("任务节点ID不能为空"), nil
	}

	// 从MongoDB查询任务节点相关的附件
	// 优先通过TaskNodeID查询（这是主要方式）
	files, err := l.svcCtx.UploadFileModel.FindByTaskNodeID(l.ctx, req.TaskNodeID)
	if err != nil {
		logx.Errorf("查询任务节点附件失败: %v", err)
		// 如果通过TaskNodeID查询失败，尝试通过module和relatedId查询（兼容旧数据）
		files, err = l.svcCtx.UploadFileModel.FindByModuleAndRelatedID(l.ctx, "tasknode", req.TaskNodeID)
		if err != nil {
			logx.Errorf("查询任务节点附件失败（兼容查询）: %v", err)
			return utils.Response.InternalError("查询附件失败"), nil
		}
	}

	// 转换为响应格式
	attachments := make([]types.AttachmentInfo, 0, len(files))
	for _, f := range files {
		attachments = append(attachments, types.AttachmentInfo{
			FileID:      f.FileID,
			FileName:    f.FileName,
			FileURL:     f.FileURL,
			FileType:    f.FileType,
			FileSize:    f.FileSize,
			Module:      f.Module,
			Category:    f.Category,
			RelatedID:   f.RelatedID,
			Description: f.Description,
			Tags:        f.Tags,
			CreateTime:  f.CreateAt.Format(time.RFC3339),
			UpdateTime:  f.UpdateAt.Format(time.RFC3339),
		})
	}

	return utils.Response.Success(map[string]interface{}{
		"list":  attachments,
		"total": len(attachments),
	}), nil
}
