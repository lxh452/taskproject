// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"context"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

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

// todo 这里应该是获取用户登录时候加入的公司列表 而不是所有公司的列表用户用户切换公司  根据l.ctx获取用户id还有他加入的公司
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
	emInfo, err := l.svcCtx.EmployeeModel.FindOneByUserId(l.ctx, userID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查找该用户公司信息失败: %v", err)
		return nil, err
	}
	companyInfo, err := l.svcCtx.CompanyModel.FindOne(l.ctx, emInfo.CompanyId)
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		l.Logger.WithContext(l.ctx).Errorf("查找公司错误: %v", err)
		return nil, err
	}

	//如果是某公司的员工也要展示
	filteredCompanies := []*company.Company{}
	if companyInfo != nil {
		filteredCompanies = append(filteredCompanies, companyInfo)
	}
	for _, comp := range companies {
		fmt.Println("公司", comp.Owner)
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
