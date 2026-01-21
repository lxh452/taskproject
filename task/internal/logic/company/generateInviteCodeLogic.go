package company

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateInviteCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 生成公司邀请码
func NewGenerateInviteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateInviteCodeLogic {
	return &GenerateInviteCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateInviteCodeLogic) GenerateInviteCode(req *types.GenerateInviteCodeRequest) (resp *types.BaseResponse, err error) {
	// 获取当前用户ID
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 获取当前员工信息
	employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, userID)
	if err != nil || employee == nil {
		return utils.Response.BusinessError("employee_not_in_company"), nil
	}

	// 获取公司信息
	company, err := l.svcCtx.CompanyModel.FindOne(l.ctx, employee.CompanyId)
	if err != nil {
		logx.Errorf("查询公司失败: %v", err)
		return utils.Response.InternalError("查询公司失败"), nil
	}

	// 检查是否是公司创始人或人事部门
	isFounder := company.Owner == userID
	isHR := false

	// 检查是否是人事部门
	if employee.DepartmentId.Valid {
		dept, _ := l.svcCtx.DepartmentModel.FindOne(l.ctx, employee.DepartmentId.String)
		if dept != nil && dept.DepartmentCode.Valid && dept.DepartmentCode.String == "HR" {
			isHR = true
		}
	}

	// 检查是否是管理岗
	isManager := false
	if employee.PositionId.Valid {
		pos, _ := l.svcCtx.PositionModel.FindOne(l.ctx, employee.PositionId.String)
		if pos != nil && pos.IsManagement == 1 {
			isManager = true
		}
	}

	if !isFounder && !isHR && !isManager {
		return utils.Response.BusinessError("only_admin_can_generate"), nil
	}

	// 生成邀请码
	expireDays := req.ExpireDays
	if expireDays <= 0 {
		expireDays = 7
	}

	code, err := l.svcCtx.InviteCodeService.GenerateInviteCode(
		l.ctx,
		employee.CompanyId,
		company.Name,
		employee.Id,
		expireDays,
		req.MaxUses,
	)
	if err != nil {
		logx.Errorf("生成邀请码失败: %v", err)
		return utils.Response.InternalError("生成邀请码失败"), nil
	}

	logx.Infof("生成邀请码成功: code=%s, companyId=%s, employeeId=%s", code, employee.CompanyId, employee.Id)

	return utils.Response.Success(map[string]interface{}{
		"inviteCode":  code,
		"companyId":   employee.CompanyId,
		"companyName": company.Name,
		"expireDays":  expireDays,
		"maxUses":     req.MaxUses,
	}), nil
}
