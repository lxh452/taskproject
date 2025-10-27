// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tasknode

import (
	"context"
	"task_Project/model/task"
	"task_Project/task/internal/utils"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserTaskNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取用户的任务节点信息
func NewGetUserTaskNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTaskNodeLogic {
	return &GetUserTaskNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserTaskNodeLogic) GetUserTaskNode(req *types.PageReq) (resp *types.BaseResponse, err error) {
	// 判断用户的状态
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}
	leaderTask, leaderTaskCount, err := l.svcCtx.TaskNodeModel.FindByLeader(l.ctx, userID, req.Page, req.PageSize)
	if err != nil {
		l.Logger.Errorf("查询用户的任务列表 err:%v", err)
		return nil, err
	}
	executorTask, executorTaskCount, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, userID, req.Page, req.PageSize)

	data := Data{
		ExecutorTask:      leaderTask,
		ExecutorTaskCount: executorTaskCount,
		LeaderTask:        executorTask,
		LeaderTaskCount:   leaderTaskCount,
	}

	return utils.Response.Success(data), nil
}

type Data struct {
	ExecutorTask      []*task.TaskNode `json:"executor_task"`
	ExecutorTaskCount int64            `json:"executor_task_count"`
	LeaderTask        []*task.TaskNode `json:"leader_task"`
	LeaderTaskCount   int64            `json:"leader_task_count"`
}
