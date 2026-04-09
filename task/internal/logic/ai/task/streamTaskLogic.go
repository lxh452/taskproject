package task

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"task_Project/task/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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
		// 默认 nodeType 为 2（开发任务）
		nodeType := st.NodeType
		if nodeType == 0 {
			nodeType = 2
		}
		// 将优先级转换为 nodePriority (1-5 -> 0-3)
		nodePriority := int64(2) // 默认为中
		if st.Priority >= 4 {
			nodePriority = 0 // 紧急
		} else if st.Priority == 3 {
			nodePriority = 1 // 高
		} else if st.Priority == 2 {
			nodePriority = 2 // 中
		} else {
			nodePriority = 3 // 低
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
			NodePriority:   nodePriority,
		})
	}

	return subtasks, nil
}
