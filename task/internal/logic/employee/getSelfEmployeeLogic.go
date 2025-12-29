package employee

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSelfEmployeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSelfEmployeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSelfEmployeeLogic {
	return &GetSelfEmployeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSelfEmployeeLogic) GetSelfEmployee() (resp *types.BaseResponse, err error) {
	// 从 JWT 上下文获取 userId
	userId, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok || userId == "" {
		return utils.Response.UnauthorizedError(), nil
	}

	// 通过 userId 查找员工信息
	employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, userId)
	if err != nil {
		logx.Errorf("查找员工失败: %v", err)
		return utils.Response.BusinessError("employee_not_found"), nil
	}
	if employee == nil {
		return utils.Response.BusinessError("employee_not_found"), nil
	}

	// 转换为响应格式
	converter := utils.NewConverter()
	employeeInfo := converter.ToEmployeeInfo(employee)

	// 获取职位信息（职位级别、职位代码）
	positionLevel := 0
	positionCode := ""
	isManagement := 0
	if employee.PositionId.Valid && employee.PositionId.String != "" {
		pos, err := l.svcCtx.PositionModel.FindOne(l.ctx, employee.PositionId.String)
		if err == nil && pos != nil {
			positionLevel = int(pos.PositionLevel)
			if pos.PositionCode.Valid {
				positionCode = pos.PositionCode.String
			}
			isManagement = int(pos.IsManagement)
		}
	}

	// 获取部门信息（部门优先级）
	departmentPriority := 0
	if employee.DepartmentId.Valid && employee.DepartmentId.String != "" {
		dept, err := l.svcCtx.DepartmentModel.FindOne(l.ctx, employee.DepartmentId.String)
		if err == nil && dept != nil {
			departmentPriority = int(dept.DepartmentPriority)
		}
	}

	// 检查是否是创始人（通过职位代码或公司Owner）
	isFounder := false
	if positionCode == "FOUNDER" {
		isFounder = true
	} else if employee.CompanyId != "" {
		company, err := l.svcCtx.CompanyModel.FindOne(l.ctx, employee.CompanyId)
		if err == nil && company != nil && company.Owner == employee.UserId {
			isFounder = true
		}
	}

	// 获取用户头像（从MongoDB的upload_file集合中查询）
	avatarURL := ""
	if l.svcCtx.UploadFileModel != nil {
		avatarFiles, _ := l.svcCtx.UploadFileModel.FindByModuleAndRelatedID(l.ctx, "user", userId)
		for _, f := range avatarFiles {
			if f.Category == "avatar" {
				avatarURL = f.FileURL
				break
			}
		}
	}

	// 构建包含额外信息的员工信息
	empMap := map[string]interface{}{
		"id":                 employeeInfo.ID,
		"userId":             employeeInfo.UserID,
		"companyId":          employeeInfo.CompanyID,
		"departmentId":       employeeInfo.DepartmentID,
		"positionId":         employeeInfo.PositionID,
		"employeeId":         employeeInfo.EmployeeID,
		"realName":           employeeInfo.RealName,
		"workEmail":          employeeInfo.WorkEmail,
		"workPhone":          employeeInfo.WorkPhone,
		"skills":             employeeInfo.Skills,
		"roleTags":           employeeInfo.RoleTags,
		"hireDate":           employeeInfo.HireDate,
		"leaveDate":          employeeInfo.LeaveDate,
		"status":             employeeInfo.Status,
		"createTime":         employeeInfo.CreateTime,
		"updateTime":         employeeInfo.UpdateTime,
		"positionLevel":      positionLevel,      // 职位级别
		"positionCode":       positionCode,       // 职位代码
		"isManagement":       isManagement,       // 是否管理岗
		"departmentPriority": departmentPriority, // 部门优先级
		"isFounder":          isFounder,          // 是否是创始人
		"avatar":             avatarURL,          // 用户头像URL
	}

	return utils.Response.SuccessWithKey("employee", empMap), nil
}
