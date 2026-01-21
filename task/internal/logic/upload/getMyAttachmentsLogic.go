package upload

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyAttachmentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取我的附件列表
func NewGetMyAttachmentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyAttachmentsLogic {
	return &GetMyAttachmentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyAttachmentsLogic) GetMyAttachments(req *types.GetMyAttachmentsRequest) (resp *types.BaseResponse, err error) {
	// 获取当前用户ID
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 获取当前员工信息
	employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, userID)
	if err != nil || employee == nil {
		logx.Errorf("[GetMyAttachments] 获取员工信息失败: userID=%s, err=%v", userID, err)
		return utils.Response.BusinessError("您尚未加入任何公司"), nil
	}

	logx.Infof("[GetMyAttachments] 查询参数: userID=%s, employeeID=%s, page=%d, pageSize=%d, fileType=%s, module=%s",
		userID, employee.Id, req.Page, req.PageSize, req.FileType, req.Module)

	// 分页参数
	page := int64(req.Page)
	pageSize := int64(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 查询我上传的附件
	files, total, err := l.svcCtx.UploadFileModel.FindByUploaderID(
		l.ctx,
		employee.Id,
		page,
		pageSize,
		req.FileType,
		req.Module,
	)
	if err != nil {
		logx.Errorf("[GetMyAttachments] 查询我的附件失败: uploaderID=%s, err=%v", employee.Id, err)
		return utils.Response.InternalError("查询附件失败"), nil
	}

	logx.Infof("[GetMyAttachments] 查询结果: uploaderID=%s, total=%d, filesCount=%d", employee.Id, total, len(files))

	// 构建响应数据
	var list []types.MyAttachmentInfo
	for _, file := range files {
		item := types.MyAttachmentInfo{
			FileID:      file.FileID,
			FileURL:     file.FileURL,
			FileName:    file.FileName,
			FileSize:    file.FileSize,
			FileType:    file.FileType,
			Module:      file.Module,
			Category:    file.Category,
			RelatedID:   file.RelatedID,
			TaskNodeID:  file.TaskNodeID,
			Description: file.Description,
			Tags:        file.Tags,
			UploadTime:  file.CreateAt.Format("2006-01-02 15:04:05"),
		}

		// 获取关联任务/节点名称
		if file.RelatedID != "" {
			if file.Module == "task" {
				task, _ := l.svcCtx.TaskModel.FindOne(l.ctx, file.RelatedID)
				if task != nil {
					item.RelatedName = task.TaskTitle
				}
			} else if file.Module == "tasknode" || file.TaskNodeID != "" {
				nodeID := file.TaskNodeID
				if nodeID == "" {
					nodeID = file.RelatedID
				}
				node, _ := l.svcCtx.TaskNodeModel.FindOne(l.ctx, nodeID)
				if node != nil {
					item.RelatedName = node.NodeName
				}
			}
		}

		list = append(list, item)
	}

	return utils.Response.Success(map[string]interface{}{
		"list":  list,
		"total": total,
	}), nil
}
