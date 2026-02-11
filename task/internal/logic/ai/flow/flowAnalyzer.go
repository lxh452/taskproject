package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// FlowAnalyzer 流程分析器
type FlowAnalyzer struct {
	aiClient GLMClient
}

// GLMClient 定义 GLM 客户端接口
type GLMClient interface {
	Generate(ctx context.Context, prompt string, config map[string]interface{}) (string, error)
	GenerateWithContext(ctx context.Context, prompt string, context map[string]interface{}, config map[string]interface{}) (string, error)
}

// FlowAnalysisResult 流程分析结果
type FlowAnalysisResult struct {
	FlowID          string           `json:"flow_id"`
	FlowType        string           `json:"flow_type"`
	Complexity      FlowComplexity   `json:"complexity"`
	Stages          []FlowStage      `json:"stages"`
	Bottlenecks     []Bottleneck     `json:"bottlenecks"`
	Recommendations []Recommendation `json:"recommendations"`
	Metrics         FlowMetrics      `json:"metrics"`
	AnalyzedAt      int64            `json:"analyzed_at"`
}

// FlowComplexity 流程复杂度
type FlowComplexity struct {
	Level       string `json:"level"` // simple, moderate, complex
	Score       int    `json:"score"` // 1-100
	StageCount  int    `json:"stage_count"`
	BranchCount int    `json:"branch_count"`
}

// FlowStage 流程阶段
type FlowStage struct {
	Order         int    `json:"order"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Description   string `json:"description"`
	EstimatedTime int    `json:"estimated_time"` // 预估耗时（分钟）
	Dependencies  []int  `json:"dependencies"`
}

// Bottleneck 瓶颈分析
type Bottleneck struct {
	StageID     string `json:"stage_id"`
	StageName   string `json:"stage_name"`
	Type        string `json:"type"`     // resource, dependency, approval
	Severity    string `json:"severity"` // low, medium, high, critical
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

// Recommendation 优化建议
type Recommendation struct {
	Category       string `json:"category"`
	Priority       string `json:"priority"` // low, medium, high
	Title          string `json:"title"`
	Description    string `json:"description"`
	Action         string `json:"action"`
	ExpectedImpact string `json:"expected_impact"`
}

// FlowMetrics 流程指标
type FlowMetrics struct {
	TotalEstimatedTime    int                `json:"total_estimated_time"`
	ParallelOpportunities int                `json:"parallel_opportunities"`
	AutomationPotential   float32            `json:"automation_potential"`
	RiskScore             int                `json:"risk_score"`
	CustomMetrics         map[string]float32 `json:"custom_metrics,omitempty"`
}

// AnalyzeRequest 分析请求
type AnalyzeRequest struct {
	FlowID      string                 `json:"flow_id"`
	FlowType    string                 `json:"flow_type"`
	Description string                 `json:"description"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// NewFlowAnalyzer 创建新的流程分析器
func NewFlowAnalyzer(aiClient GLMClient) *FlowAnalyzer {
	return &FlowAnalyzer{
		aiClient: aiClient,
	}
}

// AnalyzeFlow 分析流程
func (fa *FlowAnalyzer) AnalyzeFlow(ctx context.Context, req *AnalyzeRequest) (*FlowAnalysisResult, error) {
	logx.Infof("Analyzing flow: %s (type: %s)", req.FlowID, req.FlowType)

	// 构建分析提示词
	prompt := fa.buildAnalysisPrompt(req)

	// 配置 AI 参数
	config := map[string]interface{}{
		"temperature": 0.3,
		"max_tokens":  2000,
	}

	// 调用 AI 服务
	output, err := fa.aiClient.GenerateWithContext(ctx, prompt, req.Context, config)
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	// 解析 AI 输出
	result, err := fa.parseAnalysisResult(output, req.FlowID, req.FlowType)
	if err != nil {
		logx.Errorf("Failed to parse analysis result: %v", err)
		// 即使解析失败，也返回基本结果
		return fa.createBasicResult(req.FlowID, req.FlowType, output), nil
	}

	result.AnalyzedAt = time.Now().Unix()
	logx.Infof("Flow analysis completed: %s (complexity: %s, score: %d)",
		req.FlowID, result.Complexity.Level, result.Complexity.Score)

	return result, nil
}

// AnalyzeComplexity 分析流程复杂度
func (fa *FlowAnalyzer) AnalyzeComplexity(ctx context.Context, description string, context map[string]interface{}) (*FlowComplexity, error) {
	prompt := fmt.Sprintf(`请分析以下工作流程的复杂度。

流程描述：
%s

请从以下几个方面分析并返回 JSON 格式：
{
  "level": "simple|moderate|complex",
  "score": 1-100 的整数,
  "stage_count": 预估阶段数,
  "branch_count": 预估分支/条件数
}

只返回 JSON，不要有其他内容。`, description)

	config := map[string]interface{}{
		"temperature": 0.2,
		"max_tokens":  500,
	}

	output, err := fa.aiClient.GenerateWithContext(ctx, prompt, context, config)
	if err != nil {
		return nil, err
	}

	// 尝试提取 JSON
	jsonStr := extractJSON(output)
	if jsonStr == "" {
		return &FlowComplexity{
			Level:       "moderate",
			Score:       50,
			StageCount:  5,
			BranchCount: 2,
		}, nil
	}

	var complexity FlowComplexity
	if err := json.Unmarshal([]byte(jsonStr), &complexity); err != nil {
		return &FlowComplexity{
			Level:       "moderate",
			Score:       50,
			StageCount:  5,
			BranchCount: 2,
		}, nil
	}

	return &complexity, nil
}

// IdentifyBottlenecks 识别流程瓶颈
func (fa *FlowAnalyzer) IdentifyBottlenecks(ctx context.Context, stages []FlowStage, context map[string]interface{}) ([]Bottleneck, error) {
	stagesJSON, _ := json.Marshal(stages)

	prompt := fmt.Sprintf(`请分析以下流程阶段，识别潜在的瓶颈。

流程阶段：
%s

请识别瓶颈并返回 JSON 数组格式：
[
  {
    "stage_id": "阶段标识",
    "stage_name": "阶段名称",
    "type": "resource|dependency|approval",
    "severity": "low|medium|high|critical",
    "description": "瓶颈描述",
    "impact": "影响说明"
  }
]

最多返回3个最严重的瓶颈，只返回 JSON 数组。`, string(stagesJSON))

	config := map[string]interface{}{
		"temperature": 0.3,
		"max_tokens":  1000,
	}

	output, err := fa.aiClient.GenerateWithContext(ctx, prompt, context, config)
	if err != nil {
		return nil, err
	}

	jsonStr := extractJSON(output)
	if jsonStr == "" {
		return []Bottleneck{}, nil
	}

	var bottlenecks []Bottleneck
	if err := json.Unmarshal([]byte(jsonStr), &bottlenecks); err != nil {
		return []Bottleneck{}, nil
	}

	return bottlenecks, nil
}

// GenerateRecommendations 生成优化建议
func (fa *FlowAnalyzer) GenerateRecommendations(ctx context.Context, bottlenecks []Bottleneck, context map[string]interface{}) ([]Recommendation, error) {
	bottlenecksJSON, _ := json.Marshal(bottlenecks)

	prompt := fmt.Sprintf(`针对以下流程瓶颈，请提供优化建议。

瓶颈列表：
%s

请返回优化建议 JSON 数组：
[
  {
    "category": "分类",
    "priority": "low|medium|high",
    "title": "建议标题",
    "description": "详细描述",
    "action": "具体行动",
    "expected_impact": "预期效果"
  }
]

最多返回5条建议，按优先级排序，只返回 JSON 数组。`, string(bottlenecksJSON))

	config := map[string]interface{}{
		"temperature": 0.4,
		"max_tokens":  1500,
	}

	output, err := fa.aiClient.GenerateWithContext(ctx, prompt, context, config)
	if err != nil {
		return nil, err
	}

	jsonStr := extractJSON(output)
	if jsonStr == "" {
		return []Recommendation{}, nil
	}

	var recommendations []Recommendation
	if err := json.Unmarshal([]byte(jsonStr), &recommendations); err != nil {
		return []Recommendation{}, nil
	}

	return recommendations, nil
}

// CalculateMetrics 计算流程指标
func (fa *FlowAnalyzer) CalculateMetrics(stages []FlowStage) *FlowMetrics {
	totalTime := 0
	for _, stage := range stages {
		totalTime += stage.EstimatedTime
	}

	// 简单计算并行机会
	parallelOpportunities := 0
	for i, stage := range stages {
		if len(stage.Dependencies) > 0 {
			// 检查是否可以并行
			for _, dep := range stage.Dependencies {
				if dep < i-1 {
					parallelOpportunities++
				}
			}
		}
	}

	// 计算自动化潜力（基于阶段类型）
	automationCount := 0
	for _, stage := range stages {
		if stage.Type == "automated" || stage.Type == "script" {
			automationCount++
		}
	}

	automationPotential := float32(0)
	if len(stages) > 0 {
		automationPotential = float32(automationCount) / float32(len(stages)) * 100
	}

	return &FlowMetrics{
		TotalEstimatedTime:    totalTime,
		ParallelOpportunities: parallelOpportunities,
		AutomationPotential:   automationPotential,
		RiskScore:             calculateRiskScore(stages),
	}
}

// buildAnalysisPrompt 构建分析提示词
func (fa *FlowAnalyzer) buildAnalysisPrompt(req *AnalyzeRequest) string {
	return fmt.Sprintf(`请作为流程分析专家，分析以下工作流程。

流程ID：%s
流程类型：%s
流程描述：
%s

请提供以下分析结果，以 JSON 格式返回：
{
  "flow_type": "流程类型",
  "complexity": {
    "level": "simple|moderate|complex",
    "score": 1-100,
    "stage_count": 阶段数,
    "branch_count": 分支数
  },
  "stages": [
    {
      "order": 1,
      "name": "阶段名称",
      "type": "manual|automated|approval|review",
      "description": "阶段描述",
      "estimated_time": 预估分钟数,
      "dependencies": []
    }
  ],
  "bottlenecks": [...],
  "recommendations": [...]
}

请确保返回有效的 JSON 格式。`, req.FlowID, req.FlowType, req.Description)
}

// parseAnalysisResult 解析分析结果
func (fa *FlowAnalyzer) parseAnalysisResult(output, flowID, flowType string) (*FlowAnalysisResult, error) {
	jsonStr := extractJSON(output)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in output")
	}

	var result FlowAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	result.FlowID = flowID
	result.FlowType = flowType

	// 计算指标
	result.Metrics = *fa.CalculateMetrics(result.Stages)

	return &result, nil
}

// createBasicResult 创建基础结果
func (fa *FlowAnalyzer) createBasicResult(flowID, flowType, rawOutput string) *FlowAnalysisResult {
	return &FlowAnalysisResult{
		FlowID:   flowID,
		FlowType: flowType,
		Complexity: FlowComplexity{
			Level:       "moderate",
			Score:       50,
			StageCount:  3,
			BranchCount: 1,
		},
		Stages: []FlowStage{
			{
				Order:         1,
				Name:          "初始阶段",
				Type:          "manual",
				Description:   "流程开始",
				EstimatedTime: 30,
			},
		},
		Bottlenecks:     []Bottleneck{},
		Recommendations: []Recommendation{},
		Metrics: FlowMetrics{
			TotalEstimatedTime:    30,
			ParallelOpportunities: 0,
			AutomationPotential:   0.3,
			RiskScore:             30,
		},
		AnalyzedAt: time.Now().Unix(),
	}
}

// extractJSON 从文本中提取 JSON
func extractJSON(text string) string {
	// 尝试找到 JSON 对象
	re := regexp.MustCompile(`(?s)\{.*\}`)
	match := re.FindString(text)
	if match != "" {
		return match
	}

	// 尝试找到 JSON 数组
	re = regexp.MustCompile(`(?s)\[.*\]`)
	match = re.FindString(text)
	if match != "" {
		return match
	}

	return ""
}

// calculateRiskScore 计算风险评分
func calculateRiskScore(stages []FlowStage) int {
	baseScore := 30

	// 基于阶段数增加风险
	if len(stages) > 10 {
		baseScore += 20
	} else if len(stages) > 5 {
		baseScore += 10
	}

	// 基于手动阶段增加风险
	manualCount := 0
	for _, stage := range stages {
		if stage.Type == "manual" || stage.Type == "approval" {
			manualCount++
		}
	}

	if manualCount > len(stages)/2 {
		baseScore += 15
	}

	if baseScore > 100 {
		return 100
	}
	return baseScore
}
