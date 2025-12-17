package upload

import (
	"context"
	"strings"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAttachmentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取任务附件列表
func NewGetAttachmentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAttachmentsLogic {
	return &GetAttachmentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetTaskAttachments 获取任务的附件列表
func (l *GetAttachmentsLogic) GetTaskAttachments(req *types.GetTaskAttachmentsRequest) (resp *types.BaseResponse, err error) {
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

	// 同时查询任务节点的附件（通过TaskNodeID）
	// 获取任务的所有节点
	taskNodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, req.TaskID)
	if err == nil {
		// 为每个节点查询附件
		for _, node := range taskNodes {
			nodeFiles, err := l.svcCtx.UploadFileModel.FindByTaskNodeID(l.ctx, node.TaskNodeId)
			if err == nil {
				files = append(files, nodeFiles...)
			}
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

// GetTaskNodeAttachments 获取任务节点的附件列表
func (l *GetAttachmentsLogic) GetTaskNodeAttachments(req *types.GetTaskNodeAttachmentsRequest) (resp *types.BaseResponse, err error) {
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

// GetFileDetail 获取文件详情（带权限验证）
func (l *GetAttachmentsLogic) GetFileDetail(req *types.GetFileDetailRequest) (resp *types.BaseResponse, err error) {
	if req.FileID == "" {
		return utils.Response.ValidationError("文件ID不能为空"), nil
	}

	// 获取当前用户ID
	currentUserID, _ := utils.Common.GetCurrentUserID(l.ctx)

	// 从MongoDB查询文件信息
	file, err := l.svcCtx.UploadFileModel.FindByFileID(l.ctx, req.FileID)
	if err != nil {
		logx.Errorf("查询文件详情失败: %v", err)
		return utils.Response.NotFoundError("文件不存在"), nil
	}

	// 如果是任务附件，验证用户是否有权限访问
	if file.Module == "task" && file.TaskNodeID != "" {
		hasAccess := false

		// 如果是上传者，允许访问
		if file.UploaderID == currentUserID {
			hasAccess = true
		}

		// 查询任务节点，验证是否是执行人或负责人
		if !hasAccess && file.TaskNodeID != "" {
			node, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, file.TaskNodeID)
			if err == nil && node != nil {
				// 检查是否是负责人
				if node.LeaderId == currentUserID {
					hasAccess = true
				}
				// 检查是否是执行人
				if !hasAccess && node.ExecutorId != "" {
					executorIDs := strings.Split(node.ExecutorId, ",")
					for _, eid := range executorIDs {
						if strings.TrimSpace(eid) == currentUserID {
							hasAccess = true
							break
						}
					}
				}
			}
		}

		if !hasAccess {
			return utils.Response.ForbiddenError("您没有权限查看此附件"), nil
		}
	}

	return utils.Response.Success(types.AttachmentInfo{
		FileID:      file.FileID,
		FileName:    file.FileName,
		FileURL:     file.FileURL,
		FileType:    file.FileType,
		FileSize:    file.FileSize,
		Module:      file.Module,
		Category:    file.Category,
		RelatedID:   file.RelatedID,
		TaskNodeID:  file.TaskNodeID,
		UploaderID:  file.UploaderID,
		Description: file.Description,
		Tags:        file.Tags,
		CreateTime:  file.CreateAt.Format(time.RFC3339),
		UpdateTime:  file.UpdateAt.Format(time.RFC3339),
	}), nil
}

// DeleteAttachment 删除附件
func (l *GetAttachmentsLogic) DeleteAttachment(req *types.DeleteAttachmentRequest) (resp *types.BaseResponse, err error) {
	if req.FileID == "" {
		return utils.Response.ValidationError("文件ID不能为空"), nil
	}

	// 先查询附件信息获取文件路径
	fileInfo, err := l.svcCtx.UploadFileModel.FindByFileID(l.ctx, req.FileID)
	if err != nil {
		logx.Errorf("查询附件信息失败: %v", err)
		return utils.Response.NotFoundError("附件不存在"), nil
	}

	// 删除物理文件
	if fileInfo.FilePath != "" {
		if err := l.svcCtx.FileStorageService.DeleteFile(fileInfo.FilePath); err != nil {
			logx.Errorf("删除物理文件失败: %v", err)
			// 继续删除数据库记录
		}
	}

	// 从MongoDB删除附件记录
	_, err = l.svcCtx.UploadFileModel.DeleteByFileID(l.ctx, req.FileID)
	if err != nil {
		logx.Errorf("删除附件记录失败: %v", err)
		return utils.Response.InternalError("删除附件失败"), nil
	}

	logx.Infof("附件删除成功: fileID=%s, path=%s", req.FileID, fileInfo.FilePath)
	return utils.Response.Success(nil), nil
}
