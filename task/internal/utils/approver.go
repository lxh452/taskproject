package utils

import (
	"context"

	"task_Project/model/company"
	"task_Project/model/user"
)

// ApproverFinder 审批人查找器
type ApproverFinder struct {
	EmployeeModel   user.EmployeeModel
	DepartmentModel company.DepartmentModel
	CompanyModel    company.CompanyModel
	PositionModel   company.PositionModel
}

// NewApproverFinder 创建审批人查找器
func NewApproverFinder(employeeModel user.EmployeeModel, departmentModel company.DepartmentModel, companyModel company.CompanyModel) *ApproverFinder {
	return &ApproverFinder{
		EmployeeModel:   employeeModel,
		DepartmentModel: departmentModel,
		CompanyModel:    companyModel,
	}
}

// NewApproverFinderWithPosition 创建带职位模型的审批人查找器
func NewApproverFinderWithPosition(employeeModel user.EmployeeModel, departmentModel company.DepartmentModel, companyModel company.CompanyModel, positionModel company.PositionModel) *ApproverFinder {
	return &ApproverFinder{
		EmployeeModel:   employeeModel,
		DepartmentModel: departmentModel,
		CompanyModel:    companyModel,
		PositionModel:   positionModel,
	}
}

// ApproverResult 审批人查找结果
type ApproverResult struct {
	ApproverID   string // 审批人员工ID
	ApproverName string // 审批人姓名
	ApproverType string // 审批人类型: supervisor/department_manager/position_superior/founder
}

// FindApprover 查找审批人（自动推断）
// 优先级:
// 1. 直属上级（如果已设置）
// 2. 部门经理
// 3. 同部门内职位级别更高的员工（管理岗优先）
// 4. 上级部门的经理
// 5. 公司创始人
// 如果都找不到，返回空（不允许自己审批自己）
func (f *ApproverFinder) FindApprover(ctx context.Context, employeeID string) (*ApproverResult, error) {
	// 1. 获取员工信息
	employee, err := f.EmployeeModel.FindOne(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	// 2. 优先查找已设置的直属上级
	if employee.SupervisorId.Valid && employee.SupervisorId.String != "" {
		supervisor, err := f.EmployeeModel.FindOne(ctx, employee.SupervisorId.String)
		if err == nil && supervisor.Status == 1 { // 上级在职
			return &ApproverResult{
				ApproverID:   supervisor.Id,
				ApproverName: supervisor.RealName,
				ApproverType: "supervisor",
			}, nil
		}
	}

	// 3. 查找部门经理
	if employee.DepartmentId.Valid && employee.DepartmentId.String != "" {
		dept, err := f.DepartmentModel.FindOne(ctx, employee.DepartmentId.String)
		if err == nil && dept.ManagerId.Valid && dept.ManagerId.String != "" {
			// 确保部门经理不是员工自己
			if dept.ManagerId.String != employeeID {
				manager, err := f.EmployeeModel.FindOne(ctx, dept.ManagerId.String)
				if err == nil && manager.Status == 1 { // 经理在职
					return &ApproverResult{
						ApproverID:   manager.Id,
						ApproverName: manager.RealName,
						ApproverType: "department_manager",
					}, nil
				}
			}
		}

		// 4. 如果部门经理是自己或不存在，查找同部门内职位级别更高的员工
		if f.PositionModel != nil {
			superior := f.findPositionSuperior(ctx, employee)
			if superior != nil {
				return superior, nil
			}
		}

		// 5. 查找上级部门的经理
		if dept != nil && dept.ParentId.Valid && dept.ParentId.String != "" {
			parentDept, err := f.DepartmentModel.FindOne(ctx, dept.ParentId.String)
			if err == nil && parentDept.ManagerId.Valid && parentDept.ManagerId.String != "" {
				if parentDept.ManagerId.String != employeeID {
					parentManager, err := f.EmployeeModel.FindOne(ctx, parentDept.ManagerId.String)
					if err == nil && parentManager.Status == 1 {
						return &ApproverResult{
							ApproverID:   parentManager.Id,
							ApproverName: parentManager.RealName,
							ApproverType: "parent_department_manager",
						}, nil
					}
				}
			}
		}
	}

	// 6. 查找公司创始人
	companyInfo, err := f.CompanyModel.FindOne(ctx, employee.CompanyId)
	if err == nil && companyInfo.Owner != "" {
		// 通过 owner (user_id) 查找创始人员工信息
		founder, err := f.EmployeeModel.FindByUserID(ctx, companyInfo.Owner)
		if err == nil && founder.Id != employeeID && founder.Status == 1 { // 创始人在职且不是自己
			return &ApproverResult{
				ApproverID:   founder.Id,
				ApproverName: founder.RealName,
				ApproverType: "founder",
			}, nil
		}
	}

	// 7. 找不到合适的审批人
	return nil, nil
}

// findPositionSuperior 查找同部门内职位级别更高的员工
func (f *ApproverFinder) findPositionSuperior(ctx context.Context, employee *user.Employee) *ApproverResult {
	if !employee.DepartmentId.Valid || employee.DepartmentId.String == "" {
		return nil
	}

	// 获取当前员工的职位级别
	var currentLevel int64 = 0
	if employee.PositionId.Valid && employee.PositionId.String != "" {
		pos, err := f.PositionModel.FindOne(ctx, employee.PositionId.String)
		if err == nil {
			currentLevel = pos.PositionLevel
		}
	}

	// 查找同部门的所有员工
	deptEmployees, err := f.EmployeeModel.FindByDepartmentID(ctx, employee.DepartmentId.String)
	if err != nil {
		return nil
	}

	// 找到职位级别最高且比当前员工高的在职员工
	var bestCandidate *user.Employee
	var bestLevel int64 = currentLevel
	var bestIsManagement int64 = 0

	for _, emp := range deptEmployees {
		// 跳过自己和离职员工
		if emp.Id == employee.Id || emp.Status != 1 {
			continue
		}

		// 获取该员工的职位信息
		if !emp.PositionId.Valid || emp.PositionId.String == "" {
			continue
		}

		pos, err := f.PositionModel.FindOne(ctx, emp.PositionId.String)
		if err != nil {
			continue
		}

		// 比较职位级别（级别越高数值越大）
		// 优先选择管理岗，其次选择级别更高的
		if pos.PositionLevel > bestLevel ||
			(pos.PositionLevel == bestLevel && pos.IsManagement > bestIsManagement) {
			bestCandidate = emp
			bestLevel = pos.PositionLevel
			bestIsManagement = pos.IsManagement
		}
	}

	if bestCandidate != nil {
		return &ApproverResult{
			ApproverID:   bestCandidate.Id,
			ApproverName: bestCandidate.RealName,
			ApproverType: "position_superior",
		}
	}

	return nil
}

// FindApproverForHandover 查找交接审批人
// 对于交接，需要考虑发起人和接收人双方的上级
func (f *ApproverFinder) FindApproverForHandover(ctx context.Context, fromEmployeeID, toEmployeeID string) (*ApproverResult, error) {
	// 优先使用发起人的上级
	result, err := f.FindApprover(ctx, fromEmployeeID)
	if err == nil && result != nil {
		return result, nil
	}

	// 如果发起人没有上级，使用接收人的上级
	return f.FindApprover(ctx, toEmployeeID)
}

// CanApprove 检查某人是否有权审批某员工的申请
func (f *ApproverFinder) CanApprove(ctx context.Context, approverID, employeeID string) (bool, string) {
	// 获取员工信息
	employee, err := f.EmployeeModel.FindOne(ctx, employeeID)
	if err != nil {
		return false, ""
	}

	// 1. 检查是否是直属上级
	if employee.SupervisorId.Valid && employee.SupervisorId.String == approverID {
		return true, "supervisor"
	}

	// 2. 检查是否是部门经理
	if employee.DepartmentId.Valid && employee.DepartmentId.String != "" {
		dept, err := f.DepartmentModel.FindOne(ctx, employee.DepartmentId.String)
		if err == nil && dept.ManagerId.Valid && dept.ManagerId.String == approverID {
			return true, "department_manager"
		}

		// 3. 检查是否是上级部门的经理
		if dept != nil && dept.ParentId.Valid && dept.ParentId.String != "" {
			parentDept, err := f.DepartmentModel.FindOne(ctx, dept.ParentId.String)
			if err == nil && parentDept.ManagerId.Valid && parentDept.ManagerId.String == approverID {
				return true, "parent_department_manager"
			}
		}
	}

	// 4. 检查是否是同部门内职位级别更高的员工
	if f.PositionModel != nil && employee.DepartmentId.Valid {
		approver, err := f.EmployeeModel.FindOne(ctx, approverID)
		if err == nil && approver.DepartmentId.Valid && approver.DepartmentId.String == employee.DepartmentId.String {
			// 同部门，比较职位级别
			if employee.PositionId.Valid && approver.PositionId.Valid {
				empPos, err1 := f.PositionModel.FindOne(ctx, employee.PositionId.String)
				approverPos, err2 := f.PositionModel.FindOne(ctx, approver.PositionId.String)
				if err1 == nil && err2 == nil {
					if approverPos.PositionLevel > empPos.PositionLevel ||
						(approverPos.PositionLevel == empPos.PositionLevel && approverPos.IsManagement > empPos.IsManagement) {
						return true, "position_superior"
					}
				}
			}
		}
	}

	// 5. 检查是否是公司创始人
	companyInfo, err := f.CompanyModel.FindOne(ctx, employee.CompanyId)
	if err == nil {
		founder, err := f.EmployeeModel.FindByUserID(ctx, companyInfo.Owner)
		if err == nil && founder.Id == approverID {
			return true, "founder"
		}
	}

	return false, ""
}

// InferSupervisor 自动推断员工的直属上级
// 用于员工入职时自动设置上级
// 规则：
// 1. 如果部门有经理且不是自己，则经理是上级
// 2. 如果自己是部门经理，则上级部门的经理是上级
// 3. 如果都没有，则公司创始人是上级（除非自己是创始人）
func (f *ApproverFinder) InferSupervisor(ctx context.Context, employeeID string) (string, error) {
	employee, err := f.EmployeeModel.FindOne(ctx, employeeID)
	if err != nil {
		return "", err
	}

	// 1. 检查部门经理
	if employee.DepartmentId.Valid && employee.DepartmentId.String != "" {
		dept, err := f.DepartmentModel.FindOne(ctx, employee.DepartmentId.String)
		if err == nil {
			// 如果部门有经理且不是自己
			if dept.ManagerId.Valid && dept.ManagerId.String != "" && dept.ManagerId.String != employeeID {
				manager, err := f.EmployeeModel.FindOne(ctx, dept.ManagerId.String)
				if err == nil && manager.Status == 1 {
					return manager.Id, nil
				}
			}

			// 如果自己是部门经理，查找上级部门的经理
			if dept.ManagerId.Valid && dept.ManagerId.String == employeeID {
				if dept.ParentId.Valid && dept.ParentId.String != "" {
					parentDept, err := f.DepartmentModel.FindOne(ctx, dept.ParentId.String)
					if err == nil && parentDept.ManagerId.Valid && parentDept.ManagerId.String != "" {
						parentManager, err := f.EmployeeModel.FindOne(ctx, parentDept.ManagerId.String)
						if err == nil && parentManager.Status == 1 {
							return parentManager.Id, nil
						}
					}
				}
			}
		}
	}

	// 2. 查找公司创始人
	companyInfo, err := f.CompanyModel.FindOne(ctx, employee.CompanyId)
	if err == nil && companyInfo.Owner != "" {
		founder, err := f.EmployeeModel.FindByUserID(ctx, companyInfo.Owner)
		if err == nil && founder.Id != employeeID && founder.Status == 1 {
			return founder.Id, nil
		}
	}

	return "", nil
}
