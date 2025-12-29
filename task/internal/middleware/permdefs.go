package middleware

// 数字化权限字典（细分权限点）
const (
	// 任务模块 (1-9)
	PermTaskRead    = 1
	PermTaskCreate  = 2
	PermTaskUpdate  = 3
	PermTaskDelete  = 4
	PermTaskApprove = 5

	// 任务节点 (10-19)
	PermTaskNodeRead   = 10
	PermTaskNodeUpdate = 11
	PermTaskNodeCreate = 12
	PermTaskNodeDelete = 13

	// 交接 (20-29)
	PermHandoverRead    = 20
	PermHandoverCreate  = 21
	PermHandoverApprove = 22
	PermHandoverReject  = 23

	// 通知 (30-39)
	PermNotificationRead   = 30
	PermNotificationCreate = 31
	PermNotificationDelete = 32

	// 组织 - 公司 (40-44)
	PermCompanyRead   = 40
	PermCompanyCreate = 41
	PermCompanyUpdate = 42
	PermCompanyDelete = 43

	// 组织 - 部门 (45-49)
	PermDepartmentRead   = 45
	PermDepartmentCreate = 46
	PermDepartmentUpdate = 47
	PermDepartmentDelete = 48

	// 组织 - 职位 (50-54)
	PermPositionRead   = 50
	PermPositionCreate = 51
	PermPositionUpdate = 52
	PermPositionDelete = 53

	// 角色 (60-65)
	PermRoleRead   = 60
	PermRoleCreate = 61
	PermRoleUpdate = 62
	PermRoleDelete = 63
	PermRoleAssign = 64
	PermRoleRevoke = 65

	// 员工 (70-74)
	PermEmployeeRead   = 70
	PermEmployeeCreate = 71
	PermEmployeeUpdate = 72
	PermEmployeeDelete = 73
	PermEmployeeLeave  = 74
)

// GetValidPermCodes 返回所有有效权限码集合（用于验证）
func GetValidPermCodes() map[int]struct{} {
	// 仅保留数字码集合
	all := []int{
		PermTaskRead, PermTaskCreate, PermTaskUpdate, PermTaskDelete, PermTaskApprove,
		PermTaskNodeRead, PermTaskNodeUpdate, PermTaskNodeCreate, PermTaskNodeDelete,
		PermHandoverRead, PermHandoverCreate, PermHandoverApprove, PermHandoverReject,
		PermNotificationRead, PermNotificationCreate, PermNotificationDelete,
		PermCompanyRead, PermCompanyCreate, PermCompanyUpdate, PermCompanyDelete,
		PermDepartmentRead, PermDepartmentCreate, PermDepartmentUpdate, PermDepartmentDelete,
		PermPositionRead, PermPositionCreate, PermPositionUpdate, PermPositionDelete,
		PermRoleRead, PermRoleCreate, PermRoleUpdate, PermRoleDelete, PermRoleAssign, PermRoleRevoke,
		PermEmployeeRead, PermEmployeeCreate, PermEmployeeUpdate, PermEmployeeDelete, PermEmployeeLeave,
	}
	valid := make(map[int]struct{}, len(all))
	for _, p := range all {
		valid[p] = struct{}{}
	}
	return valid
}
