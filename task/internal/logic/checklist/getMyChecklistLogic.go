package checklist

import (
	"context"
	"errors"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyChecklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyChecklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyChecklistLogic {
	return &GetMyChecklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyChecklistLogic) GetMyChecklist(req *types.GetMyChecklistRequest) (resp interface{}, err error) {
	// 1. 从上下文获取当前员工ID
	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}

	// 2. 设置分页参数
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// 3. 查询我的清单（根据任务节点执行人查询）
	checklists, total, err := l.svcCtx.TaskChecklistModel.FindByExecutorIdWithPage(l.ctx, employeeId, req.IsCompleted, page, pageSize)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询我的清单失败: %v", err)
		return nil, errors.New("查询清单失败")
	}

	// 4. 查询未完成数量（用于显示badge）
	uncompletedCount, err := l.svcCtx.TaskChecklistModel.CountUncompletedByExecutorId(l.ctx, employeeId)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("统计未完成清单数量失败: %v", err)
		uncompletedCount = 0
	}

	// 5. 构建响应列表，补充任务和节点信息
	list := make([]types.MyChecklistItem, 0, len(checklists))
	for _, c := range checklists {
		item := types.MyChecklistItem{
			ID:          c.ChecklistId,
			TaskNodeID:  c.TaskNodeId,
			Content:     c.Content,
			IsCompleted: int64(c.IsCompleted),
			SortOrder:   int64(c.SortOrder),
			CreateTime:  utils.Common.FormatTime(c.CreateTime),
			UpdateTime:  utils.Common.FormatTime(c.UpdateTime),
		}

		if c.CompleteTime.Valid {
			item.CompleteTime = utils.Common.FormatTime(c.CompleteTime.Time)
		}

		// 查询任务节点信息
		taskNode, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, c.TaskNodeId)
		if err == nil && taskNode != nil {
			item.TaskNodeName = taskNode.NodeName
			item.TaskID = taskNode.TaskId

			// 查询任务信息
			task, err := l.svcCtx.TaskModel.FindOne(l.ctx, taskNode.TaskId)
			if err == nil && task != nil {
				item.TaskTitle = task.TaskTitle
			}
		}

		list = append(list, item)
	}

	return map[string]interface{}{
		"list":             list,
		"total":            total,
		"page":             page,
		"pageSize":         pageSize,
		"uncompletedCount": uncompletedCount,
	}, nil
}
