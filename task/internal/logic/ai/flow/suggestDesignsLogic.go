// Code scaffolded by VibeCraft AI
// VibeCraft AI Logic

package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
)

// SuggestDesignsLogic 获取流程设计建议逻辑
type SuggestDesignsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// SuggestDesignsRequest 获取流程设计建议请求
type SuggestDesignsRequest struct {
	FlowType    string                 `json:"flowType"`         // 流程类型
	Description string                 `json:"description"`      // 流程描述
	Context     map[string]interface{} `json:"context,optional"` // 上下文信息
}

// DesignSuggestion 设计建议
type DesignSuggestion struct {
	ID          string  `json:"id"`          // 建议ID
	Name        string  `json:"name"`        // 建议名称
	Description string  `json:"description"` // 建议描述
	Confidence  float32 `json:"confidence"`  // 置信度
}

// FlowDesign 流程设计方案（与前端对齐）
type FlowDesign struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Type           string   `json:"type"`
	Description    string   `json:"description"`
	EstimatedDays  int      `json:"estimatedDays"`
	RequiredPeople int      `json:"requiredPeople"`
	RiskLevel      string   `json:"riskLevel"`
	IsRecommended  bool     `json:"isRecommended"`
	Steps          []string `json:"steps,omitempty"`
	Pros           []string `json:"pros,omitempty"`
	Cons           []string `json:"cons,omitempty"`
	Confidence     float32  `json:"confidence"`
}

// SuggestDesignsResponse 获取流程设计建议响应
type SuggestDesignsResponse struct {
	Designs []FlowDesign `json:"designs"` // 设计方案列表
	Total   int          `json:"total"`   // 总数
}

// NewSuggestDesignsLogic 创建新的 SuggestDesignsLogic
func NewSuggestDesignsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SuggestDesignsLogic {
	return &SuggestDesignsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SuggestDesigns 获取流程设计建议
func (l *SuggestDesignsLogic) SuggestDesigns(req *SuggestDesignsRequest) (*types.BaseResponse, error) {
	l.Infof("获取流程设计建议: flowType=%s", req.FlowType)

	// 检查 AI 服务是否可用
	if l.svcCtx.GLMService == nil {
		return nil, fmt.Errorf("AI 服务未配置，无法获取设计建议")
	}

	// 将 SuggestDesignsRequest 转换为 DesignRequest
	designReq := &DesignRequest{
		ProjectID:    req.FlowType,
		Title:        req.FlowType,
		Description:  req.Description,
		Requirements: []string{},
		Style:        "高效实用",
	}

	// 调用整合后的设计方案生成逻辑
	result, err := l.GenerateDesigns(designReq)
	if err != nil {
		l.Errorf("生成设计方案失败: %v", err)
		// 返回默认方案
		result = l.generateDefaultDesigns(designReq)
	}

	// 将 DesignResult 转换为 SuggestDesignsResponse
	designs := make([]FlowDesign, 0, len(result.Schemes))
	for i, scheme := range result.Schemes {
		riskLevel := "medium"
		if scheme.Type == "parallel" {
			riskLevel = "high"
		} else if scheme.Type == "sequential" {
			riskLevel = "low"
		}

		designs = append(designs, FlowDesign{
			ID:             scheme.ID,
			Name:           scheme.Name,
			Type:           scheme.Type,
			Description:    scheme.Description,
			EstimatedDays:  scheme.Estimated / 8, // 将小时转换为天
			RequiredPeople: len(scheme.Resources),
			RiskLevel:      riskLevel,
			IsRecommended:  i == 2, // 默认推荐第三个方案（混合方案）
			Steps:          scheme.Steps,
			Pros:           scheme.Pros,
			Cons:           scheme.Cons,
			Confidence:     scheme.Confidence,
		})
	}

	return &types.BaseResponse{
		Code: 0,
		Msg:  "success",
		Data: SuggestDesignsResponse{
			Designs: designs,
			Total:   len(designs),
		},
	}, nil
}

// ==================== 整合自 designGenerator.go 的业务逻辑 ====================

// GenerateDesigns 生成多种设计方案（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) GenerateDesigns(req *DesignRequest) (*DesignResult, error) {
	l.Infof("开始为项目 %s 生成设计方案", req.ProjectID)

	if l.svcCtx.GLMService == nil {
		return nil, fmt.Errorf("GLM service not available")
	}

	// 创建独立的context，设置超时
	aiCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 调用AI生成设计方案
	result, err := l.generateWithAI(aiCtx, req)
	if err != nil {
		l.Errorf("AI生成设计方案失败: %v", err)
		// 返回默认方案
		return l.generateDefaultDesigns(req), nil
	}

	result.GeneratedAt = time.Now().Unix()
	l.Infof("成功为项目 %s 生成 %d 个设计方案", req.ProjectID, len(result.Schemes))
	return result, nil
}

// GenerateParallelDesign 生成并行执行方案（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) GenerateParallelDesign(req *DesignRequest) (*DesignScheme, error) {
	l.Infof("生成并行执行方案: %s", req.ProjectID)

	prompt := l.buildParallelPrompt(req)
	return l.generateSingleDesign(req.ProjectID, "parallel", prompt)
}

// GenerateSequentialDesign 生成串行执行方案（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) GenerateSequentialDesign(req *DesignRequest) (*DesignScheme, error) {
	l.Infof("生成串行执行方案: %s", req.ProjectID)

	prompt := l.buildSequentialPrompt(req)
	return l.generateSingleDesign(req.ProjectID, "sequential", prompt)
}

// GenerateHybridDesign 生成混合执行方案（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) GenerateHybridDesign(req *DesignRequest) (*DesignScheme, error) {
	l.Infof("生成混合执行方案: %s", req.ProjectID)

	prompt := l.buildHybridPrompt(req)
	return l.generateSingleDesign(req.ProjectID, "hybrid", prompt)
}

// generateWithAI 调用AI生成设计方案（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) generateWithAI(ctx context.Context, req *DesignRequest) (*DesignResult, error) {
	prompt := l.buildDesignPrompt(req)

	response, err := l.svcCtx.GLMService.CallGLMWithPrompt(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return l.parseDesignResponse(response, req.ProjectID)
}

// generateSingleDesign 生成单个设计方案（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) generateSingleDesign(projectID, designType, prompt string) (*DesignScheme, error) {
	if l.svcCtx.GLMService == nil {
		return l.createDefaultScheme(projectID, designType), nil
	}

	aiCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := l.svcCtx.GLMService.CallGLMWithPrompt(aiCtx, prompt)
	if err != nil {
		return l.createDefaultScheme(projectID, designType), nil
	}

	scheme, err := l.parseSingleDesignResponse(response, projectID, designType)
	if err != nil {
		return l.createDefaultScheme(projectID, designType), nil
	}

	return scheme, nil
}

// buildDesignPrompt 构建设计方案提示词（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) buildDesignPrompt(req *DesignRequest) string {
	requirements := strings.Join(req.Requirements, "\n- ")
	style := req.Style
	if style == "" {
		style = "高效实用"
	}

	return fmt.Sprintf(`你是一位专业的项目设计专家。请为以下项目设计多种执行方案。

## 项目信息
- 项目ID: %s
- 标题: %s
- 描述: %s
- 风格偏好: %s

## 需求列表
- %s

## 设计要求
1. 生成3种不同的设计方案：并行执行方案、串行执行方案、混合执行方案
2. 每种方案需包含：名称、描述、执行步骤、优缺点、所需资源、预估工时
3. 分析每种方案的适用场景
4. 给出方案选择建议

## 输出格式
请严格按照以下JSON格式输出：
{
  "strategy": "推荐策略说明",
  "analysis": "整体分析",
  "schemes": [
    {
      "id": "scheme_001",
      "name": "方案名称",
      "type": "parallel|sequential|hybrid",
      "description": "方案描述",
      "steps": ["步骤1", "步骤2"],
      "pros": ["优点1"],
      "cons": ["缺点1"],
      "resources": ["资源1"],
      "estimated": 预估小时数,
      "confidence": 0.95
    }
  ]
}

只返回JSON，不要包含其他说明文字。`, req.ProjectID, req.Title, req.Description, style, requirements)
}

// buildParallelPrompt 构建并行方案提示词（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) buildParallelPrompt(req *DesignRequest) string {
	return fmt.Sprintf(`为项目设计并行执行方案。

项目: %s
描述: %s
需求: %s

请设计一个可以并行执行的方案，最大化利用资源，缩短总工期。
返回JSON格式：{"name":"","description":"","steps":[],"pros":[],"cons":[],"resources":[],"estimated":0,"confidence":0}`,
		req.Title, req.Description, strings.Join(req.Requirements, ", "))
}

// buildSequentialPrompt 构建串行方案提示词（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) buildSequentialPrompt(req *DesignRequest) string {
	return fmt.Sprintf(`为项目设计串行执行方案。

项目: %s
描述: %s
需求: %s

请设计一个按步骤顺序执行的方案，确保质量可控，风险较低。
返回JSON格式：{"name":"","description":"","steps":[],"pros":[],"cons":[],"resources":[],"estimated":0,"confidence":0}`,
		req.Title, req.Description, strings.Join(req.Requirements, ", "))
}

// buildHybridPrompt 构建混合方案提示词（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) buildHybridPrompt(req *DesignRequest) string {
	return fmt.Sprintf(`为项目设计混合执行方案。

项目: %s
描述: %s
需求: %s

请设计一个串行和并行结合的方案，在关键路径上串行，非关键路径上并行。
返回JSON格式：{"name":"","description":"","steps":[],"pros":[],"cons":[],"resources":[],"estimated":0,"confidence":0}`,
		req.Title, req.Description, strings.Join(req.Requirements, ", "))
}

// parseDesignResponse 解析设计方案响应（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) parseDesignResponse(response, projectID string) (*DesignResult, error) {
	result := &DesignResult{
		ProjectID: projectID,
		Schemes:   []DesignScheme{},
	}

	// 提取JSON
	jsonStr := l.extractJSON(response)
	if jsonStr == "" {
		return result, fmt.Errorf("no JSON found in response")
	}

	var parsed struct {
		Strategy string         `json:"strategy"`
		Analysis string         `json:"analysis"`
		Schemes  []DesignScheme `json:"schemes"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return result, fmt.Errorf("failed to parse JSON: %w", err)
	}

	result.Strategy = parsed.Strategy
	result.Analysis = parsed.Analysis
	result.Schemes = parsed.Schemes

	// 设置ID
	for i := range result.Schemes {
		if result.Schemes[i].ID == "" {
			result.Schemes[i].ID = fmt.Sprintf("scheme_%s_%d", projectID, i)
		}
	}

	return result, nil
}

// parseSingleDesignResponse 解析单个设计方案响应（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) parseSingleDesignResponse(response, projectID, designType string) (*DesignScheme, error) {
	jsonStr := l.extractJSON(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found")
	}

	var scheme DesignScheme
	if err := json.Unmarshal([]byte(jsonStr), &scheme); err != nil {
		return nil, err
	}

	scheme.ID = fmt.Sprintf("scheme_%s_%s", projectID, designType)
	scheme.Type = designType

	return &scheme, nil
}

// generateDefaultDesigns 生成默认设计方案（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) generateDefaultDesigns(req *DesignRequest) *DesignResult {
	return &DesignResult{
		ProjectID: req.ProjectID,
		Strategy:  "根据项目特点，推荐混合执行方案",
		Analysis:  "项目需求明确，可采用灵活的执行策略",
		Schemes: []DesignScheme{
			{
				ID:          fmt.Sprintf("scheme_%s_parallel", req.ProjectID),
				Name:        "并行快速方案",
				Type:        "parallel",
				Description: "多个任务并行执行，缩短总工期",
				Steps:       []string{"任务分解", "并行执行", "结果整合", "质量检查"},
				Pros:        []string{"工期短", "效率高", "资源利用率高"},
				Cons:        []string{"协调复杂", "资源需求大", "风险较高"},
				Resources:   []string{"项目经理", "开发团队", "测试团队"},
				Estimated:   40,
				Confidence:  0.85,
			},
			{
				ID:          fmt.Sprintf("scheme_%s_sequential", req.ProjectID),
				Name:        "串行稳妥方案",
				Type:        "sequential",
				Description: "按步骤顺序执行，确保每个阶段质量",
				Steps:       []string{"需求分析", "设计", "开发", "测试", "部署"},
				Pros:        []string{"质量可控", "风险低", "易于管理"},
				Cons:        []string{"工期较长", "效率较低"},
				Resources:   []string{"项目经理", "开发人员"},
				Estimated:   80,
				Confidence:  0.90,
			},
			{
				ID:          fmt.Sprintf("scheme_%s_hybrid", req.ProjectID),
				Name:        "混合优化方案",
				Type:        "hybrid",
				Description: "关键路径串行，非关键路径并行",
				Steps:       []string{"关键路径分析", "任务分组", "并行组执行", "串行集成", "最终验收"},
				Pros:        []string{"兼顾效率和质量", "灵活可控"},
				Cons:        []string{"需要精确规划", "管理复杂度中等"},
				Resources:   []string{"项目经理", "开发团队"},
				Estimated:   60,
				Confidence:  0.88,
			},
		},
		GeneratedAt: time.Now().Unix(),
	}
}

// createDefaultScheme 创建默认方案（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) createDefaultScheme(projectID, designType string) *DesignScheme {
	return &DesignScheme{
		ID:          fmt.Sprintf("scheme_%s_%s", projectID, designType),
		Name:        fmt.Sprintf("%s执行方案", getTypeName(designType)),
		Type:        designType,
		Description: fmt.Sprintf("采用%s方式执行", getTypeName(designType)),
		Steps:       []string{"分析", "设计", "执行", "验收"},
		Pros:        []string{"执行效率高", "质量可控"},
		Cons:        []string{"需要协调", "资源需求"},
		Resources:   []string{"项目经理", "执行团队"},
		Estimated:   50,
		Confidence:  0.80,
	}
}

// extractJSON 从响应中提取JSON（整合自 designGenerator.go）
func (l *SuggestDesignsLogic) extractJSON(text string) string {
	// 移除markdown代码块标记
	text = strings.TrimSpace(text)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	// 查找JSON对象
	startIdx := strings.Index(text, "{")
	endIdx := strings.LastIndex(text, "}")

	if startIdx >= 0 && endIdx > startIdx {
		return text[startIdx : endIdx+1]
	}

	return ""
}
