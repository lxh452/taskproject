// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package upload

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"task_Project/task/internal/utils"
	"time"

	"task_Project/model/upload"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 文件上传（支持图片、PDF、Markdown等文件）
func NewUploadInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadInfoLogic {
	return &UploadInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadInfoLogic) UploadInfo(req *types.UploadInfoRequest, handler *multipart.FileHeader, fromData multipart.File) (resp *types.UploadInfoResponse, err error) {
	// 根据请求的文件类型，进行文件的存储
	fileName := handler.Filename
	fileExt := filepath.Ext(fileName)
	fileType := handler.Header.Get("Content-Type")
	if fileType == "" {
		// 根据扩展名区分type，如果Content-Type为空，则手动推断
		switch fileExt {
		case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
			fileType = "image"
		case ".pdf":
			fileType = "pdf"
		case ".md", ".markdown":
			fileType = "markdown"
		case ".doc", ".docx":
			fileType = "word"
		case ".xls", ".xlsx":
			fileType = "excel"
		case ".ppt", ".pptx":
			fileType = "ppt"
		case ".txt":
			fileType = "text"
		default:
			fileType = "other"
		}
	}
	fileSize := handler.Size
	if err != nil {
		return nil, err
	}
	if req.Category == "" || req.RelatedID == "" || req.Module == "" {
		return nil, errors.New("category or relatedId or module is empty")
		logx.Errorw("category or relatedId or module is empty", logx.Field("category", req.Category), logx.Field("relatedId", req.RelatedID), logx.Field("module", req.Module))
	}
	filePath := fmt.Sprintf("file_upload/%s/%s", req.Module, req.RelatedID)
	fileURL := fmt.Sprintf("https://file.task-project.com/%s/%s", req.Module, req.RelatedID)
	// l.svcCtx.OssService.UploadFile(fileContent,fileExt,fileType,fileSize)
	// 假设已经存好了
	fileID := utils.Common.GenId("file")
	l.svcCtx.UploadFileModel.Insert(l.ctx, &upload.Upload_file{
		FileID:      fileID,
		FileName:    fileName,
		FilePath:    filePath,
		FileURL:     fileURL,
		FileType:    fileType,
		FileSize:    fileSize,
		Module:      req.Module,
		Category:    req.Category,
		RelatedID:   req.RelatedID,
		Description: req.Description,
		Tags:        req.Tags,
		CreateAt:    time.Now(),
		UpdateAt:    time.Now(),
	})
	return &types.UploadInfoResponse{
		FileID:    fileID,
		FileName:  fileName,
		FileURL:   fileURL,
		FileType:  fileType,
		FileSize:  fileSize,
		RelatedID: req.RelatedID,
	}, nil
}
