package utils

import (
	"context"

	"task_Project/model/company"
	"task_Project/model/user"
)

// lxh

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
func (f *ApproverFinder) FindApprover(ctx context.Context, employeeID string) (*ApproverResult, error) {
	employee, err := f.EmployeeModel.FindOne(ctx, employeeID)
	if err != nil {
		return nil, err
	}

	// 1. 优先查找已设置的直属上级
	if employee.SupervisorId.Valid && employee.SupervisorId.String != "" {
		supervisor, err := f.EmployeeModel.FindOne(ctx, employee.SupervisorId.String)
		if err == nil && supervisor.Status == 1 {
			return &ApproverResult{
				ApproverID:   supervisor.Id,
				ApproverName: supervisor.RealName,
				ApproverType: "supervisor",
			}, nil
		}
	}

	// 2. 查找部门经理
	if employee.DepartmentId.Valid && employee.DepartmentId.String != "" {
		dept, err := f.DepartmentModel.FindOne(ctx, employee.DepartmentId.String)
		if err == nil && dept.ManagerId.Valid && dept.ManagerId.String != "" {
			if dept.ManagerId.String != employeeID {
				manager, err := f.EmployeeModel.FindOne(ctx, dept.ManagerId.String)
				if err == nil && manager.Status == 1 {
					return &ApproverResult{
						ApproverID:   manager.Id,
						ApproverName: manager.RealName,
						ApproverType: "department_manager",
					}, nil
				}
			}
		}

		// 3. 查找上级部门的经理
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

	// 4. 查找公司创始人
	companyInfo, err := f.CompanyModel.FindOne(ctx, employee.CompanyId)
	if err == nil && companyInfo.Owner != "" {
		founder, err := f.EmployeeModel.FindByUserID(ctx, companyInfo.Owner)
		if err == nil && founder.Id != employeeID && founder.Status == 1 {
			return &ApproverResult{
				ApproverID:   founder.Id,
				ApproverName: founder.RealName,
				ApproverType: "founder",
			}, nil
		}
	}

	return nil, nil
}

// FindApproverForHandover 查找交接审批人
func (f *ApproverFinder) FindApproverForHandover(ctx context.Context, fromEmployeeID, toEmployeeID string) (*ApproverResult, error) {
	result, err := f.FindApprover(ctx, fromEmployeeID)
	if err == nil && result != nil {
		return result, nil
	}
	return f.FindApprover(ctx, toEmployeeID)
}

// CanApprove 检查某人是否有权审批某员工的申请
func (f *ApproverFinder) CanApprove(ctx context.Context, approverID, employeeID string) (bool, string) {
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

	// 4. 检查是否是公司创始人
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
func (f *ApproverFinder) InferSupervisor(ctx context.Context, employeeID string) (string, error) {
	employee, err := f.EmployeeModel.FindOne(ctx, employeeID)
	if err != nil {
		return "", err
	}

	// 1. 检查部门经理
	if employee.DepartmentId.Valid && employee.DepartmentId.String != "" {
		dept, err := f.DepartmentModel.FindOne(ctx, employee.DepartmentId.String)
		if err == nil {
			if dept.ManagerId.Valid && dept.ManagerId.String != "" && dept.ManagerId.String != employeeID {
				manager, err := f.EmployeeModel.FindOne(ctx, dept.ManagerId.String)
				if err == nil && manager.Status == 1 {
					return manager.Id, nil
				}
			}

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
