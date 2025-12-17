package svc

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/zeromicro/go-zero/core/logx"
)

// COSStorageService 腾讯云COS存储服务
type COSStorageService struct {
	client    *cos.Client
	bucket    string
	region    string
	urlPrefix string
}

// NewCOSStorageService 创建COS存储服务
func NewCOSStorageService(secretId, secretKey, bucket, region, urlPrefix string) (*COSStorageService, error) {
	// 验证配置
	if secretId == "" {
		return nil, fmt.Errorf("SecretId不能为空")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("SecretKey不能为空")
	}
	if bucket == "" {
		return nil, fmt.Errorf("Bucket不能为空")
	}
	if region == "" {
		return nil, fmt.Errorf("Region不能为空")
	}

	// 构建COS客户端
	u, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region))
	if err != nil {
		return nil, fmt.Errorf("解析COS URL失败: %v", err)
	}
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretId,
			SecretKey: secretKey,
		},
	})

	// 如果urlPrefix为空，使用默认的COS域名
	if urlPrefix == "" {
		urlPrefix = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region)
	}

	return &COSStorageService{
		client:    client,
		bucket:    bucket,
		region:    region,
		urlPrefix: urlPrefix,
	}, nil
}

// SaveFile 保存文件到COS
// module: 模块名称 (task/user/avatar等)
// category: 分类 (attachment/avatar/document等)
// relatedID: 关联ID
// fileID: 文件ID
// fileName: 原始文件名
// file: 文件数据
// 返回: 文件保存路径(COS Key), 访问URL, 错误
func (s *COSStorageService) SaveFile(module, category, relatedID, fileID, fileName string, file multipart.File) (string, string, error) {
	// 生成COS对象键（Key）
	ext := filepath.Ext(fileName)
	timestamp := time.Now().Format("20060102150405")
	cleanName := sanitizeFileName(fileName)

	// 构建COS Key: module/category/relatedID/fileID_timestamp_filename.ext
	key := fmt.Sprintf("%s/%s/%s/%s_%s%s", module, category, relatedID, fileID, timestamp, ext)
	if cleanName != "" && len(cleanName) <= 50 {
		key = fmt.Sprintf("%s/%s/%s/%s_%s_%s%s", module, category, relatedID, fileID, timestamp, strings.TrimSuffix(cleanName, ext), ext)
	}

	// 上传文件到COS
	ctx := context.Background()
	_, err := s.client.Object.Put(ctx, key, file, nil)
	if err != nil {
		logx.Errorf("上传文件到COS失败: key=%s, error=%v", key, err)
		return "", "", fmt.Errorf("上传文件到COS失败: %v", err)
	}

	// 生成访问URL
	fileURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(s.urlPrefix, "/"), key)

	logx.Infof("文件上传到COS成功: key=%s, url=%s", key, fileURL)
	return key, fileURL, nil
}

// SaveFileFromBytes 从字节数据保存文件到COS
func (s *COSStorageService) SaveFileFromBytes(module, category, relatedID, fileID, fileName string, data []byte) (string, string, error) {
	// 生成COS对象键（Key）
	ext := filepath.Ext(fileName)
	timestamp := time.Now().Format("20060102150405")

	// 构建COS Key
	key := fmt.Sprintf("%s/%s/%s/%s_%s%s", module, category, relatedID, fileID, timestamp, ext)

	// 上传文件到COS
	ctx := context.Background()
	_, err := s.client.Object.Put(ctx, key, strings.NewReader(string(data)), &cos.ObjectPutOptions{})
	if err != nil {
		logx.Errorf("上传文件到COS失败: key=%s, error=%v", key, err)
		return "", "", fmt.Errorf("上传文件到COS失败: %v", err)
	}

	// 生成访问URL
	fileURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(s.urlPrefix, "/"), key)

	logx.Infof("文件上传到COS成功: key=%s, url=%s", key, fileURL)
	return key, fileURL, nil
}

// DeleteFile 从COS删除文件
func (s *COSStorageService) DeleteFile(key string) error {
	if key == "" {
		return nil
	}

	ctx := context.Background()
	_, err := s.client.Object.Delete(ctx, key)
	if err != nil {
		logx.Errorf("从COS删除文件失败: key=%s, error=%v", key, err)
		return fmt.Errorf("删除文件失败: %v", err)
	}

	logx.Infof("文件从COS删除成功: key=%s", key)
	return nil
}

// GetFileURL 获取文件访问URL
func (s *COSStorageService) GetFileURL(key string) string {
	if key == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.urlPrefix, "/"), key)
}

// GetFile 从COS获取文件内容
func (s *COSStorageService) GetFile(key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("文件Key不能为空")
	}

	ctx := context.Background()
	resp, err := s.client.Object.Get(ctx, key, nil)
	if err != nil {
		logx.Errorf("从COS获取文件失败: key=%s, error=%v", key, err)
		return nil, fmt.Errorf("获取文件失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取文件内容
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logx.Errorf("读取文件内容失败: key=%s, error=%v", key, err)
		return nil, fmt.Errorf("读取文件内容失败: %v", err)
	}

	return data, nil
}
