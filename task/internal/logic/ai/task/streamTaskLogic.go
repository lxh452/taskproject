package task

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"task_Project/task/internal/svc"
)

// StreamGenerateSubtasksLogic 流式生成子任务逻辑
type StreamGenerateSubtasksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewStreamGenerateSubtasksLogic 创建新的流式生成子任务逻辑
func NewStreamGenerateSubtasksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StreamGenerateSubtasksLogic {
	return &StreamGenerateSubtasksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// StreamGenerateSubtasks 流式生成子任务
func (l *StreamGenerateSubtasksLogic) StreamGenerateSubtasks(req *GenerateSubtasksRequest, w http.ResponseWriter) {
	l.Infof("流式生成子任务: taskId=%s, taskTitle=%s", req.TaskID, req.TaskTitle)

	// 检查 AI 服务是否可用
	if l.svcCtx.GLMService == nil {
		l.sendError(w, "AI 服务未配置")
		return
	}

	// 构建 Prompt
	prompt := l.buildSubtaskPrompt(req)

	// 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 发送开始事件
	l.sendEvent(w, "start", map[string]interface{}{"message": "开始生成子任务"})

	// 收集完整内容用于解析
	var fullContent strings.Builder
	chunkCount := 0

	// 调用流式 GLM API
	err := l.svcCtx.GLMService.StreamCallGLMWithPrompt(l.ctx, prompt, func(chunk string) {
		fullContent.WriteString(chunk)
		chunkCount++
		// 发送进度事件
		l.sendEvent(w, "chunk", map[string]interface{}{
			"content": chunk,
			"index":   chunkCount,
		})
	})

	if err != nil {
		l.Errorf("流式调用 GLM 失败: %v", err)
		l.sendEvent(w, "error", map[string]interface{}{"message": err.Error()})
		return
	}

	// 解析完整内容
	subtasks, err := l.parseSubtaskResponse(fullContent.String())
	if err != nil {
		l.Errorf("解析子任务失败: %v", err)
		// 即使解析失败，也返回原始内容给前端
		l.sendEvent(w, "raw_content", map[string]interface{}{
			"content": fullContent.String(),
		})
		l.sendEvent(w, "error", map[string]interface{}{
			"message": "解析失败，返回原始内容",
		})
		return
	}

	// 发送完成事件
	l.sendEvent(w, "complete", map[string]interface{}{
		"subtasks": subtasks,
		"total":    len(subtasks),
	})
}

// sendEvent 发送 SSE 事件
func (l *StreamGenerateSubtasksLogic) sendEvent(w http.ResponseWriter, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(jsonData))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// sendError 发送错误事件
func (l *StreamGenerateSubtasksLogic) sendError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	l.sendEvent(w, "error", map[string]interface{}{"message": message})
}

// buildSubtaskPrompt 构建生成子任务的 Prompt
func (l *StreamGenerateSubtasksLogic) buildSubtaskPrompt(req *GenerateSubtasksRequest) string {
	contextStr := ""
	if req.Context != nil {
		contextJSON, _ := json.Marshal(req.Context)
		contextStr = string(contextJSON)
	}

	return fmt.Sprintf(`你是一个专业的任务拆解专家。请根据以下主任务信息，智能生成合理的任务节点列表。

## 主任务信息
- 任务标题: %s
- 任务详情: %s
- 附加信息: %s

## 生成要求
1. 将主任务拆解为 3-8 个可执行的任务节点
2. 每个节点必须包含以下字段：
   - nodeType: 节点类型 (1=里程碑, 2=开发任务, 3=测试任务, 4=文档任务, 5=评审任务)
   - nodeName: 节点名称（简洁明了，10字以内）
   - nodeDetail: 节点详细描述（包含具体要做什么、输出物、验收标准）
   - estimatedHours: 预估工时（小时）
3. 节点之间应该有合理的依赖关系和执行顺序
4. 预估工时应该合理，总工时应与任务复杂度匹配
5. 第一个节点通常是里程碑或需求分析，最后一个是测试验收

## 输出格式
请严格按照以下 JSON 格式输出，不要包含任何其他内容：
{
  "subtasks": [
    {
      "nodeType": 节点类型数字,
      "nodeName": "节点名称",
      "nodeDetail": "节点详细描述",
      "estimatedHours": 预估工时数字
    }
  ]
}`, req.TaskTitle, req.TaskDetail, contextStr)
}

// parseSubtaskResponse 解析 GLM 返回的子任务
func (l *StreamGenerateSubtasksLogic) parseSubtaskResponse(response string) ([]Subtask, error) {
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
			NodeType       int    `json:"nodeType"`
			NodeName       string `json:"nodeName"`
			NodeDetail     string `json:"nodeDetail"`
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
		// 兼容新旧字段
		title := st.NodeName
		if title == "" {
			title = st.Title
		}
		description := st.NodeDetail
		if description == "" {
			description = st.Description
		}
		// 默认nodeType为2（开发任务）
		nodeType := st.NodeType
		if nodeType == 0 {
			nodeType = 2
		}

		subtasks = append(subtasks, Subtask{
			SubtaskID:      fmt.Sprintf("sub_%03d", i+1),
			Title:          title,
			Description:    description,
			NodeType:       nodeType,
			NodeName:       title,
			NodeDetail:     description,
			EstimatedHours: st.EstimatedHours,
			Priority:       st.Priority,
		})
	}

	return subtasks, nil
}

// StreamPolishTaskLogic 流式润色任务逻辑
type StreamPolishTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewStreamPolishTaskLogic 创建新的流式润色任务逻辑
func NewStreamPolishTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StreamPolishTaskLogic {
	return &StreamPolishTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// StreamPolishTask 流式润色任务
func (l *StreamPolishTaskLogic) StreamPolishTask(req *PolishTaskRequest, w http.ResponseWriter) {
	l.Infof("流式润色任务: taskId=%s, polishType=%s", req.TaskID, req.PolishType)

	// 检查 AI 服务是否可用
	if l.svcCtx.GLMService == nil {
		l.sendPolishError(w, "AI 服务未配置")
		return
	}

	// 构建 Prompt
	prompt := l.buildPolishPrompt(req)

	// 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 发送开始事件
	l.sendPolishEvent(w, "start", map[string]interface{}{"message": "开始润色任务"})

	// 收集完整内容
	var fullContent strings.Builder
	chunkCount := 0

	// 调用流式 GLM API
	err := l.svcCtx.GLMService.StreamCallGLMWithPrompt(l.ctx, prompt, func(chunk string) {
		fullContent.WriteString(chunk)
		chunkCount++
		l.sendPolishEvent(w, "chunk", map[string]interface{}{
			"content": chunk,
			"index":   chunkCount,
		})
	})

	if err != nil {
		l.Errorf("流式调用 GLM 失败: %v", err)
		l.sendPolishEvent(w, "error", map[string]interface{}{"message": err.Error()})
		return
	}

	// 解析结果
	result, err := l.parsePolishResponse(fullContent.String(), req)
	if err != nil {
		l.Errorf("解析润色结果失败: %v", err)
		l.sendPolishEvent(w, "raw_content", map[string]interface{}{
			"content": fullContent.String(),
		})
		l.sendPolishEvent(w, "error", map[string]interface{}{
			"message": "解析失败，返回原始内容",
		})
		return
	}

	// 发送完成事件
	l.sendPolishEvent(w, "complete", map[string]interface{}{
		"result": result,
	})
}

// sendPolishEvent 发送 SSE 事件
func (l *StreamPolishTaskLogic) sendPolishEvent(w http.ResponseWriter, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(jsonData))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// sendPolishError 发送错误事件
func (l *StreamPolishTaskLogic) sendPolishError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	l.sendPolishEvent(w, "error", map[string]interface{}{"message": message})
}

// buildPolishPrompt 构建润色 Prompt
func (l *StreamPolishTaskLogic) buildPolishPrompt(req *PolishTaskRequest) string {
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
func (l *StreamPolishTaskLogic) parsePolishResponse(response string, req *PolishTaskRequest) (PolishResult, error) {
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
