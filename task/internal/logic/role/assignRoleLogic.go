// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package role

import (
	"context"
	"database/sql"
	"time"

	"task_Project/model/role"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssignRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 给职位赋予角色
func NewAssignRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRoleLogic {
	return &AssignRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssignRoleLogic) AssignRole(req *types.AssignRoleRequest) (resp *types.BaseResponse, err error) {
	if utils.Validator.IsEmpty(req.PositionId) || utils.Validator.IsEmpty(req.RoleId) {
		return utils.Response.ValidationError("positionId/roleId required"), nil
	}

	// 验证职位是否存在
	_, err = l.svcCtx.PositionModel.FindOne(l.ctx, req.PositionId)
	if err != nil {
		return utils.Response.BusinessError("Format error"), nil
	}

	// 验证角色是否存在
	_, err = l.svcCtx.RoleModel.FindOne(l.ctx, req.RoleId)
	if err != nil {
		return utils.Response.BusinessError("Format error"), nil
	}

	// 检查是否已经分配过
	_, err = l.svcCtx.PositionRoleModel.FindOneByPositionIdRoleId(l.ctx, req.PositionId, req.RoleId)
	if err == nil {
		return utils.Response.BusinessError("Format error"), nil
	}

	var expire sql.NullTime
	if req.ExpireTime != "" {
		if t, e := time.Parse("2006-01-02 15:04:05", req.ExpireTime); e == nil {
			expire = sql.NullTime{Time: t, Valid: true}
		}
	}
	uid, _ := utils.Common.GetCurrentUserID(l.ctx)
	data := &role.PositionRole{
		Id:         utils.Common.GenId("pr"),
		PositionId: req.PositionId,
		RoleId:     req.RoleId,
		GrantBy:    sql.NullString{String: uid, Valid: uid != ""},
		GrantTime:  time.Now(),
		ExpireTime: expire,
		Status:     1,
	}
	if _, err := l.svcCtx.PositionRoleModel.Insert(l.ctx, data); err != nil {
		return utils.Response.InternalError("assign role failed"), nil
	}
	return utils.Response.SuccessWithKey("operation", nil), nil
}
