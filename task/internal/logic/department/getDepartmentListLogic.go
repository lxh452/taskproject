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
	req.Page, req.PageSize, _ = validator.ValidatePageParams(req.Page, req.PageSize)

	// 获取当前用户信息
	if _, ok := utils.Common.GetCurrentUserID(l.ctx); !ok {
		return utils.Response.UnauthorizedError(), nil
	}
	if req.CompanyID == "" {
		l.Logger.WithContext(l.ctx).Errorf("公司id不能为空")
		return utils.Response.NotFoundError("公司id不能为空"), nil
	}
	l.Logger.WithContext(l.ctx).Infof("开始查询部门列表，CompanyID: %s, Page: %d, PageSize: %d", req.CompanyID, req.Page, req.PageSize)

	// 查询部门列表
	departments, total, err := l.svcCtx.DepartmentModel.FindByPageCompany(l.ctx, req.CompanyID, req.Page, req.PageSize)
	if err != nil {
		logx.Errorf("查询部门列表失败：%v", err)
		return utils.Response.InternalError("查询部门列表失败"), err
	}

	l.Logger.WithContext(l.ctx).Infof("查询到 %d 条部门记录，总数：%d", len(departments), total)
	for i, dept := range departments {
		l.Logger.WithContext(l.ctx).Infof("部门 %d: ID=%s, Name=%s, CompanyID=%s", i, dept.Id, dept.DepartmentName, dept.CompanyId)
	}

	// 转换为响应格式
	converter := utils.NewConverter()
	departmentList := converter.ToDepartmentInfoList(departments)

	// 查询负责人姓名
	for i := range departmentList {
		if departmentList[i].ManagerID != "" {
			l.Logger.WithContext(l.ctx).Infof("查询部门 %s 的负责人 %s", departmentList[i].DepartmentName, departmentList[i].ManagerID)
			emp, err := l.svcCtx.EmployeeModel.FindByEmployeeID(l.ctx, departmentList[i].ManagerID)
			if err != nil {
				l.Logger.WithContext(l.ctx).Errorf("查询负责人失败：%v", err)
			} else if emp != nil {
				departmentList[i].ManagerName = emp.RealName
				l.Logger.WithContext(l.ctx).Infof("找到负责人：%s", emp.RealName)
			} else {
				l.Logger.WithContext(l.ctx).Infof("未找到负责人：%s", departmentList[i].ManagerID)
			}
		}
	}

	// 构建分页响应
	pageResp := utils.NewConverter().ToPageResponse(departmentList, int(total), req.Page, req.PageSize)

	return utils.Response.SuccessWithKey("departments", pageResp), nil
}
