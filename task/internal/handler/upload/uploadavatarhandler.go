package upload

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"task_Project/task/internal/logic/upload"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// 上传头像
func UploadAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UploadAvatarRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 获取上传的文件
		fileData, handler, err := r.FormFile("avatar")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer fileData.Close()

		l := upload.NewUploadAvatarLogic(r.Context(), svcCtx)
		resp, err := l.UploadAvatar(&req, handler, fileData)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

