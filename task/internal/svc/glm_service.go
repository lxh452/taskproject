package svc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// GLMService 智谱AI GLM服务
type GLMService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// GLMConfig GLM配置
type GLMConfig struct {
	APIKey  string
	BaseURL string
}

// NewGLMService 创建GLM服务
func NewGLMService(config GLMConfig) *GLMService {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	}
	return &GLMService{
		apiKey:  config.APIKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GLMRequest GLM请求结构
type GLMRequest struct {
	Model    string       `json:"model"`
	Messages []GLMMessage `json:"messages"`
}

// GLMMessage GLM消息
type GLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GLMResponse GLM响应结构
type GLMResponse struct {
	ID      string `json:"id"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// EmployeeCandidate 员工候选人信息
type EmployeeCandidate struct {
	EmployeeID     string   `json:"employeeId"`
	Name           string   `json:"name"`
	Department     string   `json:"department"`
	Position       string   `json:"position"`
	Skills         []string `json:"skills"`
	TenureMonths   int      `json:"tenureMonths"`
	ActiveTasks    int      `json:"activeTasks"`
	CompletedTasks int      `json:"completedTasks"`
	AvgCompletion  float64  `json:"avgCompletionRate"` // 平均完成率
}

// TaskNodeInfo 任务节点信息
type TaskNodeInfo struct {
	NodeID       string   `json:"nodeId"`
	NodeName     string   `json:"nodeName"`
	NodeDetail   string   `json:"nodeDetail"`
	Priority     int      `json:"priority"`
	Deadline     string   `json:"deadline"`
	RequiredDays int      `json:"requiredDays"`
	TaskTitle    string   `json:"taskTitle"`
	Skills       []string `json:"requiredSkills,omitempty"`
}

// RecommendedEmployee 推荐的员工
type RecommendedEmployee struct {
	EmployeeID string  `json:"employeeId"`
	Name       string  `json:"name"`
	Score      float64 `json:"score"`
	Reason     string  `json:"reason"`
	Rank       int     `json:"rank"`
}

// DispatchRecommendation 派发推荐结果
type DispatchRecommendation struct {
	TaskNodeID   string                `json:"taskNodeId"`
	TaskNodeName string                `json:"taskNodeName"`
	Candidates   []RecommendedEmployee `json:"candidates"`
	AIAnalysis   string                `json:"aiAnalysis"`
}

// GetDispatchRecommendation 获取派发推荐
func (s *GLMService) GetDispatchRecommendation(ctx context.Context, taskNode TaskNodeInfo, candidates []EmployeeCandidate) (*DispatchRecommendation, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("GLM API Key未配置")
	}

	// 构建提示词
	prompt := s.buildDispatchPrompt(taskNode, candidates)

	// 调用GLM API
	response, err := s.callGLM(ctx, prompt)
	if err != nil {
		logx.Errorf("调用GLM API失败: %v", err)
		return nil, err
	}

	// 解析响应
	recommendation, err := s.parseDispatchResponse(response, taskNode, candidates)
	if err != nil {
		logx.Errorf("解析GLM响应失败: %v", err)
		return nil, err
	}

	return recommendation, nil
}

// buildDispatchPrompt 构建派发提示词
func (s *GLMService) buildDispatchPrompt(taskNode TaskNodeInfo, candidates []EmployeeCandidate) string {
	// 构建候选人信息
	candidatesJSON, _ := json.MarshalIndent(candidates, "", "  ")
	taskJSON, _ := json.MarshalIndent(taskNode, "", "  ")

	prompt := fmt.Sprintf(`你是一个智能任务派发助手，需要根据任务需求和员工能力，推荐最适合执行任务的员工。

## 任务信息
%s

## 候选员工列表
%s

## 评估标准
1. 技能匹配度：员工技能与任务需求的匹配程度
2. 工作负载：当前活跃任务数量，避免过度分配
3. 历史表现：已完成任务数量和平均完成率
4. 经验资历：任职时长
5. 任务优先级：高优先级任务需要更有经验的员工

## 输出要求
请从候选员工中选出最适合的5名员工（如果候选人不足5人则全部推荐），按推荐度从高到低排序。

请严格按照以下JSON格式输出，不要包含其他内容：
{
  "recommendations": [
    {
      "employeeId": "员工ID",
      "name": "员工姓名",
      "score": 评分(0-100),
      "reason": "推荐理由（简洁说明为什么推荐此人）"
    }
  ],
  "analysis": "整体分析说明（简要说明推荐逻辑和考虑因素）"
}`, string(taskJSON), string(candidatesJSON))

	return prompt
}

// callGLM 调用GLM API
func (s *GLMService) callGLM(ctx context.Context, prompt string) (string, error) {
	reqBody := GLMRequest{
		Model: "glm-4-flash", // 使用更快的模型
		Messages: []GLMMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误: %d, %s", resp.StatusCode, string(body))
	}

	var glmResp GLMResponse
	if err := json.Unmarshal(body, &glmResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(glmResp.Choices) == 0 {
		return "", fmt.Errorf("API返回空响应")
	}

	return glmResp.Choices[0].Message.Content, nil
}

// parseDispatchResponse 解析派发响应
func (s *GLMService) parseDispatchResponse(response string, taskNode TaskNodeInfo, candidates []EmployeeCandidate) (*DispatchRecommendation, error) {
	// 尝试解析JSON响应
	var result struct {
		Recommendations []struct {
			EmployeeID string  `json:"employeeId"`
			Name       string  `json:"name"`
			Score      float64 `json:"score"`
			Reason     string  `json:"reason"`
		} `json:"recommendations"`
		Analysis string `json:"analysis"`
	}

	// 尝试从响应中提取JSON
	jsonStart := -1
	jsonEnd := -1
	for i, c := range response {
		if c == '{' && jsonStart == -1 {
			jsonStart = i
		}
		if c == '}' {
			jsonEnd = i + 1
		}
	}

	if jsonStart >= 0 && jsonEnd > jsonStart {
		jsonStr := response[jsonStart:jsonEnd]
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			logx.Errorf("解析JSON失败: %v, 原始响应: %s", err, response)
			// 如果解析失败，返回基于规则的默认推荐
			return s.getDefaultRecommendation(taskNode, candidates), nil
		}
	} else {
		return s.getDefaultRecommendation(taskNode, candidates), nil
	}

	// 构建推荐结果
	recommendation := &DispatchRecommendation{
		TaskNodeID:   taskNode.NodeID,
		TaskNodeName: taskNode.NodeName,
		AIAnalysis:   result.Analysis,
	}

	for i, rec := range result.Recommendations {
		if i >= 5 {
			break
		}
		recommendation.Candidates = append(recommendation.Candidates, RecommendedEmployee{
			EmployeeID: rec.EmployeeID,
			Name:       rec.Name,
			Score:      rec.Score,
			Reason:     rec.Reason,
			Rank:       i + 1,
		})
	}

	return recommendation, nil
}

// getDefaultRecommendation 获取默认推荐（当AI解析失败时使用）
func (s *GLMService) getDefaultRecommendation(taskNode TaskNodeInfo, candidates []EmployeeCandidate) *DispatchRecommendation {
	recommendation := &DispatchRecommendation{
		TaskNodeID:   taskNode.NodeID,
		TaskNodeName: taskNode.NodeName,
		AIAnalysis:   "基于工作负载和经验的默认推荐",
	}

	// 简单排序：按活跃任务数升序，完成任务数降序
	type scored struct {
		candidate EmployeeCandidate
		score     float64
	}
	var scoredList []scored
	for _, c := range candidates {
		// 简单评分：完成任务多、活跃任务少的优先
		score := float64(c.CompletedTasks)*0.5 - float64(c.ActiveTasks)*0.3 + float64(c.TenureMonths)*0.2
		scoredList = append(scoredList, scored{candidate: c, score: score})
	}

	// 排序
	for i := 0; i < len(scoredList)-1; i++ {
		for j := i + 1; j < len(scoredList); j++ {
			if scoredList[j].score > scoredList[i].score {
				scoredList[i], scoredList[j] = scoredList[j], scoredList[i]
			}
		}
	}

	// 取前5个
	for i, s := range scoredList {
		if i >= 5 {
			break
		}
		recommendation.Candidates = append(recommendation.Candidates, RecommendedEmployee{
			EmployeeID: s.candidate.EmployeeID,
			Name:       s.candidate.Name,
			Score:      s.score * 10, // 转换为0-100分
			Reason:     fmt.Sprintf("已完成%d个任务，当前%d个活跃任务", s.candidate.CompletedTasks, s.candidate.ActiveTasks),
			Rank:       i + 1,
		})
	}

	return recommendation
}

// CallGLMWithPrompt 通用GLM调用方法
func (s *GLMService) CallGLMWithPrompt(ctx context.Context, prompt string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("GLM API Key未配置")
	}
	return s.callGLM(ctx, prompt)
}
