// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package department

import (
	"context"
	"task_Project/model/company"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDepartmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新部门信息
func NewUpdateDepartmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDepartmentLogic {
	return &UpdateDepartmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDepartmentLogic) UpdateDepartment(req *types.UpdateDepartmentRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.ID) {
		return utils.Response.ValidationError("部门ID不能为空"), nil
	}

	// 检查部门是否存在
	if _, err = l.svcCtx.DepartmentModel.FindOne(l.ctx, req.ID); err != nil {
		logx.Errorf("查询部门失败: %v", err)
		return utils.Response.ErrorWithKey("department_not_found"), nil
	}

	// 更新部门信息
	updateData := l.updateData(req)

	if len(updateData) == 0 {
		return utils.Response.ValidationError("没有需要更新的字段"), nil
	}

	updateData["update_time"] = time.Now()
	var updateDepartment company.Department
	err = utils.Common.MapToStructWithMapstructure(updateData, &updateDepartment)
	if err != nil {
		logx.Errorf("转换结构体失败: %v", err)
		return utils.Response.InternalError("转换结构体失败"), err
	}
	updateDepartment.Id = req.ID
	err = l.svcCtx.DepartmentModel.Update(l.ctx, &updateDepartment)
	if err != nil {
		logx.Errorf("更新部门信息失败: %v", err)
		return utils.Response.InternalError("更新部门信息失败"), err
	}

	return utils.Response.Success("更新部门信息成功"), nil
}

func (l *UpdateDepartmentLogic) updateData(req *types.UpdateDepartmentRequest) map[string]interface{} {
	// 更新部门信息
	updateData := make(map[string]interface{})
	if !utils.Validator.IsEmpty(req.DepartmentName) {
		updateData["department_name"] = req.DepartmentName
	}
	if !utils.Validator.IsEmpty(req.ParentID) {
		updateData["parent_id"] = req.ParentID
	}
	if !utils.Validator.IsEmpty(req.DepartmentCode) {
		updateData["department_code"] = req.DepartmentCode
	}
	if !utils.Validator.IsEmpty(req.ManagerID) {
		updateData["manager_id"] = req.ManagerID
	}
	if !utils.Validator.IsEmpty(req.Description) {
		updateData["description"] = req.Description
	}
	return updateData
}
