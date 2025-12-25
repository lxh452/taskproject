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
	"user_not_bindemployee":   "用户未绑定员工信息",

	// 公司相关错误
	"company_not_found":     "公司不存在",
	"company_name_exists":   "公司名称已存在",
	"company_name_required": "公司名称不能为空",
	"company_has_employees": "公司还有员工，无法删除",

	// 部门相关错误
	"department_not_found":     "部门不存在",
	"department_name_exists":   "部门名称已存在",
	"department_name_required": "部门名称不能为空",
	"department_has_employees": "部门还有员工，无法删除",

	// 职位相关错误
	"position_not_found":     "职位不存在",
	"position_name_exists":   "职位名称已存在",
	"position_name_required": "职位名称不能为空",
	"position_has_employees": "职位还有员工，无法删除",

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
	"format_error":            "格式错误",

	// 任务相关错误
	"task_not_found":           "任务不存在",
	"task_id_required":         "任务ID不能为空",
	"task_title_required":      "任务标题不能为空",
	"task_deadline_required":   "任务截止时间不能为空",
	"task_deadline_format":     "任务截止时间格式错误",
	"task_already_completed":   "任务已经完成",
	"task_has_active_nodes":    "任务有进行中的节点，无法删除",
	"task_nodes_not_completed": "所有任务节点完成后才能标记任务完成",
	"task_update_denied":       "无权限更新此任务，只有任务创建者或负责人可以修改",
	"task_delete_denied":       "无权限删除此任务",
	"task_complete_denied":     "无权限完成此任务",
	"task_completed_no_delete": "已完成的任务无法删除",
	"task_log_error":           "任务日志创建失败",
	"task_nodes_fetch_failed":  "获取任务节点失败",

	// 任务节点相关错误
	"task_node_not_found":           "任务节点不存在",
	"task_node_id_required":         "任务节点ID不能为空",
	"task_node_view_denied":         "无权限查看此任务节点",
	"task_node_delete_denied":       "无权限删除此任务节点",
	"task_node_completed_no_delete": "已完成的任务节点无法删除",
	"task_node_has_dependents":      "有其他任务节点依赖此节点，无法删除",
	"task_node_progress_denied":     "无权限更新此任务节点的进度",
	"task_node_prereq_denied":       "无权限更新此任务节点的前置节点",
	"taskProgress":                  "无权限更新此任务节点的进度",
	"no_root_update":                "无权限更新此任务节点的前置节点",
	"progress_range_error":          "进度值必须在0-100之间",
	"auto_dispatch_denied":          "无权限执行自动派发",

	// 通知相关错误
	"notification_not_found":        "通知不存在",
	"notification_id_required":      "通知ID不能为空",
	"notification_title_required":   "通知标题不能为空",
	"notification_content_required": "通知内容不能为空",
	"notification_view_denied":      "无权查询其他员工的通知",
	"notification_update_denied":    "无权操作其他员工的通知",

	// 交接相关错误
	"handover_not_found":      "交接记录不存在",
	"handover_id_required":    "交接ID不能为空",
	"handover_status_invalid": "只有待接收人确认的交接才能进行拒绝操作",
	"handover_reject_denied":  "只有交接接收人才能拒绝",

	// 认证相关错误
	"email_format_invalid":    "邮箱格式不正确",
	"password_format_invalid": "密码必须8-32位，包含数字、大小写字母和特殊字符",
	"login_locked":            "登录失败次数过多，账户已锁定10分钟",

	// 员工相关错误（补充）
	"employee_has_tasks":       "员工还有未完成的任务，无法删除",
	"employee_not_in_company":  "您尚未加入任何公司",
	"employee_already_left":    "员工已离职",
	"founder_cannot_leave":     "不能给公司创始人递交离职申请",
	"permission_verify_failed": "权限验证失败",
	"invalid_leave_type":       "无效的离职类型",
	"only_admin_can_view":      "只有公司创始人、人事部门或管理人员可以查看",
	"only_admin_can_approve":   "只有公司创始人、人事部门或管理人员可以审批",
	"only_admin_can_generate":  "只有公司创始人、人事部门或管理人员可以生成邀请码",

	// 申请相关错误
	"already_in_company":       "您已经加入了公司，无法再次申请",
	"pending_application":      "您已有待审批的申请，请等待审批结果",
	"invite_company_not_found": "邀请码对应的公司不存在",
	"company_disabled":         "该公司已停用，无法申请加入",
	"application_not_found":    "申请不存在",
	"application_processed":    "该申请已处理",
	"no_permission_approve":    "您无权审批此申请",

	// 审批相关错误
	"approval_id_required":       "审批ID不能为空",
	"approval_result_invalid":    "审批结果无效，1-同意，2-拒绝",
	"approval_not_found":         "审批记录不存在",
	"approval_type_invalid":      "该审批记录不是任务节点完成审批",
	"approval_already_done":      "该审批记录已处理，无法重复审批",
	"approval_permission_denied": "无权限审批，只有项目负责人可以审批",
	"approval_missing_node_id":   "审批记录缺少任务节点ID",

	// 兼容旧的英文key
	"The task deadline cannot be empty":                         "任务截止时间不能为空",
	"Task deadline format is incorrect":                         "任务截止时间格式错误",
	"Task title cannot be empty":                                "任务标题不能为空",
	"The company still has employees and cannot be deleted.":    "公司还有员工，无法删除",
	"The department still has employees and cannot be deleted.": "部门还有员工，无法删除",
	"Format error": "格式错误",
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
