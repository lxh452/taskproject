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

type GetDepartmentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取部门列表
func NewGetDepartmentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDepartmentListLogic {
	return &GetDepartmentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDepartmentListLogic) GetDepartmentList(req *types.DepartmentListRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if _, _, errors := validator.ValidatePageParams(req.Page, req.PageSize); err != nil {
		return utils.Response.ValidationError(errors[0]), nil
	}

	// 获取当前用户信息
	if _, ok := utils.Common.GetCurrentUserID(l.ctx); !ok {
		return utils.Response.UnauthorizedError(), nil
	}
	if req.CompanyID == "" {
		l.Logger.WithContext(l.ctx).Errorf("公司id不能为空")
		return utils.Response.NotFoundError("公司id不能为空"), nil
	}
	// 查询部门列表
	departments, total, err := l.svcCtx.DepartmentModel.FindByPageCompany(l.ctx, req.CompanyID, req.Page, req.PageSize)
	if err != nil {
		logx.Errorf("查询部门列表失败: %v", err)
		return utils.Response.InternalError("查询部门列表失败"), err
	}

	// 转换为响应格式
	converter := utils.NewConverter()
	departmentList := converter.ToDepartmentInfoList(departments)

	// 构建分页响应
	pageResp := utils.NewConverter().ToPageResponse(departmentList, int(total), req.Page, req.PageSize)

	return utils.Response.SuccessWithKey("departments", pageResp), nil
}
