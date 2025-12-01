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

type RoleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 角色列表
func NewRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RoleListLogic {
	return &RoleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RoleListLogic) RoleList(req *types.RoleListRequest) (resp *types.BaseResponse, err error) {
	// 校验分页
	validator := utils.NewValidator()
	page, pageSize, errs := validator.ValidatePageParams(req.Page, req.PageSize)
	if len(errs) > 0 {
		return utils.Response.ValidationError(errs[0]), nil
	}

	var (
		list  []*role.Role
		total int64
		e     error
	)
	if req.CompanyId != "" {
		list, total, e = l.svcCtx.RoleModel.FindByCompanyPage(l.ctx, req.CompanyId, page, pageSize)
	} else if req.Keyword != "" {
		list, total, e = l.svcCtx.RoleModel.SearchRoles(l.ctx, req.Keyword, page, pageSize)
	} else {
		list, total, e = l.svcCtx.RoleModel.FindByPage(l.ctx, page, pageSize)
	}
	if e != nil {
		return utils.Response.InternalError("list roles failed"), nil
	}

	return utils.Response.Success(map[string]interface{}{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}), nil
}
