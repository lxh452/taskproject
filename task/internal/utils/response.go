package utils

import (
	"task_Project/task/internal/types"
)

// 响应状态码枚举
const (
	// 成功状态码
	SUCCESS = 200

	// 客户端错误状态码
	BAD_REQUEST      = 400 // 请求参数错误
	UNAUTHORIZED     = 401 // 未授权
	FORBIDDEN        = 403 // 禁止访问
	NOT_FOUND        = 404 // 资源不存在
	CONFLICT         = 409 // 资源冲突
	VALIDATION_ERROR = 422 // 验证错误

	// 服务器错误状态码
	INTERNAL_ERROR = 500 // 内部服务器错误
	SERVICE_ERROR  = 502 // 服务错误
)

// 错误消息枚举
var ErrorMessages = map[int]string{
	BAD_REQUEST:      "请求参数错误",
	UNAUTHORIZED:     "用户未登录",
	FORBIDDEN:        "权限不足",
	NOT_FOUND:        "资源不存在",
	CONFLICT:         "资源冲突",
	VALIDATION_ERROR: "数据验证失败",
	INTERNAL_ERROR:   "系统内部错误",
	SERVICE_ERROR:    "服务错误",
}

// 成功消息枚举
var SuccessMessages = map[string]string{
	"login":     "登录成功",
	"logout":    "登出成功",
	"register":  "注册成功",
	"create":    "创建成功",
	"update":    "更新成功",
	"delete":    "删除成功",
	"query":     "查询成功",
	"operation": "操作成功",
}

// 业务错误消息枚举
var BusinessErrorMessages = map[string]string{
	// 用户相关错误
	"user_not_found":          "用户不存在",
	"user_already_exists":     "用户已存在",
	"user_already_in_company": "用户已加入其他公司，无法创建新公司",
	"username_exists":         "用户名已存在",
	"email_exists":            "邮箱已被注册",
	"email_not_found":         "该邮箱未注册",
	"phone_exists":            "手机号已被注册",
	"password_weak":           "密码强度不够",
	"user_disabled":           "用户已被禁用",
	"user_locked":             "用户已被锁定",
	"login_failed":            "用户名或密码错误",
	"code_error":              "验证码已过期，请重新获取",
	"login_failed_too_many":   "密码错误次数过多，账户已被锁定",
	"send_to_fast":            "发送频率过快，请稍后再试",

	// 公司相关错误
	"company_not_found":     "公司不存在",
	"company_name_exists":   "公司名称已存在",
	"company_name_required": "公司名称不能为空",

	// 部门相关错误
	"department_not_found":     "部门不存在",
	"department_name_exists":   "部门名称已存在",
	"department_name_required": "部门名称不能为空",

	// 职位相关错误
	"position_not_found":     "职位不存在",
	"position_name_exists":   "职位名称已存在",
	"position_name_required": "职位名称不能为空",

	// 员工相关错误
	"employee_not_found":       "员工不存在",
	"employee_already_exists":  "该用户已经是该公司的员工",
	"employee_id_exists":       "员工编号已存在",
	"employee_required_fields": "用户ID、公司ID、部门ID、职位ID和真实姓名不能为空",

	// 角色相关错误
	"role_not_found":     "角色不存在",
	"role_name_exists":   "角色名称已存在",
	"role_name_required": "角色名称不能为空",

	// 通用错误
	"invalid_params":          "参数无效",
	"missing_required_fields": "缺少必填字段",
	"data_not_found":          "数据不存在",
	"operation_failed":        "操作失败",
	"permission_denied":       "权限不足",

	// 任务相关错误
	"task_not_found":      "任务不存在",
	"task_node_not_found": "任务节点不存在",
	"task_log_error":      "任务日志创建失败",
	"taskProgress":        "无权限更新此任务节点的进度",
	"no_root_update":      "无权限更新此任务节点的前置节点",
}

// ResponseHelper 响应工具类
type ResponseHelper struct{}

// NewResponseHelper 创建响应工具实例
func NewResponseHelper() *ResponseHelper {
	return &ResponseHelper{}
}

// Success 成功响应
func (r *ResponseHelper) Success(data interface{}) *types.BaseResponse {
	return &types.BaseResponse{
		Code: SUCCESS,
		Msg:  "操作成功",
		Data: data,
	}
}

// SuccessWithMessage 带消息的成功响应
func (r *ResponseHelper) SuccessWithMessage(message string, data interface{}) *types.BaseResponse {
	return &types.BaseResponse{
		Code: SUCCESS,
		Msg:  message,
		Data: data,
	}
}

// SuccessWithKey 带消息键的成功响应
func (r *ResponseHelper) SuccessWithKey(key string, data interface{}) *types.BaseResponse {
	message := SuccessMessages[key]
	if message == "" {
		message = "操作成功"
	}
	return &types.BaseResponse{
		Code: SUCCESS,
		Msg:  message,
		Data: data,
	}
}

// Error 错误响应
func (r *ResponseHelper) Error(code int, message string) *types.BaseResponse {
	if message == "" {
		message = ErrorMessages[code]
		if message == "" {
			message = "操作失败"
		}
	}
	return &types.BaseResponse{
		Code: code,
		Msg:  message,
	}
}

// ErrorWithKey 带错误键的错误响应
func (r *ResponseHelper) ErrorWithKey(key string) *types.BaseResponse {
	message := BusinessErrorMessages[key]
	if message == "" {
		message = "操作失败"
	}
	return &types.BaseResponse{
		Code: BAD_REQUEST,
		Msg:  message,
	}
}

// BusinessError 业务错误响应
func (r *ResponseHelper) BusinessError(key string) *types.BaseResponse {
	message := BusinessErrorMessages[key]
	if message == "" {
		message = "业务错误"
	}
	return &types.BaseResponse{
		Code: BAD_REQUEST,
		Msg:  message,
	}
}

// ValidationError 验证错误响应
func (r *ResponseHelper) ValidationError(message string) *types.BaseResponse {
	return &types.BaseResponse{
		Code: VALIDATION_ERROR,
		Msg:  message,
	}
}

// UnauthorizedError 未授权错误响应
func (r *ResponseHelper) UnauthorizedError() *types.BaseResponse {
	return &types.BaseResponse{
		Code: UNAUTHORIZED,
		Msg:  ErrorMessages[UNAUTHORIZED],
	}
}

// NotFoundError 资源不存在错误响应
func (r *ResponseHelper) NotFoundError(key string) *types.BaseResponse {
	message := BusinessErrorMessages[key]
	if message == "" {
		message = "资源不存在"
	}
	return &types.BaseResponse{
		Code: NOT_FOUND,
		Msg:  message,
	}
}

// ForbiddenError 禁止访问错误响应
func (r *ResponseHelper) ForbiddenError(message string) *types.BaseResponse {
	if message == "" {
		message = ErrorMessages[FORBIDDEN]
	}
	return &types.BaseResponse{
		Code: FORBIDDEN,
		Msg:  message,
	}
}

// ConflictError 资源冲突错误响应
func (r *ResponseHelper) ConflictError(key string) *types.BaseResponse {
	message := BusinessErrorMessages[key]
	if message == "" {
		message = "资源冲突"
	}
	return &types.BaseResponse{
		Code: CONFLICT,
		Msg:  message,
	}
}

// InternalError 内部错误响应
func (r *ResponseHelper) InternalError(message string) *types.BaseResponse {
	if message == "" {
		message = ErrorMessages[INTERNAL_ERROR]
	}
	return &types.BaseResponse{
		Code: INTERNAL_ERROR,
		Msg:  message,
	}
}

// 全局响应工具实例
var Response = NewResponseHelper()
