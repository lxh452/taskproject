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
		return utils.Response.BusinessError("employee_already_left"), nil
	}

	// 5. 检查是否是创始人，禁止给创始人递交离职
	if l.isFounder(employee) {
		return utils.Response.BusinessError("founder_cannot_leave"), nil
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
		return utils.Response.BusinessError("permission_verify_failed"), nil
	}

	// 7. 根据离职类型执行不同逻辑
	switch leaveType {
	case "hr_initiated":
		return l.handleHRInitiatedLeave(employee, req, leaveDate)
	case "employee_initiated":
		return l.handleEmployeeInitiatedLeave(employee, req, leaveDate)
	default:
		return utils.Response.BusinessError("invalid_leave_type"), nil
	}
}

// determineLeaveType 判断离职类型
func (l *EmployeeLeaveLogic) determineLeaveType(currentUserID string, employee *user.Employee) (string, error) {
	// 如果是员工本人操作，则为主动离职
	if employee.UserId == currentUserID {
		return "employee_initiated", nil
	}

	// 如果是HR操作，则为HR协商离职
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
		HandoverStatus: 0, // 待接收人确认（HR协商离职需要员工确认）
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
	taskNodes := []string{}
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

	// 1. 使用审批人查找器查找合适的审批人
	approverFinder := utils.NewApproverFinder(l.svcCtx.EmployeeModel, l.svcCtx.DepartmentModel, l.svcCtx.CompanyModel)
	approverResult, err := approverFinder.FindApprover(l.ctx, employee.Id)
	if err != nil {
		logx.Errorf("查找审批人失败: %v", err)
		return utils.Response.InternalError("查找审批人失败"), err
	}

	// 如果找不到审批人，不允许提交离职申请
	if approverResult == nil {
		return utils.Response.BusinessError("no_approver_found"), nil
	}

	// 2. 创建离职审批记录
	approvalID := utils.Common.GenId("leave_approval")
	approval := &task.TaskHandover{
		HandoverId:     approvalID,
		TaskId:         "", // 离职审批不关联具体任务
		FromEmployeeId: employee.Id,
		ToEmployeeId:   "", // 待审批时指定交接人
		HandoverReason: sql.NullString{String: req.LeaveReason, Valid: true},
		HandoverNote:   sql.NullString{String: "员工主动离职申请，等待上级审批", Valid: true},
		HandoverStatus: 1, // 待上级审批（跳过接收人确认步骤）
		ApproverId:     sql.NullString{String: approverResult.ApproverID, Valid: true},
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}

	_, err = l.svcCtx.TaskHandoverModel.Insert(l.ctx, approval)
	if err != nil {
		logx.Errorf("创建离职审批记录失败: %v", err)
		return utils.Response.InternalError("创建离职审批记录失败"), err
	}

	// 3. 发送通知给审批人
	taskNodes := []string{}

	// 获取审批人邮箱
	approver, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, approverResult.ApproverID)
	if err == nil && approver.Email.Valid && approver.Email.String != "" && l.svcCtx.EmailService != nil {
		if err := l.svcCtx.EmailService.SendEmployeeLeaveEmail(l.ctx, approver.Email.String, employee.RealName, taskNodes); err != nil {
			logx.Errorf("发送离职邮件失败: %v", err)
		}
	}

	return utils.Response.Success(map[string]interface{}{
		"message":      "离职申请已提交，等待审批",
		"approvalId":   approvalID,
		"employeeName": employee.RealName,
		"leaveDate":    leaveDate.Format("2006-01-02"),
		"status":       "pending_approval",
		"approverId":   approverResult.ApproverID,
		"approverName": approverResult.ApproverName,
		"approverType": approverResult.ApproverType,
	}), nil
}

// isFounder 检查员工是否是创始人
func (l *EmployeeLeaveLogic) isFounder(employee *user.Employee) bool {
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
