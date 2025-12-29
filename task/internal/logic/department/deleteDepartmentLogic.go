// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package department

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDepartmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除部门
func NewDeleteDepartmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDepartmentLogic {
	return &DeleteDepartmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteDepartmentLogic) DeleteDepartment(req *types.DeleteDepartmentRequest) (resp *types.BaseResponse, err error) {
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

	// 检查部门是否有员工
	employeeCount, err := l.svcCtx.EmployeeModel.GetEmployeeCountByDepartment(l.ctx, req.ID)
	if err != nil {
		logx.Errorf("查询部门员工数量失败: %v", err)
		return utils.Response.InternalError("查询部门员工数量失败"), err
	}

	if employeeCount > 0 {
		return utils.Response.BusinessError("The department still has employees and cannot be deleted."), nil
	}

	// 检查是否有子部门
	subDepartmentCount, err := l.svcCtx.DepartmentModel.GetSubDepartments(l.ctx, req.ID)
	if err != nil {
		logx.Errorf("查询子部门失败: %v", err)
		return utils.Response.InternalError("查询子部门失败"), err
	}

	if len(subDepartmentCount) > 0 {
		return utils.Response.BusinessError("The department still has employees and cannot be deleted."), nil
	}

	// 软删除部门
	err = l.svcCtx.DepartmentModel.SoftDelete(l.ctx, req.ID)
	if err != nil {
		logx.Errorf("删除部门失败: %v", err)
		return utils.Response.InternalError("删除部门失败"), err
	}

	return utils.Response.Success("删除部门成功"), nil
}
