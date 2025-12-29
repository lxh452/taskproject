package upload

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"task_Project/model/upload"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 上传头像
func NewUploadAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadAvatarLogic {
	return &UploadAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadAvatarLogic) UploadAvatar(req *types.UploadAvatarRequest, handler *multipart.FileHeader, fileData multipart.File) (resp *types.UploadAvatarResponse, err error) {
	// 获取当前用户ID
	userID := req.UserID
	if userID == "" {
		currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
		if !ok {
			return nil, fmt.Errorf("无法获取当前用户ID")
		}
		userID = currentUserID
	}

	// 验证文件类型
	fileName := handler.Filename
	fileExt := strings.ToLower(filepath.Ext(fileName))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	}
	if !allowedExts[fileExt] {
		return nil, fmt.Errorf("不支持的图片格式，仅支持 jpg/jpeg/png/gif/webp")
	}

	// 验证文件大小（限制5MB）
	if handler.Size > 5*1024*1024 {
		return nil, fmt.Errorf("图片大小不能超过5MB")
	}

	fileType := "image"
	fileSize := handler.Size
	fileID := utils.Common.GenId("avatar")

	// 先删除旧头像文件和记录（如果有）
	oldFiles, _ := l.svcCtx.UploadFileModel.FindByModuleAndRelatedID(l.ctx, "user", userID)
	for _, oldFile := range oldFiles {
		if oldFile.Category == "avatar" {
			// 删除旧文件
			l.svcCtx.FileStorageService.DeleteFile(oldFile.FilePath)
			// 删除旧记录
			l.svcCtx.UploadFileModel.DeleteByFileID(l.ctx, oldFile.FileID)
			logx.Infof("删除旧头像: fileID=%s, path=%s", oldFile.FileID, oldFile.FilePath)
		}
	}

	// 使用FileStorageService保存头像到本地
	filePath, fileURL, err := l.svcCtx.FileStorageService.SaveFile(
		"user",
		"avatar",
		userID,
		fileID,
		fileName,
		fileData,
	)
	if err != nil {
		logx.Errorf("保存头像文件失败: %v", err)
		return nil, fmt.Errorf("保存头像失败")
	}

	// 存储新头像信息到MongoDB
	err = l.svcCtx.UploadFileModel.Insert(l.ctx, &upload.Upload_file{
		FileID:      fileID,
		FileName:    fileName,
		FilePath:    filePath,
		FileURL:     fileURL,
		FileType:    fileType,
		FileSize:    fileSize,
		Module:      "user",
		Category:    "avatar",
		RelatedID:   userID,
		Description: "用户头像",
		Tags:        "",
		CreateAt:    time.Now(),
		UpdateAt:    time.Now(),
	})
	if err != nil {
		logx.Errorf("存储头像信息到MongoDB失败: %v", err)
		// 删除已保存的文件
		l.svcCtx.FileStorageService.DeleteFile(filePath)
		return nil, fmt.Errorf("存储头像信息失败")
	}

	logx.Infof("头像上传成功: userID=%s, fileID=%s, url=%s", userID, fileID, fileURL)

	return &types.UploadAvatarResponse{
		AvatarURL: fileURL,
		FileID:    fileID,
	}, nil
}
