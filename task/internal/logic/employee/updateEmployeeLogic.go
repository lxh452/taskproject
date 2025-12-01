// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package employee

import (
	"context"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateEmployeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新员工信息
func NewUpdateEmployeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateEmployeeLogic {
	return &UpdateEmployeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 这里为自己修改不是hr修改
func (l *UpdateEmployeeLogic) UpdateEmployee(req *types.UpdateEmployeeRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.ID) {
		return utils.Response.ValidationError("员工ID不能为空"), nil
	}

	// 检查员工是否存在
	employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, req.ID)
	if err != nil {
		logx.Errorf("查询员工失败: %v", err)
		return utils.Response.ErrorWithKey("employee_not_found"), nil
	}

	// 记录旧职位ID，用于判断是否需要同步权限
	oldPositionID := ""
	if employee.PositionId.Valid {
		oldPositionID = employee.PositionId.String
	}

	// 构建更新数据
	updateData := make(map[string]interface{})
	if !utils.Validator.IsEmpty(req.DepartmentID) {
		updateData["department_id"] = req.DepartmentID
	}
	positionChanged := false
	if !utils.Validator.IsEmpty(req.PositionID) {
		updateData["position_id"] = req.PositionID
		positionChanged = (req.PositionID != oldPositionID)
	}
	if !utils.Validator.IsEmpty(req.EmployeeID) {
		updateData["employee_id"] = req.EmployeeID
	}
	if !utils.Validator.IsEmpty(req.RealName) {
		updateData["real_name"] = req.RealName
	}
	if !utils.Validator.IsEmpty(req.WorkEmail) {
		updateData["email"] = req.WorkEmail
	}
	if !utils.Validator.IsEmpty(req.WorkPhone) {
		updateData["phone"] = req.WorkPhone
	}
	if !utils.Validator.IsEmpty(req.Skills) {
		updateData["skills"] = req.Skills
	}
	if !utils.Validator.IsEmpty(req.RoleTags) {
		updateData["role_tags"] = req.RoleTags
	}
	if !utils.Validator.IsEmpty(req.HireDate) {
		hireDate, err := time.Parse("2006-01-02", req.HireDate)
		if err != nil {
			logx.Errorf("入职日期格式错误: %v", err)
			return utils.Response.InternalError("入职日期格式错误"), err
		}
		updateData["hire_date"] = hireDate
	}
	if !utils.Validator.IsEmpty(req.LeaveDate) {
		leaveDate, err := time.Parse("2006-01-02", req.LeaveDate)
		if err != nil {
			logx.Errorf("离职日期格式错误: %v", err)
			return utils.Response.InternalError("离职日期格式错误"), err
		}
		updateData["leave_date"] = leaveDate
	}

	if len(updateData) == 0 {
		return utils.Response.ValidationError("没有需要更新的字段"), nil
	}

	// 使用选择性更新
	err = l.svcCtx.EmployeeModel.SelectiveUpdate(l.ctx, req.ID, updateData)
	if err != nil {
		logx.Errorf("更新员工信息失败: %v", err)
		return utils.Response.InternalError("更新员工信息失败"), err
	}

	// 如果职位改变，同步员工权限
	if positionChanged {
		newPositionID := req.PositionID
		permissionSyncService := svc.NewPermissionSyncService(l.svcCtx)
		if err := permissionSyncService.SyncEmployeePermissions(l.ctx, employee.UserId, req.ID, newPositionID); err != nil {
			logx.Errorf("同步员工权限失败: %v", err)
			// 权限同步失败不影响员工更新，只记录日志
		}
	}

	return utils.Response.Success("更新员工信息成功"), nil
}
