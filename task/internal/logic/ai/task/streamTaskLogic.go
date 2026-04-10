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

// GenerateSubtasksRequest 生成子任务请求
type GenerateSubtasksRequest struct {
	TaskID     string                 `json:"taskId,optional"`     // 任务 ID
	TaskTitle  string                 `json:"taskTitle,optional"`  // 任务标题
	TaskDetail string                 `json:"taskDetail,optional"` // 任务详情
	Context    map[string]interface{} `json:"context,optional"`    // 上下文信息
}

// Subtask 子任务
type Subtask struct {
	SubtaskID      string `json:"subtaskId,optional"` // 子任务 ID
	Title          string `json:"title"`              // 子任务标题（兼容旧字段）
	Description    string `json:"description"`        // 子任务描述（兼容旧字段）
	NodeType       int    `json:"nodeType"`           // 节点类型 1=里程碑 2=开发任务 3=测试任务 4=文档任务 5=评审任务
	NodeName       string `json:"nodeName"`           // 节点名称
	NodeDetail     string `json:"nodeDetail"`         // 节点详细描述
	EstimatedHours int    `json:"estimatedHours"`     // 预估工时
	Priority       int    `json:"priority"`           // 优先级
	NodePriority   int64  `json:"nodePriority"`       // 节点优先级（用于创建节点）
}

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
	l.Infof("流式生成子任务：taskId=%s, taskTitle=%s", req.TaskID, req.TaskTitle)

	// 检查 AI 服务是否可用
	if l.svcCtx.GLMService == nil {
		l.Infof("AI 服务未配置，使用降级方案返回默认子任务")
		// 降级方案：返回默认子任务
		l.sendDefaultSubtasks(w, req)
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
		l.Errorf("流式调用 GLM 失败：%v", err)
		l.sendEvent(w, "error", map[string]interface{}{"message": err.Error()})
		return
	}

	// 解析完整内容
	subtasks, err := l.parseSubtaskResponse(fullContent.String())
	if err != nil {
		l.Errorf("解析子任务失败：%v", err)
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

// sendDefaultSubtasks 发送默认子任务（降级方案）
func (l *StreamGenerateSubtasksLogic) sendDefaultSubtasks(w http.ResponseWriter, req *GenerateSubtasksRequest) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	l.sendEvent(w, "start", map[string]interface{}{"message": "开始生成子任务"})

	// 返回默认子任务
	defaultSubtasks := []Subtask{
		{
			NodeType:       1,
			NodeName:       "项目启动",
			NodeDetail:     "明确项目目标、范围和交付物，组建项目团队",
			EstimatedHours: 4,
			NodePriority:   1,
		},
		{
			NodeType:       2,
			NodeName:       "需求分析",
			NodeDetail:     "详细分析需求，编写需求文档",
			EstimatedHours: 8,
			NodePriority:   1,
		},
		{
			NodeType:       2,
			NodeName:       "系统设计",
			NodeDetail:     "完成系统架构设计和详细设计",
			EstimatedHours: 12,
			NodePriority:   2,
		},
		{
			NodeType:       2,
			NodeName:       "编码实现",
			NodeDetail:     "按照设计文档进行代码开发",
			EstimatedHours: 40,
			NodePriority:   2,
		},
		{
			NodeType:       3,
			NodeName:       "测试验证",
			NodeDetail:     "执行单元测试、集成测试和系统测试",
			EstimatedHours: 16,
			NodePriority:   2,
		},
		{
			NodeType:       5,
			NodeName:       "项目验收",
			NodeDetail:     "项目交付物验收和总结",
			EstimatedHours: 4,
			NodePriority:   3,
		},
	}

	l.sendEvent(w, "complete", map[string]interface{}{
		"subtasks": defaultSubtasks,
		"total":    len(defaultSubtasks),
	})
}

// buildSubtaskPrompt 构建生成子任务的 Prompt
func (l *StreamGenerateSubtasksLogic) buildSubtaskPrompt(req *GenerateSubtasksRequest) string {
	contextStr := ""

	// 构建更丰富的上下文信息
	if req.Context != nil {
		// 提取部门信息
		departmentIds, _ := req.Context["departmentIds"].([]interface{})
		departmentCount, _ := req.Context["departmentCount"].(float64)
		isCrossDepartment, _ := req.Context["isCrossDepartment"].(bool)

		// 构建部门相关的上下文
		contextMap := make(map[string]interface{})
		if len(departmentIds) > 0 {
			contextMap["departmentIds"] = departmentIds
			contextMap["departmentCount"] = departmentCount
			contextMap["isCrossDepartment"] = isCrossDepartment

			// 将部门信息添加到 Prompt 中
			deptInfo := fmt.Sprintf("\n\n## 任务范围信息\n- 涉及部门数量：%d 个\n- 任务类型：%s\n- 涉及部门 ID 列表：%v",
				int(departmentCount),
				map[bool]string{true: "跨部门协作", false: "单部门任务"}[isCrossDepartment],
				departmentIds)

			if isCrossDepartment {
				deptInfo += "\n- 注意：这是一个跨部门任务，需要协调多个部门的资源和人员，请在任务节点中体现跨部门协作的特点"
			}

			contextStr = deptInfo
		}

		// 提取节点信息映射
		nodeInfoMap, hasNodeInfo := req.Context["nodeInfoMap"].(map[string]interface{})
		if hasNodeInfo && len(nodeInfoMap) > 0 {
			// 构建节点信息字符串
			nodeInfoStr := "\n\n## 已有节点信息（ID 到名称的映射）\n"
			for nodeID, nodeName := range nodeInfoMap {
				nodeInfoStr += fmt.Sprintf("- 节点 ID: %s, 节点名称：%v\n", nodeID, nodeName)
			}
			nodeInfoStr += "\n请在生成子任务时，参考上述节点信息，确保生成的子任务与已有节点保持一致性和连贯性。"
			contextStr += nodeInfoStr
		}

		// 提取完整的节点列表
		flowNodes, hasFlowNodes := req.Context["flowNodes"].([]interface{})
		if hasFlowNodes && len(flowNodes) > 0 {
			nodeListStr := fmt.Sprintf("\n\n## 流程节点列表（共 %d 个节点）\n", len(flowNodes))
			for i, node := range flowNodes {
				if nodeMap, ok := node.(map[string]interface{}); ok {
					nodeID := ""
					if idVal, ok := nodeMap["id"]; ok {
						nodeID = fmt.Sprintf("%v", idVal)
					}
					nodeLabel := ""
					if data, ok := nodeMap["data"].(map[string]interface{}); ok {
						if labelVal, ok := data["label"]; ok {
							nodeLabel = fmt.Sprintf("%v", labelVal)
						}
					}
					nodeListStr += fmt.Sprintf("%d. 节点 ID: %s, 名称：%s\n", i+1, nodeID, nodeLabel)
				}
			}
			nodeListStr += "\n请根据上述流程节点的结构和内容，生成对应的子任务，确保任务拆解与流程设计保持一致。"
			contextStr += nodeListStr
		}

		// 如果有其他上下文信息，也添加到 JSON 中
		for k, v := range req.Context {
			if k != "departmentIds" && k != "departmentCount" && k != "isCrossDepartment" &&
				k != "nodeInfoMap" && k != "flowNodes" {
				contextMap[k] = v
			}
		}

		if len(contextMap) > 0 {
			contextJSON, _ := json.Marshal(contextMap)
			if contextStr != "" {
				contextStr += "\n- 其他上下文：" + string(contextJSON)
			} else {
				contextStr = string(contextJSON)
			}
		}
	}

	return fmt.Sprintf(`你是一个专业的任务拆解专家。请根据以下主任务信息，智能生成合理的任务节点列表。

## 主任务信息
- 任务标题：%s
- 任务详情：%s%s

## 生成要求
1. 将主任务拆解为 3-8 个可执行的任务节点
2. 每个节点必须包含以下字段：
   - nodeType: 节点类型 (1=里程碑，2=开发任务，3=测试任务，4=文档任务，5=评审任务)
   - nodeName: 节点名称（简洁明了，10 字以内）
   - nodeDetail: 节点详细描述（包含具体要做什么、输出物、验收标准）
   - estimatedHours: 预估工时（小时）
   - nodePriority: 节点优先级 (0=紧急，1=高，2=中，3=低)
3. 节点之间应该有合理的依赖关系和执行顺序
4. 预估工时应该合理，总工时应与任务复杂度匹配
5. 第一个节点通常是里程碑或需求分析，最后一个是测试验收
6. 根据任务的重要性和紧急程度合理设置 nodePriority
7. 如果是跨部门任务，请在任务节点中体现跨部门协作、沟通协调的特点
8. 如果提供了已有节点信息（nodeInfoMap 或 flowNodes），请确保生成的子任务与已有节点保持一致

## 输出格式
请严格按照以下 JSON 格式输出，不要包含任何其他内容：
{
  "subtasks": [
    {
      "nodeType": 节点类型数字，
      "nodeName": "节点名称",
      "nodeDetail": "节点详细描述",
      "estimatedHours": 预估工时数字，
      "nodePriority": 优先级数字 (0-3)
    }
  ]
}`, req.TaskTitle, req.TaskDetail, contextStr)
}

// getStringValue 辅助函数，安全地获取 interface{} 的字符串值
func getStringValue(val interface{}) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
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
			NodePriority   int    `json:"nodePriority"`
		} `json:"subtasks"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败：%v", err)
	}

	// 转换为 Subtask 结构
	subtasks := make([]Subtask, 0, len(result.Subtasks))
	for _, st := range result.Subtasks {
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
		nodePriority := st.NodePriority
		if nodePriority == 0 {
			if st.Priority >= 4 {
				nodePriority = 3
			} else if st.Priority == 3 {
				nodePriority = 2
			} else if st.Priority == 2 {
				nodePriority = 1
			} else {
				nodePriority = 2 // 默认中优先级
			}
		}
		// 确保 nodePriority 在 0-3 范围内
		if nodePriority < 0 {
			nodePriority = 0
		} else if nodePriority > 3 {
			nodePriority = 3
		}

		subtasks = append(subtasks, Subtask{
			NodeType:       nodeType,
			NodeName:       title,
			NodeDetail:     description,
			EstimatedHours: st.EstimatedHours,
			NodePriority:   int64(nodePriority),
		})
	}

	return subtasks, nil
}
