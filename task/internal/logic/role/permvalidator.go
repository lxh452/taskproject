package role

import (
	"encoding/json"
	"fmt"
	"strings"

	mw "task_Project/task/internal/middleware"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"
)

// ValidatePermissions 验证权限字符串是否都在字典中
func ValidatePermissions(permStr string) (*types.BaseResponse, error) {
	if permStr == "" {
		return nil, nil // 允许空权限
	}
	permStr = strings.TrimSpace(permStr)

	// 尝试解析为 JSON 数组（数字）
	var arrNum []int
	if err := json.Unmarshal([]byte(permStr), &arrNum); err == nil {
		validPerms := mw.GetValidPermCodes()
		for _, p := range arrNum {
			if _, ok := validPerms[p]; !ok {
				return utils.Response.ValidationError(fmt.Sprintf("invalid permission code: %d", p)), nil
			}
		}
		return nil, nil
	}

	// 仅允许 JSON 数组[int]，其它格式一律不通过
	return utils.Response.ValidationError("permissions must be JSON array of integers"), nil
}
