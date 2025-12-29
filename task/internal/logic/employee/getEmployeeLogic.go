// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package employee

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEmployeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取员工信息
func NewGetEmployeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEmployeeLogic {
	return &GetEmployeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetEmployeeLogic) GetEmployee(req *types.GetEmployeeRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.EmployeeID) {
		return utils.Response.ValidationError("员工ID不能为空"), nil
	}

	// 查询员工信息
	employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, req.EmployeeID)
	if err != nil {
		logx.Errorf("查询员工失败: %v", err)
		return utils.Response.ErrorWithKey("employee_not_found"), nil
	}

	// 转换为响应格式
	converter := utils.NewConverter()
	employeeInfo := converter.ToEmployeeInfo(employee)

	// 获取用户头像（从MongoDB）
	avatarURL := ""
	if l.svcCtx.UploadFileModel != nil && employee.UserId != "" {
		avatarFiles, _ := l.svcCtx.UploadFileModel.FindByModuleAndRelatedID(l.ctx, "user", employee.UserId)
		for _, f := range avatarFiles {
			if f.Category == "avatar" {
				avatarURL = f.FileURL
				break
			}
		}
	}

	// 构建包含头像的员工信息
	empMap := map[string]interface{}{
		"id":           employeeInfo.ID,
		"userId":       employeeInfo.UserID,
		"companyId":    employeeInfo.CompanyID,
		"departmentId": employeeInfo.DepartmentID,
		"positionId":   employeeInfo.PositionID,
		"employeeId":   employeeInfo.EmployeeID,
		"realName":     employeeInfo.RealName,
		"workEmail":    employeeInfo.WorkEmail,
		"workPhone":    employeeInfo.WorkPhone,
		"skills":       employeeInfo.Skills,
		"roleTags":     employeeInfo.RoleTags,
		"hireDate":     employeeInfo.HireDate,
		"leaveDate":    employeeInfo.LeaveDate,
		"status":       employeeInfo.Status,
		"createTime":   employeeInfo.CreateTime,
		"updateTime":   employeeInfo.UpdateTime,
		"avatar":       avatarURL,
	}

	return utils.Response.SuccessWithKey("employee", empMap), nil
}
