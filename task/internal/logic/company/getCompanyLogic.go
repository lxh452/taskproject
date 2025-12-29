// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"context"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCompanyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// todo 获取用户登录的公司信息
func NewGetCompanyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCompanyLogic {
	return &GetCompanyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCompanyLogic) GetCompany(req *types.GetCompanyRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.CompanyID) {
		return utils.Response.ValidationError("公司ID不能为空"), nil
	}

	// 查询公司信息
	company, err := l.svcCtx.CompanyModel.FindOne(l.ctx, req.CompanyID)
	if err != nil {
		logx.Errorf("查询公司失败: %v", err)
		return utils.Response.ErrorWithKey("company_not_found"), nil
	}
	// 转换为响应格式
	converter := utils.NewConverter()
	companyInfo := converter.ToCompanyInfo(company)
	return utils.Response.SuccessWithKey("company", companyInfo), nil
}
