package employee

import (
	"context"
	"database/sql"
	"time"

	"task_Project/model/task"
	"task_Project/model/user"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type EmployeeLeaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 员工离职
func NewEmployeeLeaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EmployeeLeaveLogic {
	return &EmployeeLeaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EmployeeLeaveLogic) EmployeeLeave(req *types.EmployeeLeaveRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.EmployeeID) {
		return utils.Response.ValidationError("员工ID不能为空"), nil
	}
	if validator.IsEmpty(req.LeaveReason) {
		return utils.Response.ValidationError("离职原因不能为空"), nil
	}

	// 2. 获取当前操作用户信息
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 3. 查询员工信息
	employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, req.EmployeeID)
	if err != nil {
		logx.Errorf("查询员工失败: %v", err)
		return utils.Response.ErrorWithKey("employee_not_found"), nil
	}

	// 4. 检查员工状态
	if employee.Status == 0 { // 已离职
		return utils.Response.BusinessError("员工已离职"), nil
	}

	// 5. 解析离职时间
	var leaveDate time.Time
	if req.LeaveDate != "" {
		leaveDate, err = time.Parse("2006-01-02", req.LeaveDate)
		if err != nil {
			return utils.Response.ValidationError("离职时间格式错误，请使用 YYYY-MM-DD 格式"), nil
		}
	} else {
		leaveDate = time.Now()
	}

	// 6. 判断离职类型和权限
	leaveType, err := l.determineLeaveType(currentUserID, employee)
	if err != nil {
		logx.Errorf("判断离职类型失败: %v", err)
		return utils.Response.BusinessError("权限验证失败"), nil
	}

	// 7. 根据离职类型执行不同逻辑
	switch leaveType {
	case "hr_initiated":
		return l.handleHRInitiatedLeave(employee, req, leaveDate)
	case "employee_initiated":
		return l.handleEmployeeInitiatedLeave(employee, req, leaveDate)
	default:
		return utils.Response.BusinessError("无效的离职类型"), nil
	}
}

// determineLeaveType 判断离职类型
func (l *EmployeeLeaveLogic) determineLeaveType(currentUserID string, employee *user.Employee) (string, error) {
	// 如果是员工本人操作，则为主动离职
	if employee.UserId == currentUserID {
		return "employee_initiated", nil
	}

	// 如果是HR操作，则为HR协商离职
	// TODO: 这里需要检查当前用户是否为HR角色
	// 暂时假设非员工本人操作的都是HR操作
	return "hr_initiated", nil
}

// handleHRInitiatedLeave 处理HR协商离职
func (l *EmployeeLeaveLogic) handleHRInitiatedLeave(employee *user.Employee, req *types.EmployeeLeaveRequest, leaveDate time.Time) (*types.BaseResponse, error) {
	logx.Infof("HR为员工 %s 办理离职", employee.RealName)

	// 1. 创建离职审批记录
	approvalID := utils.Common.GenId("leave_approval")
	approval := &task.TaskHandover{
		HandoverId:     approvalID,
		TaskId:         "", // 离职审批不关联具体任务
		FromEmployeeId: employee.Id,
		ToEmployeeId:   "", // 待审批
		HandoverReason: sql.NullString{String: req.LeaveReason, Valid: true},
		HandoverNote:   sql.NullString{String: "HR协商离职，等待员工确认", Valid: true},
		HandoverStatus: 1, // 待处理
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}

	_, err := l.svcCtx.TaskHandoverModel.Insert(l.ctx, approval)
	if err != nil {
		logx.Errorf("创建离职审批记录失败: %v", err)
		return utils.Response.InternalError("创建离职审批记录失败"), err
	}

	// 2. 发送通知给员工（通过消息队列）
	// 获取员工当前负责的任务节点
	taskNodes := []string{} // TODO: 实现获取任务节点逻辑
	recipientEmail := ""

	// 优先发给部门负责人；若无部门或无负责人邮箱，则发给员工本人
	if employee.DepartmentId.Valid && employee.DepartmentId.String != "" {
		department, err := l.svcCtx.DepartmentModel.FindOne(l.ctx, employee.DepartmentId.String)
		if err == nil && department.ManagerId.Valid && department.ManagerId.String != "" {
			manager, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, department.ManagerId.String)
			if err == nil && manager.Email.Valid && manager.Email.String != "" {
				recipientEmail = manager.Email.String
			}
		}
	}

	if recipientEmail == "" && employee.Email.Valid && employee.Email.String != "" {
		recipientEmail = employee.Email.String
	}

	if recipientEmail != "" && l.svcCtx.EmailService != nil {
		if err := l.svcCtx.EmailService.SendEmployeeLeaveEmail(l.ctx, recipientEmail, employee.RealName, taskNodes); err != nil {
			logx.Errorf("发送离职邮件失败: %v", err)
		}
	}

	return utils.Response.Success(map[string]interface{}{
		"message":      "HR协商离职通知已发送",
		"approvalId":   approvalID,
		"employeeName": employee.RealName,
		"leaveDate":    leaveDate.Format("2006-01-02"),
		"status":       "pending_employee_confirmation",
	}), nil
}

// handleEmployeeInitiatedLeave 处理员工主动离职
func (l *EmployeeLeaveLogic) handleEmployeeInitiatedLeave(employee *user.Employee, req *types.EmployeeLeaveRequest, leaveDate time.Time) (*types.BaseResponse, error) {
	logx.Infof("员工 %s 主动申请离职", employee.RealName)

	// 1. 创建离职审批记录
	approvalID := utils.Common.GenId("leave_approval")
	approval := &task.TaskHandover{
		HandoverId:     approvalID,
		TaskId:         "", // 离职审批不关联具体任务
		FromEmployeeId: employee.Id,
		ToEmployeeId:   "", // 待审批
		HandoverReason: sql.NullString{String: req.LeaveReason, Valid: true},
		HandoverNote:   sql.NullString{String: "员工主动离职申请，等待HR和部门负责人审批", Valid: true},
		HandoverStatus: 1, // 待处理
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}

	_, err := l.svcCtx.TaskHandoverModel.Insert(l.ctx, approval)
	if err != nil {
		logx.Errorf("创建离职审批记录失败: %v", err)
		return utils.Response.InternalError("创建离职审批记录失败"), err
	}

	// 2. 发送通知给HR和部门负责人（通过消息队列）
	taskNodes := []string{} // TODO: 实现获取任务节点逻辑
	recipientEmail := ""

	// 优先发给部门负责人
	if employee.DepartmentId.Valid && employee.DepartmentId.String != "" {
		department, err := l.svcCtx.DepartmentModel.FindOne(l.ctx, employee.DepartmentId.String)
		if err == nil && department.ManagerId.Valid && department.ManagerId.String != "" {
			manager, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, department.ManagerId.String)
			if err == nil && manager.Email.Valid && manager.Email.String != "" {
				recipientEmail = manager.Email.String
			}
		}
	}

	if recipientEmail == "" && employee.Email.Valid && employee.Email.String != "" {
		recipientEmail = employee.Email.String
	}

	if recipientEmail != "" && l.svcCtx.EmailService != nil {
		if err := l.svcCtx.EmailService.SendEmployeeLeaveEmail(l.ctx, recipientEmail, employee.RealName, taskNodes); err != nil {
			logx.Errorf("发送离职邮件失败: %v", err)
		}
	}

	return utils.Response.Success(map[string]interface{}{
		"message":      "离职申请已提交，等待审批",
		"approvalId":   approvalID,
		"employeeName": employee.RealName,
		"leaveDate":    leaveDate.Format("2006-01-02"),
		"status":       "pending_approval",
	}), nil
}
