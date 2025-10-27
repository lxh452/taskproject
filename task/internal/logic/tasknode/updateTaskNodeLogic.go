// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tasknode

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTaskNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新任务节点
func NewUpdateTaskNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTaskNodeLogic {
	return &UpdateTaskNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTaskNodeLogic) UpdateTaskNode(req *types.UpdateTaskNodeRequest) (resp *types.BaseResponse, err error) {

	if req.PrerequisiteNodes != "" {
		//情况1： 前置任务条件的选择可以等后续节点人都完成该总任务的任务填写后，告知所有节点负责人自己的任务节点的前置条件
		//（当所有节点都完成编写后，发送邮件和消息给节点负责人和负责人 需要他们补充进展） //这里放在update中 主动/被动
		// 这个情况跟情况七一起处理

		// 情况七，需要先完成节点的修正和删除
	}
	if len(req.Progress) > 0 {
		// 情况二，执行人每天修正自己的任务节点情况，任务节点进度的修正 被动调用该接口
	}
	if len(req.ExecutorID) > 0 && len(req.LastExecutorID) > 0 {
		// 情况三，人员变动
	}

	if req.NodeDetail != "" {
		// 情况四，节点负责人修改任务细节
	}
	if len(req.ExecutorID) > 0 && len(req.LastExecutorID) == 0 {
		// 情况五，人员增派
	}
	// 情况六，节点状态修正,一般情况是得系统查询到上个所需节点完成会自动修正
	if len(req.NodeStatus) > 0 {

	}

	return
}
