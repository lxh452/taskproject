package checklist

import (
	"context"
	"database/sql"
	"errors"
	"task_Project/model/task"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateChecklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateChecklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChecklistLogic {
	return &UpdateChecklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateChecklistLogic) UpdateChecklist(req *types.UpdateChecklistRequest) (resp *types.ChecklistInfo, err error) {
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

	// 3. 验证权限：只有创建者可以修改清单
	if checklist.CreatorId != employeeId {
		return nil, errors.New("您没有权限修改此清单，只有清单创建者可以修改")
	}

	// 4. 更新清单字段
	if req.Content != "" {
		checklist.Content = req.Content
	}
	if req.SortOrder != 0 {
		checklist.SortOrder = req.SortOrder
	}

	// 处理完成状态变更
	oldCompletedStatus := checklist.IsCompleted
	if req.IsCompleted == 1 && oldCompletedStatus == 0 {
		// 标记为完成
		checklist.IsCompleted = 1
		checklist.CompleteTime = sql.NullTime{Time: time.Now(), Valid: true}
	} else if req.IsCompleted == 0 && oldCompletedStatus == 1 {
		// 取消完成
		checklist.IsCompleted = 0
		checklist.CompleteTime = sql.NullTime{Valid: false}
	}

	// 5. 保存更新
	err = l.svcCtx.TaskChecklistModel.Update(l.ctx, checklist)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新清单失败: %v", err)
		return nil, errors.New("更新清单失败")
	}

	// 6. 如果完成状态发生变化，更新任务节点的清单统计
	if oldCompletedStatus != checklist.IsCompleted {
		err = l.updateNodeChecklistCount(checklist.TaskNodeId)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务节点清单统计失败: %v", err)
			// 不影响主流程，只记录日志
		}

		// 同时更新任务节点的进度
		err = l.updateNodeProgress(checklist.TaskNodeId)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务节点进度失败: %v", err)
		}
	}

	// 7. 查询更新后的清单
	updatedChecklist, err := l.svcCtx.TaskChecklistModel.FindOne(l.ctx, req.ChecklistID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询更新后的清单失败: %v", err)
		return nil, errors.New("清单更新成功但查询失败")
	}
	one, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, updatedChecklist.TaskNodeId)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询任务失败: %v", err)
		return nil, errors.New("任务不存在")
	}
	// 创建任务日志
	// 7. 创建任务日志
	taskLog := &task.TaskLog{
		LogId:      utils.NewCommon().GenerateIDWithPrefix("task_log"),
		TaskId:     one.TaskId,
		LogType:    2, // 创建类型
		LogContent: "更新任务" + req.Content,
		EmployeeId: employeeId,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务日志失败: %v", err)
	}

	// 8. 转换为响应
	resp = &types.ChecklistInfo{
		ID:          updatedChecklist.ChecklistId,
		TaskNodeID:  updatedChecklist.TaskNodeId,
		CreatorID:   updatedChecklist.CreatorId,
		Content:     updatedChecklist.Content,
		IsCompleted: updatedChecklist.IsCompleted,
		SortOrder:   updatedChecklist.SortOrder,
		CreateTime:  utils.Common.FormatTime(updatedChecklist.CreateTime),
		UpdateTime:  utils.Common.FormatTime(updatedChecklist.UpdateTime),
	}

	if updatedChecklist.CompleteTime.Valid {
		resp.CompleteTime = utils.Common.FormatTime(updatedChecklist.CompleteTime.Time)
	}

	return resp, nil
}

// updateNodeChecklistCount 更新任务节点的清单统计
func (l *UpdateChecklistLogic) updateNodeChecklistCount(taskNodeId string) error {
	total, completed, err := l.svcCtx.TaskChecklistModel.CountByTaskNodeId(l.ctx, taskNodeId)
	if err != nil {
		return err
	}

	return l.svcCtx.TaskNodeModel.UpdateChecklistCount(l.ctx, taskNodeId, total, completed)
}

// updateNodeProgress 根据清单完成情况更新任务节点进度
func (l *UpdateChecklistLogic) updateNodeProgress(taskNodeId string) error {
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
	if progress == 100 {
		err = l.svcCtx.TaskNodeModel.UpdateStatus(l.ctx, taskNodeId, 2)
	}

	// 更新任务整体进度
	return l.updateTaskProgress(taskNodeId)
}

// updateTaskProgress 根据所有任务节点进度更新任务整体进度
func (l *UpdateChecklistLogic) updateTaskProgress(taskNodeId string) error {
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
		totalProgress += node.Progress
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

	// 当所有节点都完成时（平均进度达到100%），更新任务状态为已完成
	if avgProgress == 100 {
		err = l.svcCtx.TaskModel.UpdateStatus(l.ctx, taskNode.TaskId, 2)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务状态失败: %v", err)
		}
	}

	// 更新任务节点统计
	err = l.svcCtx.TaskModel.UpdateNodeCount(l.ctx, taskNode.TaskId, int64(len(nodes)), completedCount)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务节点统计失败: %v", err)
	}

	return nil
}
