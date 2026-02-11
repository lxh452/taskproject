// Code scaffolded by VibeCraft AI
// VibeCraft AI Logic

package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// GenerateSubtasksLogic AI 生成子任务逻辑
type GenerateSubtasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// GenerateSubtasksRequest 生成子任务请求
type GenerateSubtasksRequest struct {
	TaskID     string                 `json:"taskId"`           // 任务ID
	TaskTitle  string                 `json:"taskTitle"`        // 任务标题
	TaskDetail string                 `json:"taskDetail"`       // 任务详情
	Context    map[string]interface{} `json:"context,optional"` // 上下文信息
}

// Subtask 子任务
type Subtask struct {
	SubtaskID      string `json:"subtaskId"`      // 子任务ID
	Title          string `json:"title"`          // 子任务标题（兼容旧字段）
	Description    string `json:"description"`    // 子任务描述（兼容旧字段）
	NodeType       int    `json:"nodeType"`       // 节点类型 1=里程碑 2=开发任务 3=测试任务 4=文档任务 5=评审任务
	NodeName       string `json:"nodeName"`       // 节点名称
	NodeDetail     string `json:"nodeDetail"`     // 节点详细描述
	EstimatedHours int    `json:"estimatedHours"` // 预估工时
	Priority       int    `json:"priority"`       // 优先级
}

// GenerateSubtasksResponse AI 生成子任务响应
type GenerateSubtasksResponse struct {
	Subtasks []Subtask `json:"subtasks"` // 子任务列表
	Total    int       `json:"total"`    // 总数
	Success  bool      `json:"success"`  // 是否成功
}

// NewGenerateSubtasksLogic 创建新的 GenerateSubtasksLogic
func NewGenerateSubtasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateSubtasksLogic {
	return &GenerateSubtasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GenerateSubtasks AI 生成子任务
func (l *GenerateSubtasksLogic) GenerateSubtasks(req *GenerateSubtasksRequest) (*types.BaseResponse, error) {
	l.Infof("AI 生成子任务: taskId=%s, taskTitle=%s", req.TaskID, req.TaskTitle)

	// 检查 AI 服务是否可用
	if l.svcCtx.GLMService == nil {
		return nil, fmt.Errorf("AI 服务未配置，无法生成子任务")
	}

	// 构建 Prompt
	prompt := l.buildSubtaskPrompt(req)

	// 调用 GLM API
	response, err := l.svcCtx.GLMService.CallGLMWithPrompt(l.ctx, prompt)
	if err != nil {
		l.Errorf("调用 GLM API 失败: %v", err)
		// 降级方案：返回基于规则的默认子任务
		subtasks := l.getDefaultSubtasks(req)
		return &types.BaseResponse{
			Code: 0,
			Msg:  "AI 服务暂时不可用，返回默认子任务",
			Data: GenerateSubtasksResponse{
				Subtasks: subtasks,
				Total:    len(subtasks),
				Success:  true,
			},
		}, nil
	}

	// 解析 GLM 响应
	subtasks, err := l.parseSubtaskResponse(response)
	if err != nil {
		l.Errorf("解析 GLM 响应失败: %v", err)
		// 降级方案：返回基于规则的默认子任务
		subtasks = l.getDefaultSubtasks(req)
	}

	return &types.BaseResponse{
		Code: 0,
		Msg:  "success",
		Data: GenerateSubtasksResponse{
			Subtasks: subtasks,
			Total:    len(subtasks),
			Success:  true,
		},
	}, nil
}

// buildSubtaskPrompt 构建生成子任务的 Prompt
func (l *GenerateSubtasksLogic) buildSubtaskPrompt(req *GenerateSubtasksRequest) string {
	contextStr := ""
	if req.Context != nil {
		contextJSON, _ := json.Marshal(req.Context)
		contextStr = string(contextJSON)
	}

	return fmt.Sprintf(`你是一个专业的任务拆解专家。请根据以下主任务信息，智能生成合理的子任务列表。

## 主任务信息
- 任务标题: %s
- 任务详情: %s
- 附加信息: %s

## 生成要求
1. 将主任务拆解为 3-8 个可执行的子任务
2. 每个子任务应包含：标题、详细描述、预估工时(小时)、优先级(1-5,1最高)
3. 子任务之间应该有合理的依赖关系和执行顺序
4. 预估工时应该合理，总工时应与任务复杂度匹配
5. 优先级应根据任务的关键程度设定

## 输出格式
请严格按照以下 JSON 格式输出，不要包含任何其他内容：
{
  "subtasks": [
    {
      "title": "子任务标题",
      "description": "子任务详细描述",
      "estimatedHours": 预估工时数字,
      "priority": 优先级数字
    }
  ]
}`, req.TaskTitle, req.TaskDetail, contextStr)
}

// parseSubtaskResponse 解析 GLM 返回的子任务
func (l *GenerateSubtasksLogic) parseSubtaskResponse(response string) ([]Subtask, error) {
	// 尝试从响应中提取 JSON
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("无法从响应中提取 JSON")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	var result struct {
		Subtasks []struct {
			Title          string `json:"title"`
			Description    string `json:"description"`
			EstimatedHours int    `json:"estimatedHours"`
			Priority       int    `json:"priority"`
		} `json:"subtasks"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %v", err)
	}

	// 转换为 Subtask 结构
	subtasks := make([]Subtask, 0, len(result.Subtasks))
	for i, st := range result.Subtasks {
		subtasks = append(subtasks, Subtask{
			SubtaskID:      fmt.Sprintf("sub_%03d", i+1),
			Title:          st.Title,
			Description:    st.Description,
			EstimatedHours: st.EstimatedHours,
			Priority:       st.Priority,
		})
	}

	return subtasks, nil
}

// getDefaultSubtasks 获取默认子任务（降级方案）
func (l *GenerateSubtasksLogic) getDefaultSubtasks(req *GenerateSubtasksRequest) []Subtask {
	return []Subtask{
		{
			SubtaskID:      "sub_001",
			Title:          "需求分析与理解",
			Description:    fmt.Sprintf("深入理解任务'%s'的具体需求和目标", req.TaskTitle),
			EstimatedHours: 4,
			Priority:       1,
		},
		{
			SubtaskID:      "sub_002",
			Title:          "方案设计与规划",
			Description:    "制定详细的执行方案和步骤",
			EstimatedHours: 6,
			Priority:       1,
		},
		{
			SubtaskID:      "sub_003",
			Title:          "核心内容实现",
			Description:    "完成主要工作内容",
			EstimatedHours: 12,
			Priority:       2,
		},
		{
			SubtaskID:      "sub_004",
			Title:          "质量检查与验证",
			Description:    "验证完成质量，确保符合要求",
			EstimatedHours: 4,
			Priority:       2,
		},
		{
			SubtaskID:      "sub_005",
			Title:          "成果整理与提交",
			Description:    "整理工作成果并提交",
			EstimatedHours: 2,
			Priority:       3,
		},
	}
}
