package checklist

import (
	"context"
	"errors"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChecklistListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetChecklistListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChecklistListLogic {
	return &GetChecklistListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetChecklistListLogic) GetChecklistList(req *types.GetChecklistListRequest) (resp interface{}, err error) {
	// 1. 验证任务节点是否存在
	taskNode, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, req.TaskNodeID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询任务节点失败: %v", err)
		return nil, errors.New("任务节点不存在")
	}
	if taskNode.DeleteTime.Valid {
		return nil, errors.New("任务节点已被删除")
	}

	// 2. 设置分页默认值
	page := utils.Common.DefaultInt(req.Page, 1)
	pageSize := utils.Common.DefaultInt(req.PageSize, 20)

	// 3. 查询清单列表
	checklists, total, err := l.svcCtx.TaskChecklistModel.FindByTaskNodeIdWithPage(l.ctx, req.TaskNodeID, page, pageSize)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询清单列表失败: %v", err)
		return nil, errors.New("查询清单列表失败")
	}

	// 4. 获取清单统计信息
	totalCount, completedCount, err := l.svcCtx.TaskChecklistModel.CountByTaskNodeId(l.ctx, req.TaskNodeID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询清单统计失败: %v", err)
	}

	// 计算进度
	var progress int64
	if totalCount > 0 {
		progress = completedCount * 100 / totalCount
	}

	// 5. 转换为响应列表
	list := make([]types.ChecklistInfo, 0, len(checklists))
	for _, checklist := range checklists {
		// 获取创建者信息
		var creatorName string
		employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, checklist.CreatorId)
		if err == nil {
			creatorName = employee.RealName
		}

		item := types.ChecklistInfo{
			ID:          checklist.ChecklistId,
			TaskNodeID:  checklist.TaskNodeId,
			CreatorID:   checklist.CreatorId,
			CreatorName: creatorName,
			Content:     checklist.Content,
			IsCompleted: checklist.IsCompleted,
			SortOrder:   checklist.SortOrder,
			CreateTime:  utils.Common.FormatTime(checklist.CreateTime),
			UpdateTime:  utils.Common.FormatTime(checklist.UpdateTime),
		}

		if checklist.CompleteTime.Valid {
			item.CompleteTime = utils.Common.FormatTime(checklist.CompleteTime.Time)
		}

		list = append(list, item)
	}

	// 6. 返回结果
	return map[string]interface{}{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"stats": types.ChecklistStats{
			TaskNodeID:     req.TaskNodeID,
			TotalCount:     totalCount,
			CompletedCount: completedCount,
			Progress:       progress,
		},
	}, nil
}
