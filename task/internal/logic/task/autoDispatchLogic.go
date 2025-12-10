// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"

	"task_Project/model/task"
	"task_Project/model/user"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type AutoDispatchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 任务自动派发
func NewAutoDispatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AutoDispatchLogic {
	return &AutoDispatchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 根据路径请求确认自动派发
func (l *AutoDispatchLogic) AutoDispatch(req *types.AutoDispatchRequest) (resp *types.BaseResponse, err error) {
	// 权限检查
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 获取任务信息
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, req.TaskID)
	if err != nil {
		l.Logger.Errorf("获取任务信息失败: %v", err)
		return utils.Response.BusinessError("任务不存在"), nil
	}

	// 检查用户权限（只有任务负责人或管理员可以触发自动派发）
	if !l.hasDispatchPermission(userID, taskInfo) {
		return utils.Response.BusinessError("无权限执行自动派发"), nil
	}

	// 获取任务的所有节点
	taskNodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, req.TaskID)
	if err != nil {
		l.Logger.Errorf("获取任务节点失败: %v", err)
		return utils.Response.BusinessError("获取任务节点失败"), nil
	}

	// 创建派发服务
	dispatchService := svc.NewDispatchService(l.svcCtx)

	// 统计派发结果
	var dispatchResults []DispatchResult
	successCount := 0
	failCount := 0

	// 遍历所有任务节点进行自动派发
	for _, taskNode := range taskNodes {
		// 跳过已经有执行人的节点
		if taskNode.ExecutorId != "" {
			continue
		}

		// 执行自动派发
		err := dispatchService.AutoDispatchTask(l.ctx, taskNode.TaskNodeId)
		if err != nil {
			l.Logger.Errorf("自动派发任务节点 %s 失败: %v", taskNode.TaskNodeId, err)
			dispatchResults = append(dispatchResults, DispatchResult{
				TaskNodeID: taskNode.TaskNodeId,
				NodeName:   taskNode.NodeName,
				Success:    false,
				Reason:     err.Error(),
			})
			failCount++
		} else {
			dispatchResults = append(dispatchResults, DispatchResult{
				TaskNodeID: taskNode.TaskNodeId,
				NodeName:   taskNode.NodeName,
				Success:    true,
				Reason:     "派发成功",
			})
			successCount++
		}
	}

	// 构建响应数据
	result := AutoDispatchResult{
		TaskID:          req.TaskID,
		TaskTitle:       taskInfo.TaskTitle,
		TotalNodes:      len(taskNodes),
		SuccessCount:    successCount,
		FailCount:       failCount,
		DispatchResults: dispatchResults,
	}

	return utils.Response.Success(result), nil
}

// DispatchResult 派发结果
type DispatchResult struct {
	TaskNodeID string `json:"taskNodeId"`
	NodeName   string `json:"nodeName"`
	Success    bool   `json:"success"`
	Reason     string `json:"reason"`
}

// AutoDispatchResult 自动派发结果
type AutoDispatchResult struct {
	TaskID          string           `json:"taskId"`
	TaskTitle       string           `json:"taskTitle"`
	TotalNodes      int              `json:"totalNodes"`
	SuccessCount    int              `json:"successCount"`
	FailCount       int              `json:"failCount"`
	DispatchResults []DispatchResult `json:"dispatchResults"`
}

// hasDispatchPermission 检查用户是否有派发权限
func (l *AutoDispatchLogic) hasDispatchPermission(userID string, taskInfo *task.Task) bool {
	// 创建者可派发
	if taskInfo.TaskCreator == userID {
		return true
	}

	// 获取员工信息
	employee, err := l.svcCtx.EmployeeModel.FindOneByUserId(l.ctx, userID)
	if err != nil {
		return false
	}

	// 检查是否是创始人（创始人可以给所有人派发任务，不受部门限制）
	if l.isFounder(employee) {
		return true
	}

	// 部门经理可派发（仅限本部门）
	if taskInfo.DepartmentIds.Valid && taskInfo.DepartmentIds.String != "" && employee.DepartmentId.Valid {
		deptID := taskInfo.DepartmentIds.String
		// 仅当员工属于该部门再校验
		if deptID == employee.DepartmentId.String {
			dept, err := l.svcCtx.DepartmentModel.FindOne(l.ctx, deptID)
			if err == nil && dept.ManagerId.Valid && dept.ManagerId.String == employee.EmployeeId {
				return true
			}
		}
	}
	return false
}

// isFounder 检查员工是否是创始人
func (l *AutoDispatchLogic) isFounder(employee *user.Employee) bool {
	// 检查职位代码是否为 FOUNDER
	if employee.PositionId.Valid {
		pos, err := l.svcCtx.PositionModel.FindOne(l.ctx, employee.PositionId.String)
		if err == nil && pos != nil && pos.PositionCode.Valid {
			if pos.PositionCode.String == "FOUNDER" {
				return true
			}
		}
	}
	// 检查是否是公司Owner
	company, err := l.svcCtx.CompanyModel.FindOne(l.ctx, employee.CompanyId)
	if err == nil && company != nil && company.Owner == employee.UserId {
		return true
	}
	return false
}
