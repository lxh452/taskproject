package utils

import (
	"regexp"
	"strings"
)

// SQL注入检测模式
var sqlInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(\b(SELECT|INSERT|UPDATE|DELETE|DROP|UNION|ALTER|CREATE|TRUNCATE)\b.*\b(FROM|INTO|TABLE|DATABASE)\b)`),
	regexp.MustCompile(`(?i)(\bOR\b\s+\d+\s*=\s*\d+)`),
	regexp.MustCompile(`(?i)(\bAND\b\s+\d+\s*=\s*\d+)`),
	regexp.MustCompile(`(?i)(--\s*$|/\*|\*/)`),
	regexp.MustCompile(`(?i)(\bEXEC\b|\bEXECUTE\b)`),
	regexp.MustCompile(`(?i)(\bXP_\w+)`),
	regexp.MustCompile(`(?i)(;\s*(SELECT|INSERT|UPDATE|DELETE|DROP))`),
	regexp.MustCompile(`(?i)('\s*(OR|AND)\s*')`),
	regexp.MustCompile(`(?i)(SLEEP\s*\(\s*\d+\s*\))`),
	regexp.MustCompile(`(?i)(BENCHMARK\s*\()`),
	regexp.MustCompile(`(?i)(LOAD_FILE\s*\()`),
	regexp.MustCompile(`(?i)(INTO\s+(OUT|DUMP)FILE)`),
}

// 危险字符
var dangerousChars = []string{"'", "\"", ";", "--", "/*", "*/", "@@", "@"}

// DetectSQLInjection 检测字符串是否包含SQL注入特征
func DetectSQLInjection(input string) (bool, string) {
	if input == "" {
		return false, ""
	}
	// 检查正则模式
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(input) {
			return true, pattern.String()
		}
	}
	// 检查危险字符组合
	lower := strings.ToLower(input)
	if strings.Contains(lower, "' or ") || strings.Contains(lower, "' and ") ||
		strings.Contains(lower, "1=1") || strings.Contains(lower, "1'='1") {
		return true, "dangerous_pattern"
	}
	return false, ""
}

// SanitizeInput 清理输入，移除潜在危险字符
func SanitizeInput(input string) string {
	result := input
	for _, char := range dangerousChars {
		result = strings.ReplaceAll(result, char, "")
	}
	return result
}

// IsSafeIdentifier 检查是否是安全的标识符（表名、列名等）
func IsSafeIdentifier(identifier string) bool {
	if identifier == "" {
		return false
	}
	// 只允许字母、数字、下划线
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, identifier)
	return matched
}
