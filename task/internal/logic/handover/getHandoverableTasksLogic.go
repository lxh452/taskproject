package handover

import (
	"context"
	"strings"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHandoverableTasksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHandoverableTasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHandoverableTasksLogic {
	return &GetHandoverableTasksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// HandoverableTask 可交接的任务
type HandoverableTask struct {
	TaskId      string `json:"taskId"`
	TaskTitle   string `json:"taskTitle"`
	TaskStatus  int64  `json:"taskStatus"`
	Deadline    string `json:"deadline"`
	Role        string `json:"role"` // creator/responsible/executor/leader
	RoleDisplay string `json:"roleDisplay"`
}

// GetHandoverableTasks 获取当前用户可以交接的任务列表
func (l *GetHandoverableTasksLogic) GetHandoverableTasks(req *types.PageReq) (resp *types.BaseResponse, err error) {
	// 1. 获取当前用户ID
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 2. 获取员工信息
	employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, currentUserID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("获取员工信息失败: %v", err)
		return utils.Response.ValidationError("用户未绑定员工信息"), nil
	}
	employeeID := employee.Id

	// 3. 收集可交接的任务（使用map去重）
	taskMap := make(map[string]*HandoverableTask)

	// 3.1 查找用户作为任务创建者的任务
	creatorTasks, _, err := l.svcCtx.TaskModel.FindByCreator(l.ctx, employeeID, 1, 100)
	if err == nil {
		for _, task := range creatorTasks {
			// 只保留进行中的任务（状态1）或已到开始日期的未开始任务（状态0）
			if l.isTaskHandoverable(task.TaskStatus, task.TaskStartTime) {
				if _, exists := taskMap[task.TaskId]; !exists {
					taskMap[task.TaskId] = &HandoverableTask{
						TaskId:      task.TaskId,
						TaskTitle:   task.TaskTitle,
						TaskStatus:  task.TaskStatus,
						Deadline:    task.TaskDeadline.Format("2006-01-02"),
						Role:        "creator",
						RoleDisplay: "创建者",
					}
				}
			}
		}
	}

	// 3.2 查找用户作为任务负责人的任务
	allTasks, _, err := l.svcCtx.TaskModel.FindByPage(l.ctx, 1, 500)
	if err == nil {
		for _, task := range allTasks {
			if !l.isTaskHandoverable(task.TaskStatus, task.TaskStartTime) {
				continue
			}
			// 检查是否是负责人
			if task.ResponsibleEmployeeIds.Valid && task.ResponsibleEmployeeIds.String != "" {
				responsibleIds := strings.Split(task.ResponsibleEmployeeIds.String, ",")
				for _, id := range responsibleIds {
					if strings.TrimSpace(id) == employeeID {
						if existing, exists := taskMap[task.TaskId]; exists {
							// 如果已存在，更新角色
							existing.Role = existing.Role + ",responsible"
							existing.RoleDisplay = existing.RoleDisplay + "/负责人"
						} else {
							taskMap[task.TaskId] = &HandoverableTask{
								TaskId:      task.TaskId,
								TaskTitle:   task.TaskTitle,
								TaskStatus:  task.TaskStatus,
								Deadline:    task.TaskDeadline.Format("2006-01-02"),
								Role:        "responsible",
								RoleDisplay: "负责人",
							}
						}
						break
					}
				}
			}
		}
	}

	// 3.3 查找用户作为任务节点执行人的任务
	executorNodes, _, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, employeeID, 1, 100)
	if err == nil {
		for _, node := range executorNodes {
			// 只保留进行中(1)或已到开始日期的未开始(0)节点
			if l.isNodeHandoverable(node.NodeStatus, node.NodeStartTime) {
				// 获取任务信息
				task, taskErr := l.svcCtx.TaskModel.FindOne(l.ctx, node.TaskId)
				if taskErr != nil {
					continue
				}
				if existing, exists := taskMap[node.TaskId]; exists {
					existing.Role = existing.Role + ",executor"
					existing.RoleDisplay = existing.RoleDisplay + "/执行人"
				} else {
					taskMap[node.TaskId] = &HandoverableTask{
						TaskId:      node.TaskId,
						TaskTitle:   task.TaskTitle,
						TaskStatus:  task.TaskStatus,
						Deadline:    node.NodeDeadline.Format("2006-01-02"),
						Role:        "executor",
						RoleDisplay: "执行人",
					}
				}
			}
		}
	}

	// 3.4 查找用户作为任务节点负责人的任务
	leaderNodes, _, err := l.svcCtx.TaskNodeModel.FindByLeader(l.ctx, employeeID, 1, 100)
	if err == nil {
		for _, node := range leaderNodes {
			// 只保留进行中(1)或已到开始日期的未开始(0)节点
			if l.isNodeHandoverable(node.NodeStatus, node.NodeStartTime) {
				// 获取任务信息
				task, taskErr := l.svcCtx.TaskModel.FindOne(l.ctx, node.TaskId)
				if taskErr != nil {
					continue
				}
				if existing, exists := taskMap[node.TaskId]; exists {
					existing.Role = existing.Role + ",leader"
					existing.RoleDisplay = existing.RoleDisplay + "/节点负责人"
				} else {
					taskMap[node.TaskId] = &HandoverableTask{
						TaskId:      node.TaskId,
						TaskTitle:   task.TaskTitle,
						TaskStatus:  task.TaskStatus,
						Deadline:    node.NodeDeadline.Format("2006-01-02"),
						Role:        "leader",
						RoleDisplay: "节点负责人",
					}
				}
			}
		}
	}

	// 4. 转换为列表
	tasks := make([]HandoverableTask, 0, len(taskMap))
	for _, task := range taskMap {
		tasks = append(tasks, *task)
	}

	return utils.Response.Success(map[string]interface{}{
		"list":  tasks,
		"total": len(tasks),
	}), nil
}

// isTaskHandoverable 判断任务是否可交接
// 任务状态: 0-未开始, 1-进行中, 2-已完成, 3-逾期完成
func (l *GetHandoverableTasksLogic) isTaskHandoverable(status int64, startTime time.Time) bool {
	if status == 1 {
		return true // 进行中
	}
	if status == 0 && !startTime.IsZero() && !time.Now().Before(startTime) {
		return true // 未开始但已到开始日期
	}
	return false
}

// isNodeHandoverable 判断节点是否可交接
// 节点状态: 0-未开始, 1-进行中, 2-已完成, 3-已逾期
func (l *GetHandoverableTasksLogic) isNodeHandoverable(status int64, startTime time.Time) bool {
	if status == 1 {
		return true // 进行中
	}
	if status == 0 && !startTime.IsZero() && !time.Now().Before(startTime) {
		return true // 未开始但已到开始日期
	}
	return false
}
