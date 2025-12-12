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

// TaskNodeWithTitle 带任务标题的任务节点
type TaskNodeWithTitle struct {
	TaskNodeId     string `json:"taskNodeId"`
	TaskId         string `json:"taskId"`
	TaskTitle      string `json:"taskTitle"` // 任务标题
	DepartmentId   string `json:"departmentId"`
	NodeName       string `json:"nodeName"`
	NodeDetail     string `json:"nodeDetail"`
	NodeDeadline   string `json:"nodeDeadline"`
	NodeStartTime  string `json:"nodeStartTime"`
	EstimatedDays  int64  `json:"estimatedDays"`
	ActualDays     int64  `json:"actualDays"`
	NodeStatus     int64  `json:"nodeStatus"`
	NodeFinishTime string `json:"nodeFinishTime"`
	ExecutorId     string `json:"executorId"`
	LeaderId       string `json:"leaderId"`
	Progress       int64  `json:"progress"`
	NodePriority   int64  `json:"nodePriority"`
}

func (l *GetUserTaskNodeLogic) GetUserTaskNode(req *types.PageReq) (resp *types.BaseResponse, err error) {
	// 从JWT获取当前用户ID
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	page, pageSize, _ := utils.Validator.ValidatePageParams(req.Page, req.PageSize)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}
	// 将用户ID映射为员工ID（任务节点里存的是员工ID）
	emp, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, userID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("根据用户ID获取员工信息失败 userID=%s, err=%v", userID, err)
		// 返回空列表而不是报错，避免前端异常
		data := Data{ExecutorTask: []TaskNodeWithTitle{}, ExecutorTaskCount: 0, LeaderTask: []TaskNodeWithTitle{}, LeaderTaskCount: 0}
		return utils.Response.Success(data), nil
	}
	employeeID := emp.Id

	leaderTask, leaderTaskCount, err := l.svcCtx.TaskNodeModel.FindByLeader(l.ctx, employeeID, page, pageSize)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询用户的任务列表 err:%v", err)
		return nil, err
	}
	executorTask, executorTaskCount, err := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, employeeID, page, pageSize)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询用户的任务列表（执行者） err:%v", err)
		return nil, err
	}

	// 转换为带任务标题的结构
	executorTaskWithTitle := l.convertToTaskNodeWithTitle(executorTask)
	leaderTaskWithTitle := l.convertToTaskNodeWithTitle(leaderTask)

	data := Data{
		ExecutorTask:      executorTaskWithTitle,
		ExecutorTaskCount: executorTaskCount,
		LeaderTask:        leaderTaskWithTitle,
		LeaderTaskCount:   leaderTaskCount,
	}

	return utils.Response.Success(data), nil
}

// convertToTaskNodeWithTitle 将任务节点转换为带任务标题的结构
func (l *GetUserTaskNodeLogic) convertToTaskNodeWithTitle(nodes []*task.TaskNode) []TaskNodeWithTitle {
	result := make([]TaskNodeWithTitle, 0, len(nodes))

	// 缓存任务信息，避免重复查询
	taskCache := make(map[string]string)

	for _, node := range nodes {
		taskTitle := ""
		if title, ok := taskCache[node.TaskId]; ok {
			taskTitle = title
		} else {
			// 查询任务信息
			taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, node.TaskId)
			if err == nil {
				taskTitle = taskInfo.TaskTitle
				taskCache[node.TaskId] = taskTitle
			}
		}

		nodeDetail := ""
		if node.NodeDetail.Valid {
			nodeDetail = node.NodeDetail.String
		}

		actualDays := int64(0)
		if node.ActualDays.Valid {
			actualDays = node.ActualDays.Int64
		}

		nodeFinishTime := ""
		if node.NodeFinishTime.Valid {
			nodeFinishTime = node.NodeFinishTime.Time.Format("2006-01-02 15:04:05")
		}

		result = append(result, TaskNodeWithTitle{
			TaskNodeId:     node.TaskNodeId,
			TaskId:         node.TaskId,
			TaskTitle:      taskTitle,
			DepartmentId:   node.DepartmentId,
			NodeName:       node.NodeName,
			NodeDetail:     nodeDetail,
			NodeDeadline:   node.NodeDeadline.Format("2006-01-02 15:04:05"),
			NodeStartTime:  node.NodeStartTime.Format("2006-01-02 15:04:05"),
			EstimatedDays:  node.EstimatedDays,
			ActualDays:     actualDays,
			NodeStatus:     node.NodeStatus,
			NodeFinishTime: nodeFinishTime,
			ExecutorId:     node.ExecutorId,
			LeaderId:       node.LeaderId,
			Progress:       node.Progress,
			NodePriority:   node.NodePriority,
		})
	}

	return result
}

type Data struct {
	ExecutorTask      []TaskNodeWithTitle `json:"executor_task"`
	ExecutorTaskCount int64               `json:"executor_task_count"`
	LeaderTask        []TaskNodeWithTitle `json:"leader_task"`
	LeaderTaskCount   int64               `json:"leader_task_count"`
}
