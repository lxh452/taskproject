// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package employee

import (
	"context"
	"errors"
	"task_Project/model/user"
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
	if _, err = l.svcCtx.EmployeeModel.FindOne(l.ctx, req.ID); err != nil {
		logx.Errorf("查询员工失败: %v", err)
		return utils.Response.ErrorWithKey("employee_not_found"), nil
	}

	updateData, err := l.updateData(req)
	if err != nil {
		logx.Errorf("格式错误: %v", err)
		return utils.Response.InternalError("格式错误"), err
	}

	if len(updateData) == 0 {
		return utils.Response.ValidationError("没有需要更新的字段"), nil
	}

	updateData["update_time"] = time.Now()
	var employee user.Employee
	err = utils.Common.MapToStructWithMapstructure(updateData, &employee)
	if err != nil {
		logx.Errorf("格式错误: %v", err)
		return utils.Response.InternalError("格式错误"), err
	}
	employee.Id = req.ID
	err = l.svcCtx.EmployeeModel.Update(l.ctx, &employee)
	if err != nil {
		logx.Errorf("更新员工信息失败: %v", err)
		return utils.Response.InternalError("更新员工信息失败"), err
	}

	return utils.Response.Success("更新员工信息成功"), nil
}

func (l *UpdateEmployeeLogic) updateData(req *types.UpdateEmployeeRequest) (map[string]interface{}, error) {
	// 更新员工信息
	updateData := make(map[string]interface{})
	if !utils.Validator.IsEmpty(req.DepartmentID) {
		updateData["department_id"] = req.DepartmentID
	}
	if !utils.Validator.IsEmpty(req.PositionID) {
		updateData["position_id"] = req.PositionID
	}
	if !utils.Validator.IsEmpty(req.EmployeeID) {
		updateData["employee_id"] = req.EmployeeID
	}
	if !utils.Validator.IsEmpty(req.RealName) {
		updateData["real_name"] = req.RealName
	}
	if !utils.Validator.IsEmpty(req.WorkEmail) {
		updateData["work_email"] = req.WorkEmail
	}
	if !utils.Validator.IsEmpty(req.WorkPhone) {
		updateData["work_phone"] = req.WorkPhone
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
			return nil, errors.New("入职日期格式错误")
		}
		updateData["hire_date"] = hireDate
	}
	if !utils.Validator.IsEmpty(req.LeaveDate) {
		leaveDate, err := time.Parse("2006-01-02", req.LeaveDate)
		if err != nil {
			return nil, errors.New("离职日期格式错误")
		}
		updateData["leave_date"] = leaveDate
	}
	return updateData, nil
}
