package checklist

import (
	"context"
	"errors"
	"strings"
	"time"

	"task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChecklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateChecklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChecklistLogic {
	return &CreateChecklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateChecklistLogic) CreateChecklist(req *types.CreateChecklistRequest) (resp *types.ChecklistInfo, err error) {
	// 1. 从上下文获取当前员工ID
	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}

	// 2. 验证任务节点是否存在
	taskNode, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, req.TaskNodeID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询任务节点失败: %v", err)
		return nil, errors.New("任务节点不存在")
	}
	//校验一下这个项目是否已经启动了
	nodeTask, _ := l.svcCtx.TaskModel.FindOne(l.ctx, taskNode.TaskId)
	if nodeTask.TaskStatus <= 0 {
		l.Logger.WithContext(l.ctx).Errorf("查询任务失败: %v", err)
		return nil, errors.New("任务未开始")
	}
	if taskNode.DeleteTime.Valid {
		return nil, errors.New("任务节点已被删除")
	}

	// 3. 验证权限：只有任务节点的执行人才能创建清单
	if !l.isExecutor(taskNode.ExecutorId, employeeId) {
		return nil, errors.New("只有该任务节点的执行人才能创建清单")
	}

	// 4. 生成清单ID
	checklistId := utils.Common.GenId("checklist")

	// 5. 创建清单记录
	checklist := &task.TaskChecklist{
		ChecklistId: checklistId,
		TaskNodeId:  req.TaskNodeID,
		CreatorId:   employeeId,
		Content:     req.Content,
		IsCompleted: 0,
		SortOrder:   req.SortOrder,
	}

	_, err = l.svcCtx.TaskChecklistModel.Insert(l.ctx, checklist)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务清单失败: %v", err)
		return nil, errors.New("创建任务清单失败")
	}

	// 6. 更新任务节点的清单统计
	err = l.updateNodeChecklistCount(req.TaskNodeID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务节点清单统计失败: %v", err)
		// 不影响主流程，只记录日志
	}

	// 7. 查询新创建的清单
	newChecklist, err := l.svcCtx.TaskChecklistModel.FindOne(l.ctx, checklistId)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询新创建的清单失败: %v", err)
		return nil, errors.New("清单创建成功但查询失败")
	}

	// 7. 创建任务日志
	taskLog := &task.TaskLog{
		LogId:      utils.NewCommon().GenerateIDWithPrefix("task_log"),
		TaskId:     nodeTask.TaskId,
		LogType:    1, // 创建类型
		LogContent: "创建任务清单: " + req.Content,
		EmployeeId: employeeId,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务日志失败: %v", err)
	}

	// 8. 转换为响应
	resp = &types.ChecklistInfo{
		ID:          newChecklist.ChecklistId,
		TaskNodeID:  newChecklist.TaskNodeId,
		CreatorID:   newChecklist.CreatorId,
		Content:     newChecklist.Content,
		IsCompleted: newChecklist.IsCompleted,
		SortOrder:   newChecklist.SortOrder,
		CreateTime:  utils.Common.FormatTime(newChecklist.CreateTime),
		UpdateTime:  utils.Common.FormatTime(newChecklist.UpdateTime),
	}

	return resp, nil
}

// updateNodeChecklistCount 更新任务节点的清单统计
func (l *CreateChecklistLogic) updateNodeChecklistCount(taskNodeId string) error {
	total, completed, err := l.svcCtx.TaskChecklistModel.CountByTaskNodeId(l.ctx, taskNodeId)
	if err != nil {
		return err
	}
	_ = l.svcCtx.TaskNodeModel.UpdateStatus(l.ctx, taskNodeId, 1)
	return l.svcCtx.TaskNodeModel.UpdateChecklistCount(l.ctx, taskNodeId, total, completed)
}

// isExecutor 检查员工是否是任务节点的执行人
// executorId 可能是逗号分隔的多个ID
func (l *CreateChecklistLogic) isExecutor(executorId, employeeId string) bool {
	if executorId == "" || employeeId == "" {
		return false
	}
	// 检查是否在执行人列表中
	executors := strings.Split(executorId, ",")
	for _, id := range executors {
		if strings.TrimSpace(id) == employeeId {
			return true
		}
	}
	return false
}
