// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package upload

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
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

func (l *UploadInfoLogic) UploadInfo(req *types.UploadInfoRequest, handler *multipart.FileHeader, fileData multipart.File) (resp *types.UploadInfoResponse, err error) {
	// 获取文件信息
	fileName := handler.Filename
	fileExt := strings.ToLower(filepath.Ext(fileName))
	fileType := l.getFileType(fileExt, handler.Header.Get("Content-Type"))
	fileSize := handler.Size

	// 获取当前用户ID
	uploaderID, _ := utils.Common.GetCurrentUserID(l.ctx)

	// 验证必要参数
	if req.Module == "" {
		req.Module = "general"
	}
	if req.Category == "" {
		req.Category = "attachment"
	}
	if req.RelatedID == "" {
		req.RelatedID = "default"
	}

	// 任务附件必须选择节点
	if req.Module == "task" && req.Category == "attachment" && req.TaskNodeID == "" {
		return nil, errors.New("任务附件必须关联到具体的任务节点")
	}

	// 文件大小限制（50MB）
	if fileSize > 50*1024*1024 {
		return nil, errors.New("文件大小不能超过50MB")
	}

	// 生成文件ID
	fileID := utils.Common.GenId("file")

	// 使用FileStorageService保存文件
	filePath, fileURL, err := l.svcCtx.FileStorageService.SaveFile(
		req.Module,
		req.Category,
		req.RelatedID,
		fileID,
		fileName,
		fileData,
	)
	if err != nil {
		logx.Errorf("保存文件失败: %v", err)
		return nil, fmt.Errorf("文件保存失败: %v", err)
	}

	// 保存文件信息到MongoDB
	err = l.svcCtx.UploadFileModel.Insert(l.ctx, &upload.Upload_file{
		FileID:      fileID,
		FileName:    fileName,
		FilePath:    filePath,
		FileURL:     fileURL,
		FileType:    fileType,
		FileSize:    fileSize,
		Module:      req.Module,
		Category:    req.Category,
		RelatedID:   req.RelatedID,
		TaskNodeID:  req.TaskNodeID,
		UploaderID:  uploaderID,
		Description: req.Description,
		Tags:        req.Tags,
		CreateAt:    time.Now(),
		UpdateAt:    time.Now(),
	})
	if err != nil {
		logx.Errorf("保存文件信息到MongoDB失败: %v", err)
		// 删除已保存的文件
		l.svcCtx.FileStorageService.DeleteFile(filePath)
		return nil, errors.New("保存文件信息失败")
	}

	logx.Infof("文件上传成功: fileID=%s, fileName=%s, module=%s, relatedID=%s, taskNodeID=%s",
		fileID, fileName, req.Module, req.RelatedID, req.TaskNodeID)

	return &types.UploadInfoResponse{
		FileID:    fileID,
		FileName:  fileName,
		FileURL:   fileURL,
		FileType:  fileType,
		FileSize:  fileSize,
		RelatedID: req.RelatedID,
	}, nil
}

// getFileType 根据扩展名和Content-Type获取文件类型
func (l *UploadInfoLogic) getFileType(ext, contentType string) string {
	// 先根据扩展名判断
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico":
		return "image"
	case ".pdf":
		return "pdf"
	case ".md", ".markdown":
		return "markdown"
	case ".doc", ".docx":
		return "word"
	case ".xls", ".xlsx":
		return "excel"
	case ".ppt", ".pptx":
		return "ppt"
	case ".txt", ".log":
		return "text"
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv":
		return "video"
	case ".mp3", ".wav", ".flac", ".aac", ".ogg":
		return "audio"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "archive"
	case ".json", ".xml", ".yaml", ".yml":
		return "config"
	case ".go", ".py", ".js", ".ts", ".java", ".c", ".cpp", ".h", ".css", ".html":
		return "code"
	}

	// 根据Content-Type判断
	if strings.HasPrefix(contentType, "image/") {
		return "image"
	}
	if strings.HasPrefix(contentType, "video/") {
		return "video"
	}
	if strings.HasPrefix(contentType, "audio/") {
		return "audio"
	}
	if strings.Contains(contentType, "pdf") {
		return "pdf"
	}
	if strings.Contains(contentType, "word") || strings.Contains(contentType, "document") {
		return "word"
	}
	if strings.Contains(contentType, "excel") || strings.Contains(contentType, "spreadsheet") {
		return "excel"
	}
	if strings.Contains(contentType, "powerpoint") || strings.Contains(contentType, "presentation") {
		return "ppt"
	}

	return "other"
}
