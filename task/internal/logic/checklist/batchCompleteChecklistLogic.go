package checklist

import (
	"context"
	"errors"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchCompleteChecklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchCompleteChecklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchCompleteChecklistLogic {
	return &BatchCompleteChecklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchCompleteChecklistLogic) BatchCompleteChecklist(req *types.BatchCompleteChecklistRequest) (resp interface{}, err error) {
	// 1. 从上下文获取当前员工ID
	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}

	// 2. 验证参数
	if len(req.ChecklistIDs) == 0 {
		return nil, errors.New("请选择要操作的清单")
	}

	// 3. 验证权限并收集需要更新的节点ID
	nodeIds := make(map[string]bool)
	validChecklistIds := make([]string, 0)

	for _, checklistId := range req.ChecklistIDs {
		checklist, err := l.svcCtx.TaskChecklistModel.FindOne(l.ctx, checklistId)
		if err != nil {
			l.Logger.Errorf("查询清单失败: checklistId=%s, error=%v", checklistId, err)
			continue
		}
		if checklist.DeleteTime.Valid {
			continue
		}

		// 验证权限：只有创建者可以修改清单完成状态
		if checklist.CreatorId != employeeId {
			l.Logger.Infof("用户无权修改清单: userId=%s, checklistId=%s, creatorId=%s",
				employeeId, checklistId, checklist.CreatorId)
			continue
		}

		validChecklistIds = append(validChecklistIds, checklistId)
		nodeIds[checklist.TaskNodeId] = true
	}

	if len(validChecklistIds) == 0 {
		return nil, errors.New("没有可操作的清单项（您只能修改自己创建的清单）")
	}

	// 4. 批量更新清单状态
	err = l.svcCtx.TaskChecklistModel.BatchUpdateCompleteStatus(l.ctx, validChecklistIds, req.IsCompleted)
	if err != nil {
		l.Logger.Errorf("批量更新清单状态失败: %v", err)
		return nil, errors.New("批量更新清单状态失败")
	}

	// 5. 更新相关任务节点的清单统计和进度
	for nodeId := range nodeIds {
		err = l.updateNodeChecklistCount(nodeId)
		if err != nil {
			l.Logger.Errorf("更新任务节点清单统计失败: nodeId=%s, error=%v", nodeId, err)
		}

		err = l.updateNodeProgress(nodeId)
		if err != nil {
			l.Logger.Errorf("更新任务节点进度失败: nodeId=%s, error=%v", nodeId, err)
		}
	}

	// 6. 返回结果
	statusText := "已完成"
	if req.IsCompleted == 0 {
		statusText = "未完成"
	}

	return map[string]interface{}{
		"message":      "批量操作成功",
		"updatedCount": len(validChecklistIds),
		"newStatus":    statusText,
	}, nil
}

// updateNodeChecklistCount 更新任务节点的清单统计
func (l *BatchCompleteChecklistLogic) updateNodeChecklistCount(taskNodeId string) error {
	total, completed, err := l.svcCtx.TaskChecklistModel.CountByTaskNodeId(l.ctx, taskNodeId)
	if err != nil {
		return err
	}

	return l.svcCtx.TaskNodeModel.UpdateChecklistCount(l.ctx, taskNodeId, total, completed)
}

// updateNodeProgress 根据清单完成情况更新任务节点进度
func (l *BatchCompleteChecklistLogic) updateNodeProgress(taskNodeId string) error {
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
func (l *BatchCompleteChecklistLogic) updateTaskProgress(taskNodeId string) error {
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
		l.Logger.Errorf("更新任务进度失败: %v", err)
	}

	// 更新任务节点统计
	err = l.svcCtx.TaskModel.UpdateNodeCount(l.ctx, taskNode.TaskId, int64(len(nodes)), completedCount)
	if err != nil {
		l.Logger.Errorf("更新任务节点统计失败: %v", err)
	}

	return nil
}
