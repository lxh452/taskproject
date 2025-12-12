package checklist

import (
	"context"
	"errors"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChecklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetChecklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChecklistLogic {
	return &GetChecklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetChecklistLogic) GetChecklist(req *types.GetChecklistRequest) (resp *types.ChecklistInfo, err error) {
	// 1. 查询清单
	checklist, err := l.svcCtx.TaskChecklistModel.FindOne(l.ctx, req.ChecklistID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询清单失败: %v", err)
		return nil, errors.New("清单不存在")
	}
	if checklist.DeleteTime.Valid {
		return nil, errors.New("清单已被删除")
	}

	// 2. 获取创建者信息
	var creatorName string
	employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, checklist.CreatorId)
	if err == nil {
		creatorName = employee.RealName
	}

	// 3. 转换为响应
	resp = &types.ChecklistInfo{
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
		resp.CompleteTime = utils.Common.FormatTime(checklist.CompleteTime.Time)
	}

	return resp, nil
}
