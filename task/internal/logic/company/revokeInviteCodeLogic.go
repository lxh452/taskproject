package company

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type RevokeInviteCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 撤销邀请码
func NewRevokeInviteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeInviteCodeLogic {
	return &RevokeInviteCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RevokeInviteCodeLogic) RevokeInviteCode(req *types.RevokeInviteCodeRequest) (resp *types.BaseResponse, err error) {
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

	// 检查权限
	isFounder := company.Owner == userID
	isHR := false
	isManager := false

	if employee.DepartmentId.Valid {
		dept, _ := l.svcCtx.DepartmentModel.FindOne(l.ctx, employee.DepartmentId.String)
		if dept != nil && dept.DepartmentCode.Valid && dept.DepartmentCode.String == "HR" {
			isHR = true
		}
	}

	if employee.PositionId.Valid {
		pos, _ := l.svcCtx.PositionModel.FindOne(l.ctx, employee.PositionId.String)
		if pos != nil && pos.IsManagement == 1 {
			isManager = true
		}
	}

	if !isFounder && !isHR && !isManager {
		return utils.Response.BusinessError("only_admin_can_revoke"), nil
	}

	// 撤销邀请码
	if err := l.svcCtx.InviteCodeService.RevokeInviteCode(l.ctx, req.InviteCode); err != nil {
		logx.Errorf("撤销邀请码失败: %v", err)
		return utils.Response.InternalError("撤销邀请码失败"), nil
	}

	return utils.Response.Success(nil), nil
}
