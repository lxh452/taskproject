package checklist

import (
	"context"
	"errors"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteChecklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteChecklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteChecklistLogic {
	return &DeleteChecklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteChecklistLogic) DeleteChecklist(req *types.DeleteChecklistRequest) (resp interface{}, err error) {
	// 1. 从上下文获取当前员工ID
	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}

	// 2. 查询清单
	checklist, err := l.svcCtx.TaskChecklistModel.FindOne(l.ctx, req.ChecklistID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询清单失败: %v", err)
		return nil, errors.New("清单不存在")
	}
	if checklist.DeleteTime.Valid {
		return nil, errors.New("清单已被删除")
	}

	// 3. 验证权限：只有创建者可以删除清单
	if checklist.CreatorId != employeeId {
		return nil, errors.New("您没有权限删除此清单，只有清单创建者可以删除")
	}

	taskNodeId := checklist.TaskNodeId

	// 4. 软删除清单
	err = l.svcCtx.TaskChecklistModel.SoftDelete(l.ctx, req.ChecklistID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("删除清单失败: %v", err)
		return nil, errors.New("删除清单失败")
	}

	// 5. 更新任务节点的清单统计
	err = l.updateNodeChecklistCount(taskNodeId)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务节点清单统计失败: %v", err)
		// 不影响主流程，只记录日志
	}

	// 6. 更新任务节点进度
	err = l.updateNodeProgress(taskNodeId)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务节点进度失败: %v", err)
	}

	return map[string]interface{}{
		"message": "删除成功",
	}, nil
}

// updateNodeChecklistCount 更新任务节点的清单统计
func (l *DeleteChecklistLogic) updateNodeChecklistCount(taskNodeId string) error {
	total, completed, err := l.svcCtx.TaskChecklistModel.CountByTaskNodeId(l.ctx, taskNodeId)
	if err != nil {
		return err
	}

	return l.svcCtx.TaskNodeModel.UpdateChecklistCount(l.ctx, taskNodeId, total, completed)
}

// updateNodeProgress 根据清单完成情况更新任务节点进度
func (l *DeleteChecklistLogic) updateNodeProgress(taskNodeId string) error {
	total, completed, err := l.svcCtx.TaskChecklistModel.CountByTaskNodeId(l.ctx, taskNodeId)
	if err != nil {
		return err
	}

	var progress int
	if total > 0 {
		progress = int(completed * 100 / total)
	}

	// 更新任务节点进度
	err = l.svcCtx.TaskNodeModel.UpdateProgress(l.ctx, taskNodeId, progress)
	if err != nil {
		return err
	}

	// 更新任务整体进度
	return l.updateTaskProgress(taskNodeId)
}

// updateTaskProgress 根据所有任务节点进度更新任务整体进度
func (l *DeleteChecklistLogic) updateTaskProgress(taskNodeId string) error {
	// 获取任务节点信息
	taskNode, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, taskNodeId)
	if err != nil {
		return err
	}

	// 获取该任务的所有节点
	nodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, taskNode.TaskId)
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return nil
	}

	// 计算平均进度和完成节点数
	var totalProgress int64
	var completedCount int64
	for _, node := range nodes {
		totalProgress += int64(node.Progress)
		if node.Progress >= 100 {
			completedCount++
		}
	}
	avgProgress := int(totalProgress / int64(len(nodes)))

	// 更新任务进度
	err = l.svcCtx.TaskModel.UpdateProgress(l.ctx, taskNode.TaskId, avgProgress)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务进度失败: %v", err)
	}

	// 更新任务节点统计
	err = l.svcCtx.TaskModel.UpdateNodeCount(l.ctx, taskNode.TaskId, int64(len(nodes)), completedCount)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务节点统计失败: %v", err)
	}

	return nil
}
