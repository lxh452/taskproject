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

type DeleteCompanyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除公司
func NewDeleteCompanyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCompanyLogic {
	return &DeleteCompanyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCompanyLogic) DeleteCompany(req *types.DeleteCompanyRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.CompanyID) {
		return utils.Response.ValidationError("公司ID不能为空"), nil
	}

	// 检查公司是否存在
	if _, err = l.svcCtx.CompanyModel.FindOne(l.ctx, req.CompanyID); err != nil {
		logx.Errorf("查询公司失败: %v", err)
		return utils.Response.ErrorWithKey("company_not_found"), nil
	}
	// 校验该公司是否是该用户的公司

	// 检查公司是否有员工
	employeeCount, err := l.svcCtx.EmployeeModel.GetEmployeeCountByCompany(l.ctx, req.CompanyID)
	if err != nil {
		logx.Errorf("查询公司员工数量失败: %v", err)
		return utils.Response.InternalError("查询公司员工数量失败"), err
	}

	if employeeCount > 0 {
		return utils.Response.BusinessError("The company still has employees and cannot be deleted."), nil
	}

	// 软删除公司
	err = l.svcCtx.CompanyModel.SoftDelete(l.ctx, req.CompanyID)
	if err != nil {
		logx.Errorf("删除公司失败: %v", err)
		return utils.Response.InternalError("删除公司失败"), err
	}

	return utils.Response.Success("删除公司成功"), nil
}
