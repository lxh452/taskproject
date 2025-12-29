// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package upload

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/upload"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"
)

// 文件上传（支持图片、PDF、Markdown等文件）
func UploadInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UploadInfoRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		// 获取文件信息，进行处理
		fileData, handler, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer fileData.Close()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := upload.NewUploadInfoLogic(r.Context(), svcCtx)
		resp, err := l.UploadInfo(&req, handler, fileData)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, utils.Response.ValidationError(err.Error()))
		} else {
			// 返回标准格式
			httpx.OkJsonCtx(r.Context(), w, utils.Response.Success(map[string]interface{}{
				"fileId":    resp.FileID,
				"fileName":  resp.FileName,
				"fileUrl":   resp.FileURL,
				"fileType":  resp.FileType,
				"fileSize":  resp.FileSize,
				"relatedId": resp.RelatedID,
			}))
		}
	}
}
