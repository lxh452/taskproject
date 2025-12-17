package upload

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"task_Project/task/internal/svc"
)

type ProxyFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProxyFileLogic {
	return &ProxyFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ProxyFile 代理文件内容（解决CORS问题）
func (l *ProxyFileLogic) ProxyFile(w http.ResponseWriter, r *http.Request) {
	// 从查询参数获取fileId
	fileId := r.URL.Query().Get("fileId")
	if fileId == "" {
		http.Error(w, "文件ID不能为空", http.StatusBadRequest)
		return
	}

	// 从请求头获取token并验证
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		logx.Errorf("代理文件请求缺少Authorization header")
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}

	// 提取token（格式：Bearer <token>）
	var token string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		logx.Errorf("代理文件请求Authorization格式错误: %s", authHeader)
		http.Error(w, "无效的授权格式", http.StatusUnauthorized)
		return
	}

	// 解析token获取用户ID
	claims, err := l.svcCtx.JWTMiddleware.ParseToken(token)
	if err != nil {
		logx.Errorf("Token解析失败: %v", err)
		http.Error(w, "Token无效或已过期", http.StatusUnauthorized)
		return
	}

	// 验证token是否在Redis中有效
	if err := l.svcCtx.JWTMiddleware.ValidateTokenWithRedis(token, claims.UserID); err != nil {
		logx.Errorf("Token验证失败: %v", err)
		http.Error(w, "Token无效或已过期", http.StatusUnauthorized)
		return
	}

	currentUserID := claims.UserID
	logx.Infof("代理文件请求，用户ID: %s, 文件ID: %s", currentUserID, fileId)

	// 从MongoDB查询文件信息
	file, err := l.svcCtx.UploadFileModel.FindByFileID(l.ctx, fileId)
	if err != nil {
		logx.Errorf("查询文件详情失败: %v", err)
		http.Error(w, "文件不存在", http.StatusNotFound)
		return
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
			http.Error(w, "您没有权限查看此附件", http.StatusForbidden)
			return
		}
	}

	// 从COS获取文件内容
	// file.FilePath 是COS的Key
	// 检查FileStorageService是否是COSStorageService
	cosService, ok := l.svcCtx.FileStorageService.(*svc.COSStorageService)
	if !ok {
		http.Error(w, "存储服务类型错误", http.StatusInternalServerError)
		return
	}
	fileData, err := cosService.GetFile(file.FilePath)
	if err != nil {
		logx.Errorf("从COS获取文件失败: %v", err)
		http.Error(w, "获取文件失败", http.StatusInternalServerError)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", getContentType(file.FileName))
	w.Header().Set("Content-Disposition", `inline; filename="`+file.FileName+`"`)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fileData)))

	// 写入文件内容
	if _, err := w.Write(fileData); err != nil {
		logx.Errorf("写入响应失败: %v", err)
	}
}

// getContentType 根据文件名获取Content-Type
func getContentType(fileName string) string {
	idx := strings.LastIndex(fileName, ".")
	if idx == -1 {
		return "application/octet-stream"
	}
	ext := strings.ToLower(fileName[idx:])
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".txt":
		return "text/plain"
	case ".md":
		return "text/markdown"
	case ".html":
		return "text/html"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}
