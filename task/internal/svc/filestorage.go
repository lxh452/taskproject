package svc

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// FileStorageService 文件存储服务
// 目前实现本地存储，后续可以替换为COS/OSS等云存储
type FileStorageService struct {
	// 本地存储根目录
	StorageRoot string
	// 访问URL前缀
	URLPrefix string
}

// NewFileStorageService 创建文件存储服务
func NewFileStorageService(storageRoot, urlPrefix string) *FileStorageService {
	// 确保存储目录存在
	if err := os.MkdirAll(storageRoot, 0755); err != nil {
		logx.Errorf("创建存储目录失败: %v", err)
	}
	return &FileStorageService{
		StorageRoot: storageRoot,
		URLPrefix:   urlPrefix,
	}
}

// SaveFile 保存文件到本地
// module: 模块名称 (task/user/avatar等)
// category: 分类 (attachment/avatar/document等)
// relatedID: 关联ID
// fileID: 文件ID
// fileName: 原始文件名
// file: 文件数据
// 返回: 文件保存路径, 访问URL, 错误
func (s *FileStorageService) SaveFile(module, category, relatedID, fileID, fileName string, file multipart.File) (string, string, error) {
	// 生成存储路径: storage_root/module/category/relatedID/
	dir := filepath.Join(s.StorageRoot, module, category, relatedID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成文件名: fileID_timestamp_originalName
	ext := filepath.Ext(fileName)
	timestamp := time.Now().Format("20060102150405")
	// 清理原始文件名，移除特殊字符
	cleanName := sanitizeFileName(fileName)
	newFileName := fmt.Sprintf("%s_%s%s", fileID, timestamp, ext)
	if cleanName != "" && len(cleanName) <= 50 {
		newFileName = fmt.Sprintf("%s_%s_%s%s", fileID, timestamp, strings.TrimSuffix(cleanName, ext), ext)
	}

	// 完整文件路径
	filePath := filepath.Join(dir, newFileName)

	// 创建目标文件
	dst, err := os.Create(filePath)
	if err != nil {
		return "", "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		return "", "", fmt.Errorf("保存文件失败: %v", err)
	}

	// 生成访问URL
	relativePath := filepath.Join(module, category, relatedID, newFileName)
	// 将Windows路径分隔符转换为URL分隔符
	urlPath := strings.ReplaceAll(relativePath, "\\", "/")
	fileURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(s.URLPrefix, "/"), urlPath)

	logx.Infof("文件保存成功: path=%s, url=%s", filePath, fileURL)
	return filePath, fileURL, nil
}

// SaveFileFromBytes 从字节数据保存文件
func (s *FileStorageService) SaveFileFromBytes(module, category, relatedID, fileID, fileName string, data []byte) (string, string, error) {
	// 生成存储路径
	dir := filepath.Join(s.StorageRoot, module, category, relatedID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 生成文件名
	ext := filepath.Ext(fileName)
	timestamp := time.Now().Format("20060102150405")
	newFileName := fmt.Sprintf("%s_%s%s", fileID, timestamp, ext)

	// 完整文件路径
	filePath := filepath.Join(dir, newFileName)

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", "", fmt.Errorf("保存文件失败: %v", err)
	}

	// 生成访问URL
	relativePath := filepath.Join(module, category, relatedID, newFileName)
	urlPath := strings.ReplaceAll(relativePath, "\\", "/")
	fileURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(s.URLPrefix, "/"), urlPath)

	logx.Infof("文件保存成功: path=%s, url=%s", filePath, fileURL)
	return filePath, fileURL, nil
}

// DeleteFile 删除文件
func (s *FileStorageService) DeleteFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	// 如果是完整路径，直接删除
	if filepath.IsAbs(filePath) {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除文件失败: %v", err)
		}
		return nil
	}

	// 否则拼接存储根目录
	fullPath := filepath.Join(s.StorageRoot, filePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	logx.Infof("文件删除成功: path=%s", fullPath)
	return nil
}

// GetFilePath 获取文件完整路径
func (s *FileStorageService) GetFilePath(relativePath string) string {
	return filepath.Join(s.StorageRoot, relativePath)
}

// sanitizeFileName 清理文件名，移除特殊字符
func sanitizeFileName(fileName string) string {
	// 移除路径分隔符和特殊字符
	fileName = filepath.Base(fileName)
	// 替换特殊字符
	replacer := strings.NewReplacer(
		" ", "_",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"{", "",
		"}", "",
		"<", "",
		">", "",
		":", "",
		";", "",
		"'", "",
		"\"", "",
		"|", "",
		"?", "",
		"*", "",
		"&", "",
		"#", "",
		"%", "",
		"$", "",
		"@", "",
		"!", "",
		"=", "",
		"+", "",
		"`", "",
		"~", "",
		"^", "",
	)
	return replacer.Replace(fileName)
}

