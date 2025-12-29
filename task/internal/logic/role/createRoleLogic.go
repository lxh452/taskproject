// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"context"

	"task_Project/model/role"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建角色
func NewCreateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRoleLogic {
	return &CreateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateRoleLogic) CreateRole(req *types.CreateRoleRequest) (resp *types.BaseResponse, err error) {
	// 参数校验
	if utils.Validator.IsEmpty(req.CompanyId) || utils.Validator.IsEmpty(req.RoleName) || utils.Validator.IsEmpty(req.RoleCode) {
		return utils.Response.ValidationError("companyId/roleName/roleCode required"), nil
	}

	// 验证权限是否在字典中
	if resp, err := ValidatePermissions(req.Permissions); resp != nil || err != nil {
		if resp != nil {
			return resp, nil
		}
		return nil, err
	}

	data := &role.Role{
		Id:              utils.Common.GenId("role"),
		CompanyId:       req.CompanyId,
		RoleName:        req.RoleName,
		RoleCode:        req.RoleCode,
		RoleDescription: utils.Common.ToSqlNullString(req.Description),
		IsSystem:        0,
		Permissions:     utils.Common.ToSqlNullString(req.Permissions),
		Status:          1,
	}
	if _, err := l.svcCtx.RoleModel.Insert(l.ctx, data); err != nil {
		return utils.Response.InternalError("create role failed"), nil
	}
	return utils.Response.SuccessWithKey("create", map[string]string{"id": data.Id}), nil
}
