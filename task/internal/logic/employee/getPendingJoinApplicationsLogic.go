package employee

import (
	"context"

	"task_Project/model/user"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPendingJoinApplicationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取待审批的加入申请列表
func NewGetPendingJoinApplicationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPendingJoinApplicationsLogic {
	return &GetPendingJoinApplicationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPendingJoinApplicationsLogic) GetPendingJoinApplications(req *types.GetPendingJoinApplicationsRequest) (resp *types.BaseResponse, err error) {
	// 获取当前用户ID
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 获取当前员工信息
	employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, userID)
	if err != nil || employee == nil {
		return utils.Response.BusinessError("您尚未加入任何公司"), nil
	}

	// 检查审批权限
	company, _ := l.svcCtx.CompanyModel.FindOne(l.ctx, employee.CompanyId)
	isFounder := company != nil && company.Owner == userID
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
		return utils.Response.BusinessError("只有公司创始人、人事部门或管理人员可以查看"), nil
	}

	// 查询待审批的申请
	status := user.JoinApplicationStatusPending
	applications, err := l.svcCtx.JoinApplicationModel.FindByCompanyId(l.ctx, employee.CompanyId, &status)
	if err != nil {
		logx.Errorf("查询申请列表失败: %v", err)
		return utils.Response.InternalError("查询失败"), nil
	}

	// 构建响应数据
	list := make([]types.JoinApplicationInfo, 0, len(applications))
	for _, app := range applications {
		// 获取申请人信息
		applicantUser, _ := l.svcCtx.UserModel.FindOne(l.ctx, app.UserId)
		username := ""
		realName := ""
		if applicantUser != nil {
			username = applicantUser.Username
			if applicantUser.RealName.Valid {
				realName = applicantUser.RealName.String
			}
		}

		info := types.JoinApplicationInfo{
			ID:          app.Id,
			UserID:      app.UserId,
			Username:    username,
			RealName:    realName,
			CompanyID:   app.CompanyId,
			CompanyName: company.Name,
			Status:      int(app.Status),
			CreateTime:  app.CreateTime.Format("2006-01-02 15:04:05"),
		}

		if app.InviteCode.Valid {
			info.InviteCode = app.InviteCode.String
		}
		if app.ApplyReason.Valid {
			info.ApplyReason = app.ApplyReason.String
		}

		list = append(list, info)
	}

	return utils.Response.Success(map[string]interface{}{
		"list":  list,
		"total": len(list),
	}), nil
}

