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

type UpdateRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新角色
func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRoleLogic) UpdateRole(req *types.UpdateRoleRequest) (resp *types.BaseResponse, err error) {
	if utils.Validator.IsEmpty(req.Id) {
		return utils.Response.ValidationError("id required"), nil
	}
	current, err := l.svcCtx.RoleModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return utils.Response.BusinessError("role_not_found"), nil
	}
	if req.RoleName != "" {
		current.RoleName = req.RoleName
	}
	if req.RoleCode != "" {
		current.RoleCode = req.RoleCode
	}
	if req.Description != "" {
		current.RoleDescription = utils.Common.ToSqlNullString(req.Description)
	}
	if req.Permissions != "" {
		// 验证权限是否在字典中
		if resp, err := ValidatePermissions(req.Permissions); resp != nil || err != nil {
			if resp != nil {
				return resp, nil
			}
			return nil, err
		}
		current.Permissions = utils.Common.ToSqlNullString(req.Permissions)
	}
	if req.Status != 0 {
		current.Status = int64(req.Status)
	}
	if err := l.svcCtx.RoleModel.Update(l.ctx, current); err != nil {
		return utils.Response.InternalError("update role failed"), nil
	}
	return utils.Response.SuccessWithKey("update", nil), nil
}
