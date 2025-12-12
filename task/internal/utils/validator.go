package utils

import (
	"regexp"
	"strings"
)

// Validator 验证工具类
type validator struct{}

var Validator = NewValidator()

// NewValidator 创建验证工具实例
func NewValidator() *validator {
	return &validator{}
}

// IsEmpty 检查字符串是否为空
func (v *validator) IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// IsNotEmpty 检查字符串是否不为空
func (v *validator) IsNotEmpty(str string) bool {
	return !v.IsEmpty(str)
}

// IsValidEmail 验证邮箱格式
func (v *validator) IsValidEmail(email string) bool {
	if v.IsEmpty(email) {
		return false
	}
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// IsValidPhone 验证手机号格式
func (v *validator) IsValidPhone(phone string) bool {
	if v.IsEmpty(phone) {
		return false
	}
	// 中国大陆手机号格式
	pattern := `^1[3-9]\d{9}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// IsValidPassword 验证密码强度
func (v *validator) IsValidPassword(password string) bool {
	if v.IsEmpty(password) {
		return false
	}
	// 密码长度至少8位，最多32位
	if len(password) < 8 || len(password) > 32 {
		return false
	}
	// 必须包含数字
	hasDigit, _ := regexp.MatchString(`\d`, password)
	if !hasDigit {
		return false
	}
	// 必须包含字母
	hasLetter, _ := regexp.MatchString(`[a-zA-Z]`, password)
	if !hasLetter {
		return false
	}
	// 必须包含特殊字符
	hasSpecial, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`, password)
	if !hasSpecial {
		return false
	}
	return true
}

// IsValidUsername 验证用户名格式
func (v *validator) IsValidUsername(username string) bool {
	if v.IsEmpty(username) {
		return false
	}
	// 用户名长度3-20位，只能包含字母、数字、下划线
	pattern := `^[a-zA-Z0-9_]{3,20}$`
	matched, _ := regexp.MatchString(pattern, username)
	return matched
}

// IsValidID 验证ID格式
func (v *validator) IsValidID(id string) bool {
	if v.IsEmpty(id) {
		return false
	}
	// ID长度至少1位
	return len(id) >= 1
}

// ValidateRequired 验证必填字段
func (v *validator) ValidateRequired(fields map[string]string) []string {
	var errors []string
	for field, value := range fields {
		if v.IsEmpty(value) {
			errors = append(errors, field+"不能为空")
		}
	}
	return errors
}

// ValidateEmail 验证邮箱
func (v *validator) ValidateEmail(email string) string {
	if v.IsEmpty(email) {
		return ""
	}
	if !v.IsValidEmail(email) {
		return "邮箱格式不正确"
	}
	return ""
}

// ValidatePhone 验证手机号
func (v *validator) ValidatePhone(phone string) string {
	if v.IsEmpty(phone) {
		return ""
	}
	if !v.IsValidPhone(phone) {
		return "手机号格式不正确"
	}
	return ""
}

// ValidatePassword 验证密码
func (v *validator) ValidatePassword(password string) string {
	if v.IsEmpty(password) {
		return "密码不能为空"
	}
	if len(password) < 8 {
		return "密码长度不能少于8位"
	}
	if len(password) > 32 {
		return "密码长度不能超过32位"
	}
	hasDigit, _ := regexp.MatchString(`\d`, password)
	if !hasDigit {
		return "密码必须包含数字"
	}
	hasLetter, _ := regexp.MatchString(`[a-zA-Z]`, password)
	if !hasLetter {
		return "密码必须包含字母"
	}
	hasSpecial, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`, password)
	if !hasSpecial {
		return "密码必须包含特殊字符(!@#$%^&*等)"
	}
	return ""
}

// ValidateUsername 验证用户名
func (v *validator) ValidateUsername(username string) string {
	if v.IsEmpty(username) {
		return "用户名不能为空"
	}
	if !v.IsValidUsername(username) {
		return "用户名格式不正确，只能包含字母、数字、下划线，长度3-20位"
	}
	return ""
}

// ValidateID 验证ID
func (v *validator) ValidateID(id string, fieldName string) string {
	if v.IsEmpty(id) {
		return fieldName + "不能为空"
	}
	if !v.IsValidID(id) {
		return fieldName + "格式不正确"
	}
	return ""
}

// ValidateLength 验证字符串长度
func (v *validator) ValidateLength(str string, min, max int, fieldName string) string {
	if v.IsEmpty(str) {
		return fieldName + "不能为空"
	}
	length := len(str)
	if length < min {
		return fieldName + "长度不能少于" + string(rune(min)) + "位"
	}
	if length > max {
		return fieldName + "长度不能超过" + string(rune(max)) + "位"
	}
	return ""
}

// ValidateRange 验证数值范围
func (v *validator) ValidateRange(value int, min, max int, fieldName string) string {
	if value < min {
		return fieldName + "不能小于" + string(rune(min))
	}
	if value > max {
		return fieldName + "不能大于" + string(rune(max))
	}
	return ""
}

// ValidatePositive 验证正数
func (v *validator) ValidatePositive(value int, fieldName string) string {
	if value <= 0 {
		return fieldName + "必须大于0"
	}
	return ""
}

// ValidateNonNegative 验证非负数
func (v *validator) ValidateNonNegative(value int, fieldName string) string {
	if value < 0 {
		return fieldName + "不能为负数"
	}
	return ""
}

// ValidateStatus 验证状态值
func (v *validator) ValidateStatus(status int, validStatuses []int, fieldName string) string {
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return ""
		}
	}
	return fieldName + "状态值无效"
}

// ValidatePageParams 验证分页参数
func (v *validator) ValidatePageParams(page, pageSize int) (int, int, []string) {
	var errors []string

	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	if pageSize > 1000 {
		pageSize = 1000
		errors = append(errors, "每页大小不能超过100")
	}

	return page, pageSize, errors
}
