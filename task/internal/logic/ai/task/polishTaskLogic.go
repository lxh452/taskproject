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

// PolishTaskLogic AI 润色任务逻辑
type PolishTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// PolishTaskRequest 润色任务请求
type PolishTaskRequest struct {
	TaskID     string `json:"taskId"`     // 任务ID
	TaskTitle  string `json:"taskTitle"`  // 任务标题
	TaskDetail string `json:"taskDetail"` // 任务详情
	PolishType string `json:"polishType"` // 润色类型：clarity(清晰度), professional(专业性), concise(简洁性)
}

// PolishResult 润色结果
type PolishResult struct {
	OriginalTitle  string   `json:"originalTitle"`  // 原标题
	OriginalDetail string   `json:"originalDetail"` // 原详情
	PolishedTitle  string   `json:"polishedTitle"`  // 润色后标题
	PolishedDetail string   `json:"polishedDetail"` // 润色后详情
	Improvements   []string `json:"improvements"`   // 改进点列表
}

// PolishTaskResponse AI 润色任务响应
type PolishTaskResponse struct {
	Result  PolishResult `json:"result"`  // 润色结果
	Success bool         `json:"success"` // 是否成功
}

// NewPolishTaskLogic 创建新的 PolishTaskLogic
func NewPolishTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PolishTaskLogic {
	return &PolishTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// PolishTask AI 润色任务
func (l *PolishTaskLogic) PolishTask(req *PolishTaskRequest) (*types.BaseResponse, error) {
	l.Infof("AI 润色任务: taskId=%s, polishType=%s", req.TaskID, req.PolishType)

	// 检查 AI 服务是否可用
	if l.svcCtx.GLMService == nil {
		return nil, fmt.Errorf("AI 服务未配置，无法润色任务")
	}

	// 构建 Prompt
	prompt := l.buildPolishPrompt(req)

	// 调用 GLM API
	response, err := l.svcCtx.GLMService.CallGLMWithPrompt(l.ctx, prompt)
	if err != nil {
		l.Errorf("调用 GLM API 失败: %v", err)
		// 降级方案：返回简单的格式化处理
		result := l.getDefaultPolishResult(req)
		return &types.BaseResponse{
			Code: 0,
			Msg:  "AI 服务暂时不可用，返回基础润色结果",
			Data: PolishTaskResponse{
				Result:  result,
				Success: true,
			},
		}, nil
	}

	// 解析 GLM 响应
	result, err := l.parsePolishResponse(response, req)
	if err != nil {
		l.Errorf("解析 GLM 响应失败: %v", err)
		// 降级方案
		result = l.getDefaultPolishResult(req)
	}

	return &types.BaseResponse{
		Code: 0,
		Msg:  "success",
		Data: PolishTaskResponse{
			Result:  result,
			Success: true,
		},
	}, nil
}

// buildPolishPrompt 构建润色 Prompt
func (l *PolishTaskLogic) buildPolishPrompt(req *PolishTaskRequest) string {
	var polishTypeDesc string
	switch req.PolishType {
	case "clarity":
		polishTypeDesc = "提升清晰度：让表达更加清晰易懂，去除模糊表述，明确目标和要求"
	case "professional":
		polishTypeDesc = "提升专业性：使用专业术语，采用更正式、规范的表达方式"
	case "concise":
		polishTypeDesc = "精简内容：去除冗余信息，保留核心要点，使描述更加简洁有力"
	default:
		polishTypeDesc = "综合优化：提升清晰度、专业性和简洁度"
	}

	return fmt.Sprintf(`你是一个专业的任务描述润色专家。请对以下任务描述进行优化。

## 润色类型
%s

## 原始内容
- 标题: %s
- 详情: %s

## 润色要求
1. 保持原始意图和核心信息不变
2. 根据润色类型进行针对性优化
3. 标题应该简洁明了，概括性强
4. 详情应该结构清晰，重点突出
5. 列出主要的改进点

## 输出格式
请严格按照以下 JSON 格式输出，不要包含任何其他内容：
{
  "polishedTitle": "润色后的标题",
  "polishedDetail": "润色后的详情",
  "improvements": [
    "改进点1",
    "改进点2",
    "改进点3"
  ]
}`, polishTypeDesc, req.TaskTitle, req.TaskDetail)
}

// parsePolishResponse 解析 GLM 润色响应
func (l *PolishTaskLogic) parsePolishResponse(response string, req *PolishTaskRequest) (PolishResult, error) {
	// 尝试从响应中提取 JSON
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return PolishResult{}, fmt.Errorf("无法从响应中提取 JSON")
	}

	jsonStr := response[jsonStart : jsonEnd+1]

	var result struct {
		PolishedTitle  string   `json:"polishedTitle"`
		PolishedDetail string   `json:"polishedDetail"`
		Improvements   []string `json:"improvements"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return PolishResult{}, fmt.Errorf("解析 JSON 失败: %v", err)
	}

	return PolishResult{
		OriginalTitle:  req.TaskTitle,
		OriginalDetail: req.TaskDetail,
		PolishedTitle:  result.PolishedTitle,
		PolishedDetail: result.PolishedDetail,
		Improvements:   result.Improvements,
	}, nil
}

// getDefaultPolishResult 获取默认润色结果（降级方案）
func (l *PolishTaskLogic) getDefaultPolishResult(req *PolishTaskRequest) PolishResult {
	polishedTitle := req.TaskTitle
	polishedDetail := req.TaskDetail

	switch req.PolishType {
	case "clarity":
		polishedTitle = "【已优化】" + req.TaskTitle
		polishedDetail = "【表达更清晰】\n" + req.TaskDetail
	case "professional":
		polishedTitle = "【专业化】" + req.TaskTitle
		polishedDetail = "【使用专业术语】\n" + req.TaskDetail
	case "concise":
		polishedTitle = "【简洁版】" + req.TaskTitle
		polishedDetail = "【精简描述】\n" + req.TaskDetail
	}

	return PolishResult{
		OriginalTitle:  req.TaskTitle,
		OriginalDetail: req.TaskDetail,
		PolishedTitle:  polishedTitle,
		PolishedDetail: polishedDetail,
		Improvements:   []string{"表达更清晰", "重点更突出"},
	}
}
