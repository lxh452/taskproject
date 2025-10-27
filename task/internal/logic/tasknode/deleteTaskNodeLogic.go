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

type DeleteTaskNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除任务节点
func NewDeleteTaskNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTaskNodeLogic {
	return &DeleteTaskNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTaskNodeLogic) DeleteTaskNode(req *types.DeleteTaskNodeRequest) (resp *types.BaseResponse, err error) {
	// 查询用户是否合法
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), err
	}
	nodeInfo, err := l.svcCtx.TaskNodeModel.FindByTaskNodeIDLeader(l.ctx, req.TaskNodeID, userID)
	if err != nil {
		logx.Infof("查询任务节点失败: %v", err)
		return utils.Response.NotFoundError("该任务节点查询失败"), err
	}
	//删除任务节点需要告知上级 请示，所有这里不能做操作，只能做一个发送操作，由引擎操作该任务节点的删除

	return
}
