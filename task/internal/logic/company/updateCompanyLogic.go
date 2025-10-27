// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"context"
	"task_Project/model/company"
	"time"

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

	// 更新公司信息
	updateData := l.updatedata(req)

	if len(updateData) == 0 {
		return utils.Response.ValidationError("没有需要更新的字段"), nil
	}

	updateData["update_time"] = time.Now()

	// 遍历转换成结构体
	var updateCompany company.Company
	if err = utils.Common.MapToStructWithMapstructure(updateData, &updateCompany); err != nil {
		logx.Errorf("转换结构体失败信息失败: %v", err)
		return utils.Response.ErrorWithKey("operation_failed"), err
	}
	updateCompany.Id = req.ID
	err = l.svcCtx.CompanyModel.Update(l.ctx, &updateCompany)
	if err != nil {
		logx.Errorf("更新公司信息失败: %v", err)
		return utils.Response.InternalError("更新公司信息失败"), err
	}

	return utils.Response.Success("更新公司信息成功"), nil
}

// 判断修改类型
func (l *UpdateCompanyLogic) updatedata(req *types.UpdateCompanyRequest) map[string]interface{} {
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
	return updateData
}
