// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package upload

import (
	"context"
	"strings"
	"task_Project/task/internal/utils"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFileDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取文件详情
func NewGetFileDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFileDetailLogic {
	return &GetFileDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFileDetailLogic) GetFileDetail(req *types.GetFileDetailRequest) (resp *types.BaseResponse, err error) {
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
