package utils

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"task_Project/task/internal/middleware"

	"github.com/mitchellh/mapstructure"
)

var Common = NewCommon()

// Common 通用工具类
type common struct{}

// NewCommon 创建通用工具实例
func NewCommon() *common {
	return &common{}
}

// GetCurrentUserID 获取当前用户ID
func (c *common) GetCurrentUserID(ctx context.Context) (string, bool) {
	return middleware.GetUserID(ctx)
}

// GetCurrentUsername 获取当前用户名
func (c *common) GetCurrentUsername(ctx context.Context) (string, bool) {
	return middleware.GetUsername(ctx)
}

// GetCurrentRealName 获取当前用户真实姓名
func (c *common) GetCurrentRealName(ctx context.Context) (string, bool) {
	return middleware.GetRealName(ctx)
}

// GetCurrentUserRole 获取当前用户角色
func (c *common) GetCurrentUserRole(ctx context.Context) (string, bool) {
	return middleware.GetRole(ctx)
}

// GetCurrentEmployeeID 获取当前员工ID
func (c *common) GetCurrentEmployeeID(ctx context.Context) (string, bool) {
	return middleware.GetEmployeeID(ctx)
}

// GetCurrentCompanyID 获取当前公司ID
func (c *common) GetCurrentCompanyID(ctx context.Context) (string, bool) {
	return middleware.GetCompanyID(ctx)
}

// GenerateID 生成ID
func (c *common) GenerateID() string {
	return time.Now().Format("20060102150405") + "0001"
}

// GenerateIDWithPrefix 生成带前缀的ID
func (c *common) GenerateIDWithPrefix(prefix string) string {
	return prefix + time.Now().Format("20060102150405") + "0001"
}

// GetCurrentTime 获取当前时间
func (c *common) GetCurrentTime() time.Time {
	return time.Now()
}

// FormatTime 格式化时间
func (c *common) FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDate 格式化日期
func (c *common) FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// IsValidTime 验证时间是否有效
func (c *common) IsValidTime(t *time.Time) bool {
	return t != nil && !t.IsZero()
}

// IsValidDate 验证日期是否有效
func (c *common) IsValidDate(t *time.Time) bool {
	return t != nil && !t.IsZero()
}

// GetTimeString 获取时间字符串
func (c *common) GetTimeString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// GetDateString 获取日期字符串
func (c *common) GetDateString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// ParseTime 解析时间字符串
func (c *common) ParseTime(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", timeStr)
}

// ParseDate 解析日期字符串
func (c *common) ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// IsEmptyString 检查字符串是否为空
func (c *common) IsEmptyString(str string) bool {
	return len(str) == 0
}

// IsNotEmptyString 检查字符串是否不为空
func (c *common) IsNotEmptyString(str string) bool {
	return len(str) > 0
}

// TrimString 去除字符串首尾空格
func (c *common) TrimString(str string) string {
	return strings.TrimSpace(str)
}

// DefaultString 获取默认字符串
func (c *common) DefaultString(str, defaultStr string) string {
	if c.IsEmptyString(str) {
		return defaultStr
	}
	return str
}

// DefaultInt 获取默认整数值
func (c *common) DefaultInt(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

// DefaultInt64 获取默认int64值
func (c *common) DefaultInt64(value, defaultValue int64) int64 {
	if value == 0 {
		return defaultValue
	}
	return value
}

// MaxInt 获取两个整数的最大值
func (c *common) MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinInt 获取两个整数的最小值
func (c *common) MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt64 获取两个int64的最大值
func (c *common) MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// MinInt64 获取两个int64的最小值
func (c *common) MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// 将string类型转换为sql.
func (c *common) ToSqlNullString(str string) sql.NullString {
	sqlNullString := sql.NullString{}
	if str == "" {
		sqlNullString.Valid = false
		return sqlNullString
	}
	sqlNullString.String = str
	sqlNullString.Valid = true
	return sqlNullString
}

func (c *common) ToSqlNullTime(timeStr string) sql.NullTime {
	var nullTime sql.NullTime

	timeFormat := "2006-01-02 15:04:05"

	parsedTime, err := time.Parse(timeFormat, timeStr)
	if err != nil {
		nullTime.Valid = false
		return nullTime
	}
	if parsedTime.IsZero() {
		nullTime.Valid = false
	} else {
		nullTime.Time = parsedTime
		nullTime.Valid = true
	}

	return nullTime
}

// 将int类型转换为sql.
func (c *common) ToSqlNullFloat64(num int) sql.NullFloat64 {
	sqlNullFloat64 := sql.NullFloat64{}
	sqlNullFloat64.Float64 = float64(num)
	sqlNullFloat64.Valid = true
	return sqlNullFloat64
}

func (c *common) MapToStructWithMapstructure(data map[string]interface{}, result interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   result,
		TagName:  "db", // 使用db标签匹配数据库模型
		// 可以添加更多配置
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			// 字符串转时间
			mapstructure.StringToTimeHookFunc(time.RFC3339),
			mapstructure.StringToTimeHookFunc("2006-01-02 15:04:05"),
			mapstructure.StringToTimeHookFunc("2006-01-02"),
			// 处理sql.NullString
			c.stringToSqlNullStringHook(),
			// 处理sql.NullTime
			c.stringToSqlNullTimeHook(),
			// 处理int到int64的转换
			c.stringToInt64Hook(),
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(data)
}

// stringToSqlNullStringHook 字符串转sql.NullString的钩子函数
func (c *common) stringToSqlNullStringHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() == reflect.String && t == reflect.TypeOf(sql.NullString{}) {
			str := data.(string)
			if str == "" {
				return sql.NullString{Valid: false}, nil
			}
			return sql.NullString{String: str, Valid: true}, nil
		}
		return data, nil
	}
}

// stringToSqlNullTimeHook 字符串转sql.NullTime的钩子函数
func (c *common) stringToSqlNullTimeHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() == reflect.String && t == reflect.TypeOf(sql.NullTime{}) {
			str := data.(string)
			if str == "" {
				return sql.NullTime{Valid: false}, nil
			}
			// 尝试解析时间
			timeFormats := []string{
				"2006-01-02 15:04:05",
				"2006-01-02",
				time.RFC3339,
			}
			for _, format := range timeFormats {
				if parsedTime, err := time.Parse(format, str); err == nil {
					return sql.NullTime{Time: parsedTime, Valid: true}, nil
				}
			}
			return sql.NullTime{Valid: false}, nil
		}
		return data, nil
	}
}

// stringToInt64Hook 字符串转int64的钩子函数
func (c *common) stringToInt64Hook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() == reflect.String && t.Kind() == reflect.Int64 {
			str := data.(string)
			if str == "" {
				return int64(0), nil
			}
			// 使用strconv.ParseInt解析
			if val, err := strconv.ParseInt(str, 10, 64); err == nil {
				return val, nil
			}
		}
		return data, nil
	}
}

func (c *common) GenId(name string) string {
	bytes := make([]byte, 5)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%v_%X", name, bytes)
}

// IsValidEmail 验证邮箱格式
func (c *common) IsValidEmail(email string) bool {
	return Validator.IsValidEmail(email)
}

// IsValidPassword 验证密码强度
func (c *common) IsValidPassword(password string) bool {
	return Validator.IsValidPassword(password)
}
