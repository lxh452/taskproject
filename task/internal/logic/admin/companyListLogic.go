package admin

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CompanyListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理员获取公司列表
func NewCompanyListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompanyListLogic {
	return &CompanyListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CompanyListLogic) CompanyList(req *types.AdminCompanyListRequest) (resp *types.BaseResponse, err error) {
	// 设置默认分页参数
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 查询公司列表
	var companies []types.AdminCompanyInfo
	var total int64

	if req.Name != "" {
		// 搜索公司
		companyList, count, err := l.svcCtx.CompanyModel.SearchCompanies(l.ctx, req.Name, page, pageSize)
		if err != nil {
			logx.Errorf("搜索公司失败: %v", err)
			return utils.Response.Error(500, "搜索公司失败"), nil
		}
		total = count

		for _, c := range companyList {
			// 获取每个公司的员工数量
			employeeCount, _ := l.svcCtx.EmployeeModel.GetEmployeeCountByCompany(l.ctx, c.Id)
			// 获取部门数量
			departmentCount, _ := l.svcCtx.DepartmentModel.GetDepartmentCountByCompany(l.ctx, c.Id)
			// 获取任务数量
			taskCount, _ := l.svcCtx.TaskModel.GetTaskCountByCompany(l.ctx, c.Id)

			companies = append(companies, types.AdminCompanyInfo{
				ID:                c.Id,
				Name:              c.Name,
				CompanyAttributes: int(c.CompanyAttributes),
				CompanyBusiness:   int(c.CompanyBusiness),
				Owner:             c.Owner,
				Description:       c.Description.String,
				Address:           c.Address.String,
				Phone:             c.Phone.String,
				Email:             c.Email.String,
				Status:            int(c.Status),
				EmployeeCount:     employeeCount,
				DepartmentCount:   departmentCount,
				TaskCount:         taskCount,
				CreateTime:        c.CreateTime.Format("2006-01-02 15:04:05"),
				UpdateTime:        c.UpdateTime.Format("2006-01-02 15:04:05"),
			})
		}
	} else {
		// 分页查询所有公司
		companyList, count, err := l.svcCtx.CompanyModel.FindByPage(l.ctx, page, pageSize)
		if err != nil {
			logx.Errorf("查询公司列表失败: %v", err)
			return utils.Response.Error(500, "查询公司列表失败"), nil
		}
		total = count

		for _, c := range companyList {
			// 获取每个公司的员工数量
			employeeCount, _ := l.svcCtx.EmployeeModel.GetEmployeeCountByCompany(l.ctx, c.Id)
			// 获取部门数量
			departmentCount, _ := l.svcCtx.DepartmentModel.GetDepartmentCountByCompany(l.ctx, c.Id)
			// 获取任务数量
			taskCount, _ := l.svcCtx.TaskModel.GetTaskCountByCompany(l.ctx, c.Id)

			companies = append(companies, types.AdminCompanyInfo{
				ID:                c.Id,
				Name:              c.Name,
				CompanyAttributes: int(c.CompanyAttributes),
				CompanyBusiness:   int(c.CompanyBusiness),
				Owner:             c.Owner,
				Description:       c.Description.String,
				Address:           c.Address.String,
				Phone:             c.Phone.String,
				Email:             c.Email.String,
				Status:            int(c.Status),
				EmployeeCount:     employeeCount,
				DepartmentCount:   departmentCount,
				TaskCount:         taskCount,
				CreateTime:        c.CreateTime.Format("2006-01-02 15:04:05"),
				UpdateTime:        c.UpdateTime.Format("2006-01-02 15:04:05"),
			})
		}
	}

	// 按状态筛选
	if req.Status != 0 {
		var filtered []types.AdminCompanyInfo
		for _, c := range companies {
			if c.Status == req.Status {
				filtered = append(filtered, c)
			}
		}
		companies = filtered
	}

	return utils.Response.SuccessWithData(map[string]interface{}{
		"list":     companies,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}), nil
}
