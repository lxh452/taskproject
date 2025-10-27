package employee

import (
	"context"
	"time"

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
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.EmployeeID) {
		return utils.Response.ValidationError("员工ID不能为空"), nil
	}
	if validator.IsEmpty(req.LeaveReason) {
		return utils.Response.ValidationError("离职原因不能为空"), nil
	}

	// 获取当前用户信息
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 查询员工信息
	employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, req.EmployeeID)
	if err != nil {
		logx.Errorf("查询员工失败: %v", err)
		return utils.Response.ErrorWithKey("employee_not_found"), nil
	}

	// 检查权限（只有员工本人或HR可以操作离职）
	if employee.UserId != currentUserID {
		// 检查是否为HR或管理员
		// 这里可以添加角色检查逻辑
		// 暂时允许所有用户操作

		// todo
	}

	// 检查员工状态
	if employee.Status == 0 { // 已离职
		return utils.Response.BusinessError("员工已离职"), nil
	}

	// 解析离职时间
	var leaveDate time.Time
	if req.LeaveDate != "" {
		leaveDate, err = time.Parse("2006-01-02", req.LeaveDate)
		if err != nil {
			return utils.Response.ValidationError("离职时间格式错误，请使用 YYYY-MM-DD 格式"), nil
		}
	} else {
		leaveDate = time.Now()
	}

	// 更新员工状态为离职
	updateData := map[string]interface{}{
		"status":      0, // 离职
		"leave_date":  leaveDate,
		"update_time": time.Now(),
	}

	err = l.svcCtx.EmployeeModel.Update(l.ctx, req.EmployeeID, updateData)
	if err != nil {
		logx.Errorf("更新员工离职状态失败: %v", err)
		return utils.Response.InternalError("更新员工离职状态失败"), nil
	}

	// 更新用户状态为离职
	userUpdateData := map[string]interface{}{
		"status":      0, // 离职
		"update_time": time.Now(),
	}

	err = l.svcCtx.UserModel.Update(l.ctx, employee.UserId, userUpdateData)
	if err != nil {
		logx.Errorf("更新用户离职状态失败: %v", err)
		// 不影响主流程，只记录错误
	}

	// 发送离职通知给相关管理人员
	notificationService := svc.NewNotificationService(l.svcCtx)

	//todo 这里需要区分人事还是自己离职，需要告知离职原因 这里要修改一下方法
	err = notificationService.SendEmployeeLeaveNotification(req.EmployeeID, req.LeaveReason)
	if err != nil {
		logx.Errorf("发送离职通知失败: %v", err)
		// 不影响主流程，只记录错误
	}

	//todo 离职的时候需要将手头的工作递交通知到上级  and   找到交接人，并将手头的任务交付到他手里

	return utils.Response.Success("员工离职处理成功"), nil
}
