// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package upload

import (
	"context"
	"task_Project/task/internal/utils"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAttachmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除附件
func NewDeleteAttachmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAttachmentLogic {
	return &DeleteAttachmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAttachmentLogic) DeleteAttachment(req *types.DeleteAttachmentRequest) (resp *types.BaseResponse, err error) {
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
