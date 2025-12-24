// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package upload

import (
	"context"
	"task_Project/model/upload"
	"task_Project/task/internal/utils"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskAttachmentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取任务附件列表
func NewGetTaskAttachmentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskAttachmentsLogic {
	return &GetTaskAttachmentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTaskAttachmentsLogic) GetTaskAttachments(req *types.GetTaskAttachmentsRequest) (resp *types.BaseResponse, err error) {
	if req.TaskID == "" {
		return utils.Response.ValidationError("任务ID不能为空"), nil
	}

	// 从MongoDB查询任务相关的附件
	// 查询条件：module="task" 且 relatedId=taskID
	files, err := l.svcCtx.UploadFileModel.FindByModuleAndRelatedID(l.ctx, "task", req.TaskID)
	if err != nil {
		logx.Errorf("查询任务附件失败: %v", err)
		return utils.Response.InternalError("查询附件失败"), nil
	}

	// 使用map去重，避免重复显示
	fileMap := make(map[string]*upload.Upload_file)
	for _, f := range files {
		fileMap[f.FileID] = f
	}

	// 同时查询任务节点的附件（通过TaskNodeID）
	// 获取任务的所有节点
	taskNodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, req.TaskID)
	if err == nil {
		// 为每个节点查询附件
		for _, node := range taskNodes {
			nodeFiles, err := l.svcCtx.UploadFileModel.FindByTaskNodeID(l.ctx, node.TaskNodeId)
			if err == nil {
				// 只添加不重复的文件
				for _, f := range nodeFiles {
					if _, exists := fileMap[f.FileID]; !exists {
						fileMap[f.FileID] = f
					}
				}
			}
		}
	}

	// 将map转换为slice
	uniqueFiles := make([]*upload.Upload_file, 0, len(fileMap))
	for _, f := range fileMap {
		uniqueFiles = append(uniqueFiles, f)
	}

	// 转换为响应格式
	attachments := make([]types.AttachmentInfo, 0, len(uniqueFiles))
	for _, f := range uniqueFiles {
		attachments = append(attachments, types.AttachmentInfo{
			FileID:      f.FileID,
			FileName:    f.FileName,
			FileURL:     f.FileURL,
			FileType:    f.FileType,
			FileSize:    f.FileSize,
			Module:      f.Module,
			Category:    f.Category,
			RelatedID:   f.RelatedID,
			TaskNodeID:  f.TaskNodeID,
			UploaderID:  f.UploaderID,
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
