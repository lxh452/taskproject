package svc

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// TransactionService 事务服务
type TransactionService struct {
	conn sqlx.SqlConn
}

// NewTransactionService 创建事务服务
func NewTransactionService(conn sqlx.SqlConn) *TransactionService {
	return &TransactionService{conn: conn}
}

// Transaction 执行事务
func (s *TransactionService) Transaction(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error {
	return s.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		return fn(ctx, session)
	})
}

// FileStorageInterface 文件存储接口
type FileStorageInterface interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (string, error)
	Download(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	GetURL(key string) string
	// 扩展方法
	SaveFile(module, category, relatedID, fileID, fileName string, data io.Reader) (filePath, fileURL string, err error)
	DeleteFile(filePath string) error
	GetFile(filePath string) ([]byte, error)
}

// SQLExecutorService SQL执行服务
type SQLExecutorService struct {
	conn   sqlx.SqlConn
	sqlDir string
}

// NewSQLExecutorService 创建SQL执行服务
func NewSQLExecutorService(conn sqlx.SqlConn, sqlDir string) *SQLExecutorService {
	return &SQLExecutorService{conn: conn, sqlDir: sqlDir}
}

// Exec 执行SQL
func (s *SQLExecutorService) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.conn.ExecCtx(ctx, query, args...)
}

// Query 查询SQL
func (s *SQLExecutorService) Query(ctx context.Context, v interface{}, query string, args ...interface{}) error {
	return s.conn.QueryRowCtx(ctx, v, query, args...)
}

// AutoMigrate 自动执行数据库迁移
func (s *SQLExecutorService) AutoMigrate(ctx context.Context) error {
	// 检查SQL目录是否存在
	if _, err := os.Stat(s.sqlDir); os.IsNotExist(err) {
		logx.Infof("SQL目录不存在，跳过迁移: %s", s.sqlDir)
		return nil
	}

	// 读取SQL文件
	files, err := filepath.Glob(filepath.Join(s.sqlDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("读取SQL文件失败: %v", err)
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			logx.Errorf("读取SQL文件失败 %s: %v", file, err)
			continue
		}

		// 分割SQL语句并执行
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := s.conn.ExecCtx(ctx, stmt); err != nil {
				logx.Errorf("执行SQL失败 %s: %v", file, err)
			}
		}
		logx.Infof("执行SQL迁移文件: %s", filepath.Base(file))
	}

	return nil
}

// COSStorageService COS存储服务
type COSStorageService struct {
	secretID  string
	secretKey string
	bucket    string
	region    string
	urlPrefix string
	client    *cos.Client
}

// NewCOSStorageService 创建COS存储服务
func NewCOSStorageService(secretID, secretKey, bucket, region, urlPrefix string) (*COSStorageService, error) {
	if secretID == "" || secretKey == "" || bucket == "" || region == "" {
		return nil, fmt.Errorf("COS配置不完整")
	}

	// 构建COS客户端
	cosURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region)
	u, err := url.Parse(cosURL)
	if err != nil {
		return nil, fmt.Errorf("解析COS URL失败: %v", err)
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})

	return &COSStorageService{
		secretID:  secretID,
		secretKey: secretKey,
		bucket:    bucket,
		region:    region,
		urlPrefix: urlPrefix,
		client:    client,
	}, nil
}

// Upload 上传文件
func (s *COSStorageService) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("COS客户端未初始化")
	}

	// 构建上传选项
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}

	// 如果未指定ContentType，尝试从文件名推断
	if contentType == "" {
		opt.ObjectPutHeaderOptions.ContentType = getContentTypeByExt(key)
	}

	// 上传文件到COS
	_, err := s.client.Object.Put(ctx, key, bytes.NewReader(data), opt)
	if err != nil {
		return "", fmt.Errorf("上传文件到COS失败: %v", err)
	}

	return s.GetURL(key), nil
}

// Download 下载文件
func (s *COSStorageService) Download(ctx context.Context, key string) ([]byte, error) {
	if s.client == nil {
		return nil, fmt.Errorf("COS客户端未初始化")
	}

	// 从COS获取文件
	resp, err := s.client.Object.Get(ctx, key, nil)
	if err != nil {
		return nil, fmt.Errorf("从COS下载文件失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取文件内容
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %v", err)
	}

	return data, nil
}

// Delete 删除文件
func (s *COSStorageService) Delete(ctx context.Context, key string) error {
	if s.client == nil {
		return fmt.Errorf("COS客户端未初始化")
	}

	// 从COS删除文件
	_, err := s.client.Object.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("从COS删除文件失败: %v", err)
	}

	return nil
}

// GetURL 获取文件URL
func (s *COSStorageService) GetURL(key string) string {
	if s.urlPrefix != "" {
		return s.urlPrefix + "/" + key
	}
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", s.bucket, s.region, key)
}

// SaveFile 保存文件（扩展方法）
func (s *COSStorageService) SaveFile(module, category, relatedID, fileID, fileName string, data io.Reader) (filePath, fileURL string, err error) {
	// 构建文件路径
	key := fmt.Sprintf("%s/%s/%s/%s_%s", module, category, relatedID, fileID, fileName)

	// 读取文件数据
	fileData, err := io.ReadAll(data)
	if err != nil {
		return "", "", fmt.Errorf("读取文件数据失败: %v", err)
	}

	// 上传到COS
	_, err = s.Upload(context.Background(), key, fileData, "")
	if err != nil {
		return "", "", err
	}

	return key, s.GetURL(key), nil
}

// DeleteFile 删除文件（扩展方法）
func (s *COSStorageService) DeleteFile(filePath string) error {
	return s.Delete(context.Background(), filePath)
}

// GetFile 获取文件内容（扩展方法）
func (s *COSStorageService) GetFile(filePath string) ([]byte, error) {
	return s.Download(context.Background(), filePath)
}

// getContentTypeByExt 根据文件扩展名获取Content-Type
func getContentTypeByExt(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".ppt", ".pptx":
		return "application/vnd.ms-powerpoint"
	case ".txt":
		return "text/plain"
	case ".md":
		return "text/markdown"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".zip":
		return "application/zip"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}
