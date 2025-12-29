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

type PositionRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 查询职位的角色列表
func NewPositionRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PositionRolesLogic {
	return &PositionRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PositionRolesLogic) PositionRoles(req *types.PositionRolesRequest) (resp *types.BaseResponse, err error) {
	if utils.Validator.IsEmpty(req.PositionId) {
		return utils.Response.ValidationError("positionId required"), nil
	}

	// 验证职位是否存在
	_, err = l.svcCtx.PositionModel.FindOne(l.ctx, req.PositionId)
	if err != nil {
		return utils.Response.BusinessError("position_not_found"), nil
	}

	// 通过职位查询角色
	roles, err := l.svcCtx.PositionRoleModel.ListRolesByPositionId(l.ctx, req.PositionId)
	if err != nil {
		logx.Errorf("查询职位角色失败: %v", err)
		return utils.Response.InternalError("查询职位角色失败"), err
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
