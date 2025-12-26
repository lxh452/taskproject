package employee

import (
	"context"

	"task_Project/model/user"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEmployeeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取员工列表
func NewGetEmployeeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEmployeeListLogic {
	return &GetEmployeeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetEmployeeListLogic) GetEmployeeList(req *types.EmployeeListRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	var errors []string
	req.Page, req.PageSize, errors = validator.ValidatePageParams(req.Page, req.PageSize)
	if len(errors) > 0 {
		return utils.Response.ValidationError(errors[0]), nil
	}

	// 获取当前用户信息
	if _, ok := utils.Common.GetCurrentUserID(l.ctx); !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 根据条件查询员工列表
	var employees []*user.Employee
	var total int64

	if req.CompanyID != "" {
		// 如果有公司ID，使用公司查询
		if req.DepartmentID != "" {
			// 如果有部门ID，按部门查询
			employees, total, err = l.svcCtx.EmployeeModel.FindByDepartmentPage(l.ctx, req.DepartmentID, req.Page, req.PageSize)
		} else if req.PositionID != "" {
			// 如果有职务ID，按职务查询
			allEmployees, err := l.svcCtx.EmployeeModel.FindByPositionID(l.ctx, req.PositionID)
			if err != nil {
				logx.Errorf("查询员工列表失败: %v", err)
				return utils.Response.InternalError("查询员工列表失败"), err
			}
			// 过滤公司ID
			filteredEmployees := make([]*user.Employee, 0)
			for _, emp := range allEmployees {
				if emp.CompanyId == req.CompanyID {
					filteredEmployees = append(filteredEmployees, emp)
				}
			}
			total = int64(len(filteredEmployees))
			// 手动分页
			offset := (req.Page - 1) * req.PageSize
			end := offset + req.PageSize
			if end > len(filteredEmployees) {
				end = len(filteredEmployees)
			}
			if offset < len(filteredEmployees) {
				employees = filteredEmployees[offset:end]
			} else {
				employees = []*user.Employee{}
			}
		} else {
			// 只按公司查询
			employees, total, err = l.svcCtx.EmployeeModel.FindByCompanyPage(l.ctx, req.CompanyID, req.Page, req.PageSize)
		}
	} else {
		// 没有公司ID，使用通用分页查询
		employees, total, err = l.svcCtx.EmployeeModel.FindByPage(l.ctx, req.Page, req.PageSize)
	}

	if err != nil {
		logx.Errorf("查询员工列表失败: %v", err)
		return utils.Response.InternalError("查询员工列表失败"), err
	}

	// 转换为响应格式并添加任务数量
	converter := utils.NewConverter()
	employeeInfoList := converter.ToEmployeeInfoList(employees)

	// 批量获取用户头像（从MongoDB）
	avatarMap := make(map[string]string)
	if l.svcCtx.UploadFileModel != nil {
		for _, emp := range employees {
			avatarFiles, _ := l.svcCtx.UploadFileModel.FindByModuleAndRelatedID(l.ctx, "user", emp.UserId)
			for _, f := range avatarFiles {
				if f.Category == "avatar" {
					avatarMap[emp.UserId] = f.FileURL
					break
				}
			}
		}
	}

	// 构建包含任务数量的员工列表
	employeeList := make([]map[string]interface{}, 0, len(employeeInfoList))
	for _, empInfo := range employeeInfoList {
		employeeID := empInfo.ID
		// 查询员工参与的所有任务节点数量（去重，包括作为执行人和负责人的任务节点）
		taskCount, _ := l.svcCtx.TaskNodeModel.GetTaskNodeCountByEmployee(l.ctx, employeeID)

		// 获取职位信息（职位级别、职位代码）
		positionLevel := 0
		positionCode := ""
		isManagement := 0
		if empInfo.PositionID != "" {
			pos, err := l.svcCtx.PositionModel.FindOne(l.ctx, empInfo.PositionID)
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
		if empInfo.DepartmentID != "" {
			dept, err := l.svcCtx.DepartmentModel.FindOne(l.ctx, empInfo.DepartmentID)
			if err == nil && dept != nil {
				departmentPriority = int(dept.DepartmentPriority)
			}
		}

		// 检查是否是创始人（通过职位代码或公司Owner）
		isFounder := false
		if positionCode == "FOUNDER" {
			isFounder = true
		} else if empInfo.CompanyID != "" {
			company, err := l.svcCtx.CompanyModel.FindOne(l.ctx, empInfo.CompanyID)
			if err == nil && company != nil && company.Owner == empInfo.UserID {
				isFounder = true
			}
		}

		// 获取用户头像
		avatarURL := avatarMap[empInfo.UserID]

		// 构建包含任务数量的员工信息
		empMap := map[string]interface{}{
			"id":                 empInfo.ID,
			"userId":             empInfo.UserID,
			"companyId":          empInfo.CompanyID,
			"departmentId":       empInfo.DepartmentID,
			"positionId":         empInfo.PositionID,
			"employeeId":         empInfo.EmployeeID,
			"realName":           empInfo.RealName,
			"workEmail":          empInfo.WorkEmail,
			"workPhone":          empInfo.WorkPhone,
			"skills":             empInfo.Skills,
			"roleTags":           empInfo.RoleTags,
			"hireDate":           empInfo.HireDate,
			"leaveDate":          empInfo.LeaveDate,
			"status":             empInfo.Status,
			"createTime":         empInfo.CreateTime,
			"updateTime":         empInfo.UpdateTime,
			"taskCount":          taskCount,
			"positionLevel":      positionLevel,      // 职位级别
			"positionCode":       positionCode,       // 职位代码
			"isManagement":       isManagement,       // 是否管理岗
			"departmentPriority": departmentPriority, // 部门优先级
			"isFounder":          isFounder,          // 是否是创始人
			"avatar":             avatarURL,          // 用户头像URL
		}
		employeeList = append(employeeList, empMap)
	}

	// 构建分页响应
	pageResp := converter.ToPageResponse(employeeList, int(total), req.Page, req.PageSize)

	return utils.Response.SuccessWithKey("employees", pageResp), nil
}
