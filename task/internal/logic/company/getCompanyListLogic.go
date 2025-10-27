// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"context"

	"task_Project/model/company"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCompanyListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取公司列表
func NewGetCompanyListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanyListLogic {
	return &GetCompanyListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCompanyListLogic) GetCompanyList(req *types.CompanyListRequest) (resp *types.BaseResponse, err error) {
	// 参数验证和默认值设置
	validator := utils.NewValidator()
	req.Page, req.PageSize, _ = validator.ValidatePageParams(req.Page, req.PageSize)

	// 获取当前用户ID
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	var companies []*company.Company
	var total int64

	// 根据搜索条件查询
	if req.Name != "" {
		// 搜索公司
		companies, total, err = l.svcCtx.CompanyModel.SearchCompanies(l.ctx, req.Name, req.Page, req.PageSize)
		//} else if req.Status != nil {
		//	// 根据状态查询
		//	companies, err = l.svcCtx.CompanyModel.FindByStatus(l.ctx, *req.Status)
		//	if err == nil {
		//		total = int64(len(companies))
		//		// 手动分页
		//		start := (req.Page - 1) * req.PageSize
		//		end := start + req.PageSize
		//		if start >= len(companies) {
		//			companies = []*company.Company{}
		//		} else if end > len(companies) {
		//			companies = companies[start:]
		//		} else {
		//			companies = companies[start:end]
		//		}
		//	}
	} else {
		// 分页查询所有公司
		companies, total, err = l.svcCtx.CompanyModel.FindByPage(l.ctx, req.Page, req.PageSize)
	}

	if err != nil {
		logx.Errorf("查询公司列表失败: %v", err)
		return utils.Response.InternalError("查询失败"), nil
	}

	// 过滤用户自己的公司（如果不是管理员）
	// 这里可以根据用户角色来决定是否只显示自己的公司
	filteredCompanies := []*company.Company{}
	for _, comp := range companies {
		if comp.Owner == userID {
			filteredCompanies = append(filteredCompanies, comp)
		}
	}

	// 转换为响应格式
	companyList := utils.Converter.ToCompanyInfoList(filteredCompanies)

	// 构建响应
	response := make(map[string]interface{})
	response["total"] = total
	response["companyList"] = companyList
	response["page"] = req.Page
	response["pageSize"] = req.PageSize
	return utils.Response.SuccessWithKey("query", response), nil
}
