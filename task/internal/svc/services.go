package svc

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
}

// NewCOSStorageService 创建COS存储服务
func NewCOSStorageService(secretID, secretKey, bucket, region, urlPrefix string) (*COSStorageService, error) {
	if secretID == "" || secretKey == "" || bucket == "" || region == "" {
		return nil, fmt.Errorf("COS配置不完整")
	}
	return &COSStorageService{
		secretID:  secretID,
		secretKey: secretKey,
		bucket:    bucket,
		region:    region,
		urlPrefix: urlPrefix,
	}, nil
}

// Upload 上传文件
func (s *COSStorageService) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	// TODO: 实现COS上传
	return "", nil
}

// Download 下载文件
func (s *COSStorageService) Download(ctx context.Context, key string) ([]byte, error) {
	// TODO: 实现COS下载
	return nil, nil
}

// Delete 删除文件
func (s *COSStorageService) Delete(ctx context.Context, key string) error {
	// TODO: 实现COS删除
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
