// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

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

// PolishTaskRequest 润色任务请求
type PolishTaskRequest struct {
	TaskID                 string                 `json:"taskId"`                 // 任务 ID
	TaskTitle              string                 `json:"taskTitle"`              // 任务标题
	TaskDetail             string                 `json:"taskDetail"`             // 任务详情
	PolishType             string                 `json:"polishType"`             // 润色类型：clarity(清晰度), professional(专业性), concise(简洁性)
	CompanyID              string                 `json:"companyId"`              // 公司 ID
	DepartmentIDs          []string               `json:"departmentIds"`          // 涉及部门 ID 列表
	ResponsibleEmployeeIDs []string               `json:"responsibleEmployeeIds"` // 负责人员工 ID 列表
	Context                map[string]interface{} `json:"context"`                // 上下文信息（包含员工、部门等详细信息）
}

// PolishResult 润色结果
type PolishResult struct {
	OriginalTitle          string   `json:"originalTitle"`          // 原标题
	OriginalDetail         string   `json:"originalDetail"`         // 原详情
	PolishedTitle          string   `json:"polishedTitle"`          // 润色后标题
	PolishedDetail         string   `json:"polishedDetail"`         // 润色后详情（不包含部门、人员等结构化信息）
	TaskType               int      `json:"taskType"`               // 任务类型：0-单部门任务，1-跨部门任务
	TaskPriority           int      `json:"taskPriority"`           // 任务优先级：0-不重要不紧急，1-紧急不重要，2-重要但不紧急，3-重要且紧急
	EstimatedDays          int      `json:"estimatedDays"`          // 预估天数
	DepartmentIDs          []string `json:"departmentIds"`          // 涉及部门 ID 列表
	ResponsibleEmployeeIDs []string `json:"responsibleEmployeeIds"` // 负责人员工 ID 列表
	Improvements           []string `json:"improvements"`           // 改进点列表（仅用于调试，前端不显示）
}

// PolishTaskResponse AI 润色任务响应
type PolishTaskResponse struct {
	Result  PolishResult `json:"result"`  // 润色结果
	Success bool         `json:"success"` // 是否成功
}

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

	// 构建任务上下文信息
	contextInfo := ""
	if req.Context != nil {
		// 提取部门信息
		if deptNames, ok := req.Context["departmentNames"].([]string); ok && len(deptNames) > 0 {
			contextInfo += fmt.Sprintf("\n- 涉及部门：%s", strings.Join(deptNames, ", "))
		}

		// 提取负责人信息
		if leaderName, ok := req.Context["leaderName"].(string); ok && leaderName != "" {
			contextInfo += fmt.Sprintf("\n- 任务负责人：%s", leaderName)
		}

		// 提取参与员工信息
		if employeeNames, ok := req.Context["employeeNames"].([]string); ok && len(employeeNames) > 0 {
			contextInfo += fmt.Sprintf("\n- 参与员工：%s", strings.Join(employeeNames, ", "))
		}

		// 任务类型
		if taskType, ok := req.Context["taskType"].(string); ok {
			contextInfo += fmt.Sprintf("\n- 任务类型：%s", taskType)
		}
	}

	return fmt.Sprintf(`你是一个专业的任务描述润色专家。请对以下任务描述进行优化。

## 润色类型
%s

## 原始内容
- 标题：%s
- 详情：%s
%s

## 润色要求
1. 保持原始意图和核心信息不变
2. 根据润色类型进行针对性优化
3. 标题应该简洁明了，概括性强
4. 详情应该结构清晰，重点突出，**不要包含部门名称、人员姓名等具体信息**（这些会作为独立字段返回）
5. 根据任务内容和上下文信息，评估任务优先级和预估天数
6. **重要**：部门 ID 列表和负责人 ID 列表必须作为独立字段返回，不要写入详情中

## 输出格式
请严格按照以下 JSON 格式输出，不要包含任何其他内容：
{
  "polishedTitle": "润色后的标题",
  "polishedDetail": "润色后的详情（不要包含部门、人员等具体信息）",
  "taskType": 1,  // 0-单部门任务，1-跨部门任务（根据涉及部门数量判断）
  "taskPriority": 3,  // 0-不重要不紧急，1-紧急不重要，2-重要但不紧急，3-重要且紧急
  "estimatedDays": 7,  // 预估完成天数
  "departmentIds": ["dept_001", "dept_002"],  // 涉及部门 ID 列表（从上下文获取）
  "responsibleEmployeeIds": ["emp_001", "emp_002"],  // 负责人员工 ID 列表（从上下文获取）
  "improvements": [  // 改进点列表（仅用于调试，前端不显示）
    "改进点 1",
    "改进点 2",
    "改进点 3"
  ]
}`, polishTypeDesc, req.TaskTitle, req.TaskDetail, contextInfo)
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
		PolishedTitle          string   `json:"polishedTitle"`
		PolishedDetail         string   `json:"polishedDetail"`
		TaskType               int      `json:"taskType"`
		TaskPriority           int      `json:"taskPriority"`
		EstimatedDays          int      `json:"estimatedDays"`
		DepartmentIDs          []string `json:"departmentIds"`
		ResponsibleEmployeeIDs []string `json:"responsibleEmployeeIds"`
		Improvements           []string `json:"improvements"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return PolishResult{}, fmt.Errorf("解析 JSON 失败：%v", err)
	}

	// 如果没有返回 taskType，根据部门数量判断
	taskType := result.TaskType
	if taskType == 0 && len(req.DepartmentIDs) > 1 {
		taskType = 1
	}

	// 如果没有返回部门 ID 列表，使用请求中的部门 ID 列表
	departmentIDs := result.DepartmentIDs
	if len(departmentIDs) == 0 && len(req.DepartmentIDs) > 0 {
		departmentIDs = req.DepartmentIDs
	}

	// 如果没有返回负责人 ID 列表，使用请求中的负责人 ID 列表
	responsibleEmployeeIDs := result.ResponsibleEmployeeIDs
	if len(responsibleEmployeeIDs) == 0 && len(req.ResponsibleEmployeeIDs) > 0 {
		responsibleEmployeeIDs = req.ResponsibleEmployeeIDs
	}

	return PolishResult{
		OriginalTitle:          req.TaskTitle,
		OriginalDetail:         req.TaskDetail,
		PolishedTitle:          result.PolishedTitle,
		PolishedDetail:         result.PolishedDetail,
		TaskType:               taskType,
		TaskPriority:           result.TaskPriority,
		EstimatedDays:          result.EstimatedDays,
		DepartmentIDs:          departmentIDs,
		ResponsibleEmployeeIDs: responsibleEmployeeIDs,
		Improvements:           result.Improvements,
	}, nil
}
