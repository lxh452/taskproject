package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// AuthzMiddleware 细粒度权限校验
type AuthzDeps struct {
	FindEmployeeByUserID func(ctx context.Context, userId string) (interface {
		GetEmployeeId() string
		GetId() string
	}, error)
	ListRolesByEmployeeId func(ctx context.Context, employeeId string) ([]interface{ GetPermissions() string }, error)
}

type AuthzMiddleware struct {
	deps AuthzDeps
	// 路由 -> 所需权限点（数字字典）
	need map[string]int
}

func NewAuthzMiddleware(deps AuthzDeps) *AuthzMiddleware {
	// 细化的权限路由映射（所有查询接口默认放行，仅限制写接口）
	need := map[string]int{
		// task
		"POST /api/v1/task/create":  PermTaskCreate,
		"PUT /api/v1/task/update":   PermTaskUpdate,
		"POST /api/v1/task/delete":  PermTaskDelete,
		"POST /api/v1/task/approve": PermTaskApprove,
		// tasknode
		"POST /api/v1/tasknode/create": PermTaskNodeCreate,
		"PUT /api/v1/tasknode/update":  PermTaskNodeUpdate,
		"POST /api/v1/tasknode/delete": PermTaskNodeDelete,
		// handover
		"POST /api/v1/handover/create":  PermHandoverCreate,
		"POST /api/v1/handover/approve": PermHandoverApprove,
		"POST /api/v1/handover/reject":  PermHandoverReject,
		// company
		"POST /api/v1/company/create": PermCompanyCreate,
		"PUT /api/v1/company/update":  PermCompanyUpdate,
		"POST /api/v1/company/delete": PermCompanyDelete,
		// department
		"POST /api/v1/department/create": PermDepartmentCreate,
		"PUT /api/v1/department/update":  PermDepartmentUpdate,
		"POST /api/v1/department/delete": PermDepartmentDelete,
		// position
		"POST /api/v1/position/create": PermPositionCreate,
		"PUT /api/v1/position/update":  PermPositionUpdate,
		"POST /api/v1/position/delete": PermPositionDelete,
		// role
		"POST /api/v1/role/create": PermRoleCreate,
		"PUT /api/v1/role/update":  PermRoleUpdate,
		"POST /api/v1/role/delete": PermRoleDelete,
		"POST /api/v1/role/assign": PermRoleAssign,
		"POST /api/v1/role/revoke": PermRoleRevoke,
		// employee
		"POST /api/v1/employee/create": PermEmployeeCreate,
		"PUT /api/v1/employee/update":  PermEmployeeUpdate,
		"POST /api/v1/employee/delete": PermEmployeeDelete,
		"POST /api/v1/employee/leave":  PermEmployeeLeave,
		// notification
		"POST /api/v1/notification/create": PermNotificationCreate,
		"POST /api/v1/notification/delete": PermNotificationDelete,
	}
	return &AuthzMiddleware{deps: deps, need: need}
}

func (m *AuthzMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 放行无需鉴权的请求
		if r.Method == http.MethodOptions {
			next(w, r)
			return
		}
		path := r.URL.Path
		if path == "/api/v1/auth/login" || path == "/api/v1/auth/register" || path == "/api/v1/auth/logout" {
			next(w, r)
			return
		}

		key := r.Method + " " + path
		needPerm := m.need[key]
		if needPerm == 0 { // 未配置则默认放行（可按需收紧）
			next(w, r)
			return
		}

		// 从上下文取 userId（JWT 已填充），不同项目键名可能不同，尽可能兼容
		ctx := r.Context()
		var userId string
		if v := ctx.Value("userId"); v != nil {
			if s, ok := v.(string); ok {
				userId = s
			}
		}
		if userId == "" {
			if v := ctx.Value("UserId"); v != nil {
				if s, ok := v.(string); ok {
					userId = s
				}
			}
		}
		if userId == "" {
			http.Error(w, "Forbidden: no user", http.StatusForbidden)
			return
		}

		// userId -> employeeId
		emp, err := m.deps.FindEmployeeByUserID(ctx, userId)
		if err != nil || emp == nil {
			logx.Errorf("AuthZ: find employee by userId failed: %v", err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		// 优先使用内部主键 Id（与 employee.position_id 关联，通过职位获得角色），业务工号为备选
		employeeId := emp.GetId()
		if employeeId == "" {
			employeeId = emp.GetEmployeeId()
		}

		// 查询角色集合（通过职位->角色）
		roles, err := m.deps.ListRolesByEmployeeId(ctx, employeeId)
		if err != nil {
			logx.Errorf("AuthZ: list roles failed employee=%s user=%s: %v", employeeId, userId, err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if len(roles) == 0 {
			logx.Infof("AuthZ: no roles bound for employee=%s user=%s, needPerm=%d path=%s (员工可能没有职位或职位没有角色)", employeeId, userId, needPerm, key)
			http.Error(w, "Forbidden: no roles assigned", http.StatusForbidden)
			return
		}
		perms := make([]string, 0, 8)
		for _, r := range roles {
			p := strings.TrimSpace(r.GetPermissions())
			if p == "" {
				continue
			}
			perms = append(perms, p)
		}

		if len(perms) == 0 {
			logx.Infof("AuthZ: no permissions found for employee=%s user=%s, roles=%d but all permissions empty", employeeId, userId, len(roles))
			http.Error(w, "Forbidden: no permissions", http.StatusForbidden)
			return
		}

		if allowByDict(perms, needPerm) {
			next(w, r)
			return
		}
		logx.Infof("AuthZ: forbidden user=%s employee=%s needPerm=%d path=%s perms_raw=%v", userId, employeeId, needPerm, key, perms)
		http.Error(w, "Forbidden", http.StatusForbidden)
	}
}

// 旧方法已移除：仅支持数字权限

// 新：仅按数字字典校验（权限必须是 JSON 数组[int]）
func allowByDict(perms []string, need int) bool {
	if need == 0 {
		return true
	}

	// 收集所有权限码（数字集合）
	permCodes := make(map[int]struct{})
	for _, p := range perms {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var arrNum []int
		if json.Unmarshal([]byte(p), &arrNum) == nil {
			for _, n := range arrNum {
				permCodes[n] = struct{}{}
			}
		}
	}

	// 直接匹配
	if _, ok := permCodes[need]; ok {
		return true
	}
	return false
}
