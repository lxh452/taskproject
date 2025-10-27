// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tasknode

import (
	"context"
	"task_Project/task/internal/utils"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTaskNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取任务节点信息
func NewGetTaskNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskNodeLogic {
	return &GetTaskNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTaskNodeLogic) GetTaskNode(req *types.GetTaskNodeRequest) (resp *types.BaseResponse, err error) {
	detail, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, req.TaskNodeID)
	if detail == nil || err != nil {
		l.Logger.Errorf("获取任务节点详情失败")
		return utils.Response.NotFoundError("任务节点详情查找失败"), err
	}

	return utils.Response.Success(detail), nil
}
