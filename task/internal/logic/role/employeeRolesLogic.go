// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type EmployeeRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询员工的角色列表
func NewEmployeeRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EmployeeRolesLogic {
	return &EmployeeRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EmployeeRolesLogic) EmployeeRoles(req *types.EmployeeRolesRequest) (resp *types.BaseResponse, err error) {
	if utils.Validator.IsEmpty(req.EmployeeId) {
		return utils.Response.ValidationError("employeeId required"), nil
	}

	// 查询员工信息，获取职位ID
	employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, req.EmployeeId)
	if err != nil {
		return utils.Response.BusinessError("员工不存在"), nil
	}

	// 如果员工没有职位，返回空列表
	if !employee.PositionId.Valid || employee.PositionId.String == "" {
		return utils.Response.Success(map[string]interface{}{
			"list": []interface{}{},
		}), nil
	}

	// 通过职位查询角色（员工->职位->角色）
	roles, err := l.svcCtx.PositionRoleModel.ListRolesByEmployeeId(l.ctx, req.EmployeeId)
	if err != nil {
		return utils.Response.InternalError("query employee roles failed"), nil
	}

	type roleView struct {
		Id          string `json:"id"`
		RoleName    string `json:"roleName"`
		RoleCode    string `json:"roleCode"`
		Permissions string `json:"permissions"`
		Status      int64  `json:"status"`
	}
	result := make([]roleView, 0, len(roles))
	for _, r := range roles {
		result = append(result, roleView{
			Id:          r.Id,
			RoleName:    r.RoleName,
			RoleCode:    r.RoleCode,
			Permissions: r.Permissions.String,
			Status:      r.Status,
		})
	}
	return utils.Response.Success(map[string]interface{}{
		"list": result,
	}), nil
}
