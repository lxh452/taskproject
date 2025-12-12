// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tasknode

import (
	"context"
	"errors"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdatePrerequisiteNodesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePrerequisiteNodesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePrerequisiteNodesLogic {
	return &UpdatePrerequisiteNodesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 专门用于更新前置节点的方法
func (l *UpdatePrerequisiteNodesLogic) UpdatePrerequisiteNodes(req *types.UpdatePrerequisiteNodesRequest) (resp *types.BaseResponse, err error) {
	resp = new(types.BaseResponse)
	// 1. 参数验证
	if req.NodeID == "" {
		return utils.Response.BusinessError("任务节点ID不能为空"), nil
	}

	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}
	// 3. 获取任务节点信息
	taskNode, err := l.svcCtx.TaskNodeModel.FindOneSafe(l.ctx, req.NodeID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("任务节点不存在"), nil
		}
		return nil, err
	}

	// 4. 验证用户权限（节点负责人或任务负责人都可以更新前置节点）
	// 获取任务信息，检查是否是任务负责人
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, taskNode.TaskId)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("获取任务信息失败: %v", err)
		return utils.Response.InternalError("获取任务信息失败"), nil
	}

	isNodeLeader := taskNode.LeaderId == employeeId
	isTaskLeader := taskInfo.TaskCreator == employeeId

	if !isNodeLeader && !isTaskLeader {
		return utils.Response.BusinessError("no_root_update"), nil
	}

	node, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, req.NodeID)
	if err != nil {
		return utils.Response.InternalError("更新前置节点失败"), nil
	}
	// 5. 更新前置节点
	if err := l.svcCtx.TaskNodeModel.UpdateExNodeIds(l.ctx, req.NodeID, req.PrerequisiteNodes); err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新前置节点失败: %v", err)
		return utils.Response.InternalError("更新前置节点失败"), nil
	}
	// 如果创建好了流程证明该流程已经完善了，因此需要修正任务
	err = l.svcCtx.TaskModel.UpdateStatus(l.ctx, node.TaskId, 1)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务状态失败: %v", err)
		return utils.Response.InternalError("更新任务失败"), err
	}

	return utils.Response.Success(map[string]interface{}{
		"nodeId":            req.NodeID,
		"prerequisiteNodes": req.PrerequisiteNodes,
		"message":           "前置节点更新成功",
	}), nil
}
