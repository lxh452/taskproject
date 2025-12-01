// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tasknode

import (
	"context"
	"errors"
	"fmt"

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
	fmt.Println(req.PrerequisiteNodes)
	// 1. 参数验证
	if req.NodeID == "" {
		return utils.Response.BusinessError("任务节点ID不能为空"), nil
	}

	// 2. 获取当前用户ID（用于权限验证）
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 3. 获取任务节点信息
	taskNode, err := l.svcCtx.TaskNodeModel.FindOneSafe(l.ctx, req.NodeID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("任务节点不存在"), nil
		}
		return nil, err
	}
	currentEmp, err := l.svcCtx.EmployeeModel.FindOneByUserId(l.ctx, currentUserID)
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		l.Logger.Errorf("更新前置节点失败: %v", err)
		return utils.Response.InternalError("更新前置节点失败"), nil
	}
	// 4. 验证用户权限（只有负责人可以更新前置节点）
	if taskNode.LeaderId != currentEmp.Id {
		return utils.Response.BusinessError("无权限更新此任务节点的前置节点"), nil
	}

	// 5. 更新前置节点
	if err := l.svcCtx.TaskNodeModel.UpdateExNodeIds(l.ctx, req.NodeID, req.PrerequisiteNodes); err != nil {
		l.Logger.Errorf("更新前置节点失败: %v", err)
		return utils.Response.InternalError("更新前置节点失败"), nil
	}

	return utils.Response.Success(map[string]interface{}{
		"nodeId":            req.NodeID,
		"prerequisiteNodes": req.PrerequisiteNodes,
		"message":           "前置节点更新成功",
	}), nil
}
