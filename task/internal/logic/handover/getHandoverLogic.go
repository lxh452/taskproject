package handover

import (
	"context"
	"errors"

	"task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetHandoverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHandoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHandoverLogic {
	return &GetHandoverLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHandoverLogic) GetHandover(req *types.GetHandoverRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.HandoverID == "" {
		return utils.Response.ValidationError("交接ID不能为空"), nil
	}

	// 2. 获取当前用户ID
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 3. 获取当前员工信息
	employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, currentUserID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询员工失败: %v", err)
		return utils.Response.ValidationError("用户未绑定员工信息"), nil
	}
	currentEmployeeID := employee.Id

	// 4. 获取交接信息
	handover, err := l.svcCtx.TaskHandoverModel.FindOne(l.ctx, req.HandoverID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.ValidationError("交接记录不存在"), nil
		}
		l.Logger.WithContext(l.ctx).Errorf("获取交接信息失败: %v", err)
		return nil, err
	}

	// 5. 获取任务信息（离职申请的taskId可能为空）
	var taskInfo *task.Task
	taskTitle := ""
	if handover.TaskId != "" {
		task, err := l.svcCtx.TaskModel.FindOne(l.ctx, handover.TaskId)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("获取任务信息失败: %v", err)
			return utils.Response.ValidationError("任务不存在"), nil
		}
		taskInfo = task
		taskTitle = task.TaskTitle
	} else {
		// 离职申请没有关联任务
		taskTitle = "离职审批"
	}

	// 6. 验证用户权限（发起人、接收人、审批人或任务创建者可以查看）
	hasPermission := false
	approverId := ""
	if handover.ApproverId.Valid {
		approverId = handover.ApproverId.String
	}

	if handover.FromEmployeeId == currentEmployeeID ||
		handover.ToEmployeeId == currentEmployeeID ||
		approverId == currentEmployeeID {
		hasPermission = true
	}

	// 如果有任务，检查是否是任务创建者
	if taskInfo != nil && taskInfo.TaskCreator == currentEmployeeID {
		hasPermission = true
	}

	// 检查是否是部门负责人
	if !hasPermission && employee.DepartmentId.Valid && employee.DepartmentId.String != "" {
		department, deptErr := l.svcCtx.DepartmentModel.FindOne(l.ctx, employee.DepartmentId.String)
		if deptErr == nil && department.ManagerId.Valid && department.ManagerId.String == currentEmployeeID {
			hasPermission = true
		}
	}

	if !hasPermission {
		return utils.Response.ValidationError("无权限查看此交接记录"), nil
	}

	// 7. 获取发起人和接收人信息
	fromEmployeeName := ""
	fromEmployeePositionId := ""
	if fromEmp, fromErr := l.svcCtx.EmployeeModel.FindOne(l.ctx, handover.FromEmployeeId); fromErr == nil {
		fromEmployeeName = fromEmp.RealName
		if fromEmp.PositionId.Valid {
			fromEmployeePositionId = fromEmp.PositionId.String
		}
	}

	toEmployeeName := ""
	if toEmp, toErr := l.svcCtx.EmployeeModel.FindOne(l.ctx, handover.ToEmployeeId); toErr == nil {
		toEmployeeName = toEmp.RealName
	}

	approverName := ""
	if approverId != "" {
		if approver, approverErr := l.svcCtx.EmployeeModel.FindOne(l.ctx, approverId); approverErr == nil {
			approverName = approver.RealName
		}
	}

	// 8. 处理可空字段
	handoverReason := ""
	if handover.HandoverReason.Valid {
		handoverReason = handover.HandoverReason.String
	}
	handoverNote := ""
	if handover.HandoverNote.Valid {
		handoverNote = handover.HandoverNote.String
	}

	// 9. 转换为响应格式
	converter := utils.NewConverter()
	handoverDetail := map[string]interface{}{
		"handoverId":             handover.HandoverId,
		"taskId":                 handover.TaskId,
		"taskTitle":              taskTitle,
		"fromEmployeeId":         handover.FromEmployeeId,
		"fromEmployeeName":       fromEmployeeName,
		"fromEmployeePositionId": fromEmployeePositionId,
		"toEmployeeId":           handover.ToEmployeeId,
		"toEmployeeName":         toEmployeeName,
		"approverId":             approverId,
		"approverName":           approverName,
		"handoverType":           handover.HandoverType,
		"handoverStatus":         handover.HandoverStatus,
		"handoverReason":         handoverReason,
		"handoverNote":           handoverNote,
		"createTime":             handover.CreateTime.Format("2006-01-02 15:04:05"),
		"updateTime":             handover.UpdateTime.Format("2006-01-02 15:04:05"),
	}

	// 如果有任务信息，添加到响应中
	if taskInfo != nil {
		handoverDetail["task"] = converter.ToTaskInfo(taskInfo)
	}

	if handover.ApproveTime.Valid {
		handoverDetail["approveTime"] = handover.ApproveTime.Time.Format("2006-01-02 15:04:05")
	}

	return utils.Response.Success(handoverDetail), nil
}
