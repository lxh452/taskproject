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

type GetTaskNodeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取任务节点列表
func NewGetTaskNodeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskNodeListLogic {
	return &GetTaskNodeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTaskNodeListLogic) GetTaskNodeList(req *types.TaskNodeListRequest) (resp *types.BaseResponse, err error) {
	resp = new(types.BaseResponse)
	// 这里要获取该用户创建的任务节点和需要执行的任务节点
	taskNodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, req.TaskID)
	if err != nil {
		l.Errorf("查找总任务的任务节点失败：%v", err)
		return nil, err
	}
	return utils.Response.Success(taskNodes), nil
}
