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

type UpdateCompanyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新公司信息
func NewUpdateCompanyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCompanyLogic {
	return &UpdateCompanyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateCompanyLogic) UpdateCompany(req *types.UpdateCompanyRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.ID) {
		return utils.Response.ValidationError("公司ID不能为空"), nil
	}

	// 检查公司是否存在
	if _, err := l.svcCtx.CompanyModel.FindOne(l.ctx, req.ID); err != nil {
		logx.Errorf("查询公司失败: %v", err)
		return utils.Response.ErrorWithKey("company_not_found"), nil
	}
	// todo 这里要更正，由于公司的拥有者才有权限修改

	// 构建更新数据
	updateData := make(map[string]interface{})
	if !utils.Validator.IsEmpty(req.Name) {
		updateData["name"] = req.Name
	}
	if req.CompanyAttributes > 0 {
		updateData["company_attributes"] = req.CompanyAttributes
	}
	if req.CompanyBusiness > 0 {
		updateData["company_business"] = req.CompanyBusiness
	}
	if !utils.Validator.IsEmpty(req.Description) {
		updateData["description"] = req.Description
	}
	if !utils.Validator.IsEmpty(req.Address) {
		updateData["address"] = req.Address
	}
	if !utils.Validator.IsEmpty(req.Phone) {
		updateData["phone"] = req.Phone
	}
	if !utils.Validator.IsEmpty(req.Email) {
		updateData["email"] = req.Email
	}

	if len(updateData) == 0 {
		return utils.Response.ValidationError("没有需要更新的字段"), nil
	}

	// 使用选择性更新
	err = l.svcCtx.CompanyModel.SelectiveUpdate(l.ctx, req.ID, updateData)
	if err != nil {
		logx.Errorf("更新公司信息失败: %v", err)
		return utils.Response.InternalError("更新公司信息失败"), err
	}

	return utils.Response.Success("更新公司信息成功"), nil
}
