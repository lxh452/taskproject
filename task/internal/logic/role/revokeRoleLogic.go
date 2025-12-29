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

type RevokeRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 从职位撤销角色
func NewRevokeRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeRoleLogic {
	return &RevokeRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RevokeRoleLogic) RevokeRole(req *types.RevokeRoleRequest) (resp *types.BaseResponse, err error) {
	if utils.Validator.IsEmpty(req.PositionId) || utils.Validator.IsEmpty(req.RoleId) {
		return utils.Response.ValidationError("positionId/roleId required"), nil
	}
	rec, err := l.svcCtx.PositionRoleModel.FindOneByPositionIdRoleId(l.ctx, req.PositionId, req.RoleId)
	if err != nil {
		return utils.Response.BusinessError("data_not_found"), nil
	}
	if err := l.svcCtx.PositionRoleModel.Delete(l.ctx, rec.Id); err != nil {
		return utils.Response.InternalError("revoke failed"), nil
	}
	return utils.Response.SuccessWithKey("operation", nil), nil
}
