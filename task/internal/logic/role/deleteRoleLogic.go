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

type DeleteRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除角色
func NewDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteRoleLogic) DeleteRole(req *types.DeleteRoleRequest) (resp *types.BaseResponse, err error) {
	if utils.Validator.IsEmpty(req.Id) {
		return utils.Response.ValidationError("id required"), nil
	}
	if err := l.svcCtx.RoleModel.Delete(l.ctx, req.Id); err != nil {
		return utils.Response.InternalError("delete role failed"), nil
	}
	return utils.Response.SuccessWithKey("delete", nil), nil
}
