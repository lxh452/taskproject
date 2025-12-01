// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"context"
	"time"

	"task_Project/model/company"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCompanyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建公司
func NewCreateCompanyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCompanyLogic {
	return &CreateCompanyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCompanyLogic) CreateCompany(req *types.CreateCompanyRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	if utils.Validator.IsEmpty(req.Name) {
		return utils.Response.BusinessError("company_name_required"), nil
	}

	// 获取当前用户ID
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 检查公司名称是否已存在
	existingCompanies, err := l.svcCtx.CompanyModel.FindByOwner(l.ctx, userID)
	if err != nil {
		logx.Errorf("查询用户公司失败: %v", err)
		return utils.Response.InternalError("查询用户公司失败"), nil
	}

	for _, existingCompany := range existingCompanies {
		if existingCompany.Name == req.Name {
			return utils.Response.BusinessError("company_name_exists"), nil
		}
	}

	// 生成公司ID
	companyID := utils.Common.GenerateID()
	// 创建公司
	companyInfo := &company.Company{
		Id:                companyID,
		Name:              req.Name,
		CompanyAttributes: int64(req.CompanyAttributes),
		CompanyBusiness:   int64(req.CompanyBusiness),
		Owner:             userID,
		Description:       utils.Common.ToSqlNullString(req.Description),
		Address:           utils.Common.ToSqlNullString(req.Address),
		Phone:             utils.Common.ToSqlNullString(req.Phone),
		Email:             utils.Common.ToSqlNullString(req.Email),
		Status:            1, // 正常状态
		CreateTime:        time.Now(),
		UpdateTime:        time.Now(),
	}

	// 插入公司数据
	_, err = l.svcCtx.CompanyModel.Insert(l.ctx, companyInfo)
	if err != nil {
		logx.Errorf("创建公司失败: %v", err)
		return utils.Response.InternalError("创建公司失败"), nil
	}

	// 复制系统默认部门与职位
	if err := l.svcCtx.ApplyDefaultOrgStructure(l.ctx, companyID); err != nil {
		logx.Errorf("初始化公司组织结构失败: %v", err)
		return utils.Response.InternalError("初始化公司组织结构失败"), nil
	}

	// 发送创建成功通知邮件
	go func() {
		// 这里可以发送邮件通知用户公司创建成功
		logx.Infof("用户 %s 创建公司成功: %s", userID, req.Name)
	}()

	return utils.Response.SuccessWithKey("create", map[string]interface{}{
		"companyId":   companyID,
		"name":        req.Name,
		"bootstraped": true,
	}), nil
}
