package utils

import (
	"database/sql"
	"time"

	"task_Project/model/company"
	"task_Project/model/task"
	"task_Project/model/user"
	"task_Project/model/user_auth"
	"task_Project/task/internal/types"
)

var Converter = NewConverter()

// Converter 数据转换工具类
type converter struct{}

// NewConverter 创建转换工具实例
func NewConverter() *converter {
	return &converter{}
}

// ToEmployeeInfo 将Employee模型转换为EmployeeInfo类型
func (c *converter) ToEmployeeInfo(employee *user.Employee) types.EmployeeInfo {
	return types.EmployeeInfo{
		ID:           employee.Id,
		UserID:       employee.UserId,
		CompanyID:    employee.CompanyId,
		DepartmentID: getStringValue(employee.DepartmentId),
		PositionID:   getStringValue(employee.PositionId),
		SupervisorID: getStringValue(employee.SupervisorId),
		EmployeeID:   employee.EmployeeId,
		RealName:     employee.RealName,
		WorkEmail:    getStringValue(employee.Email), // 使用Email字段
		WorkPhone:    getStringValue(employee.Phone), // 使用Phone字段
		Skills:       getStringValue(employee.Skills),
		RoleTags:     getStringValue(employee.RoleTags),
		HireDate:     formatNullTime(employee.HireDate),
		LeaveDate:    formatNullTime(employee.LeaveDate),
		Status:       int(employee.Status),
		//LastLoginIP:  "", // Employee模型中没有这个字段
		CreateTime: formatTime(&employee.CreateTime),
		UpdateTime: formatTime(&employee.UpdateTime),
	}
}

// ToEmployeeInfoList 将Employee模型列表转换为EmployeeInfo列表
func (c *converter) ToEmployeeInfoList(employees []*user.Employee) []types.EmployeeInfo {
	result := make([]types.EmployeeInfo, 0, len(employees))
	for _, employee := range employees {
		result = append(result, c.ToEmployeeInfo(employee))
	}
	return result
}

// ToCompanyInfo 将Company模型转换为CompanyInfo类型
func (c *converter) ToCompanyInfo(company *company.Company) types.CompanyInfo {
	return types.CompanyInfo{
		ID:   company.Id,
		Name: company.Name,
		//CompanyCode:     "", // Company模型中没有这个字段
		//LegalPerson:     "", // Company模型中没有这个字段
		//BusinessLicense: "", // Company模型中没有这个字段
		Phone:       getStringValue(company.Phone),
		Email:       getStringValue(company.Email),
		Address:     getStringValue(company.Address),
		Description: getStringValue(company.Description),
		Status:      int(company.Status),
		Owner:       company.Owner,
		CreateTime:  formatTime(&company.CreateTime),
		UpdateTime:  formatTime(&company.UpdateTime),
	}
}

// ToCompanyInfoList 将Company模型列表转换为CompanyInfo列表
func (c *converter) ToCompanyInfoList(companies []*company.Company) []types.CompanyInfo {
	result := make([]types.CompanyInfo, 0, len(companies))
	for _, company := range companies {
		result = append(result, c.ToCompanyInfo(company))
	}
	return result
}

// ToDepartmentInfo 将Department模型转换为DepartmentInfo类型
func (c *converter) ToDepartmentInfo(department *company.Department) types.DepartmentInfo {
	return types.DepartmentInfo{
		ID:             department.Id,
		CompanyID:      department.CompanyId,
		ParentID:       getStringValue(department.ParentId),
		DepartmentName: department.DepartmentName,
		DepartmentCode: getStringValue(department.DepartmentCode),
		ManagerID:      getStringValue(department.ManagerId),
		Description:    getStringValue(department.Description),
		Status:         int(department.Status),
		CreateTime:     formatTime(&department.CreateTime),
		UpdateTime:     formatTime(&department.UpdateTime),
	}
}

// ToDepartmentInfoList 将Department模型列表转换为DepartmentInfo列表
func (c *converter) ToDepartmentInfoList(departments []*company.Department) []types.DepartmentInfo {
	result := make([]types.DepartmentInfo, 0, len(departments))
	for _, department := range departments {
		result = append(result, c.ToDepartmentInfo(department))
	}
	return result
}

// ToPositionInfo 将Position模型转换为PositionInfo类型
func (c *converter) ToPositionInfo(position *company.Position) types.PositionInfo {
	return types.PositionInfo{
		ID:               position.Id,
		DepartmentID:     position.DepartmentId,
		PositionName:     position.PositionName,
		PositionCode:     getStringValue(position.PositionCode),
		PositionLevel:    int(position.PositionLevel),
		RequiredSkills:   getStringValue(position.RequiredSkills),
		JobDescription:   getStringValue(position.JobDescription),
		Responsibilities: getStringValue(position.Responsibilities),
		Requirements:     getStringValue(position.Requirements),
		SalaryRangeMin:   int(getFloat64Value(position.SalaryRangeMin)),
		SalaryRangeMax:   int(getFloat64Value(position.SalaryRangeMax)),
		IsManagement:     int(position.IsManagement),
		MaxEmployees:     int(position.MaxEmployees),
		CurrentEmployees: int(position.CurrentEmployees),
		Status:           int(position.Status),
		CreateTime:       formatTime(&position.CreateTime),
		UpdateTime:       formatTime(&position.UpdateTime),
	}
}

// ToPositionInfoList 将Position模型列表转换为PositionInfo列表
func (c *converter) ToPositionInfoList(positions []*company.Position) []types.PositionInfo {
	result := make([]types.PositionInfo, 0, len(positions))
	for _, position := range positions {
		result = append(result, c.ToPositionInfo(position))
	}
	return result
}

// ToTaskInfo 将Task模型转换为TaskInfo类型
func (c *converter) ToTaskInfo(task *task.Task) types.TaskInfo {
	return types.TaskInfo{
		ID:                     task.TaskId,
		TaskTitle:              task.TaskTitle,
		TaskDescription:        task.TaskDetail,
		TaskType:               c.getTaskTypeString(task.TaskType),
		Priority:               int(task.TaskPriority),
		Status:                 int(task.TaskStatus),
		CompanyID:              task.CompanyId,
		DepartmentID:           getStringValue(task.DepartmentIds),
		CreatorID:              task.TaskCreator,
		LeaderId:               getStringValue(task.LeaderId),
		ResponsibleEmployeeIds: getStringValue(task.ResponsibleEmployeeIds),
		StartTime:              formatTime(&task.TaskStartTime),
		Deadline:               formatTime(&task.TaskDeadline),
		EstimatedHours:         0,                      // Task模型中没有这个字段，需要从TaskNode计算
		ActualHours:            0,                      // Task模型中没有这个字段，需要从TaskNode计算
		Progress:               int(task.TaskProgress), // 从数据库读取任务整体进度
		CreateTime:             formatTime(&task.CreateTime),
		UpdateTime:             formatTime(&task.UpdateTime),
	}
}

// ToTaskInfoList 将Task模型列表转换为TaskInfo列表
func (c *converter) ToTaskInfoList(tasks []*task.Task) []types.TaskInfo {
	result := make([]types.TaskInfo, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, c.ToTaskInfo(task))
	}
	return result
}

// ToTaskNodeInfo 将TaskNode模型转换为TaskNodeInfo类型
func (c *converter) ToTaskNodeInfo(taskNode *task.TaskNode) types.TaskNodeInfo {
	return types.TaskNodeInfo{
		ID:                taskNode.TaskNodeId,
		TaskID:            taskNode.TaskId,
		NodeName:          taskNode.NodeName,
		NodeDetail:        getStringValue(taskNode.NodeDetail),
		NodeType:          "任务节点", // 默认类型，可以根据需要扩展
		Status:            int(taskNode.NodeStatus),
		DepartmentID:      taskNode.DepartmentId,
		LeaderID:          taskNode.LeaderId,
		ExecutorID:        taskNode.ExecutorId,
		NodeDeadline:      formatTime(&taskNode.NodeDeadline),
		EstimatedHours:    int(taskNode.EstimatedDays * 8),               // 将天数转换为小时
		ActualHours:       int(c.getInt64Value(taskNode.ActualDays) * 8), // 将天数转换为小时
		Progress:          int(taskNode.Progress),
		PrerequisiteNodes: "", // TaskNode模型中没有这个字段
		CreateTime:        formatTime(&taskNode.CreateTime),
		UpdateTime:        formatTime(&taskNode.UpdateTime),
	}
}

// ToTaskNodeInfoList 将TaskNode模型列表转换为TaskNodeInfo列表
func (c *converter) ToTaskNodeInfoList(taskNodes []*task.TaskNode) []types.TaskNodeInfo {
	result := make([]types.TaskNodeInfo, 0, len(taskNodes))
	for _, taskNode := range taskNodes {
		result = append(result, c.ToTaskNodeInfo(taskNode))
	}
	return result
}

// ToTaskLogInfo 将TaskLog模型转换为TaskLogInfo类型
func (c *converter) ToTaskLogInfo(taskLog *task.TaskLog) types.TaskLogInfo {
	return types.TaskLogInfo{
		ID:         taskLog.LogId,
		TaskID:     taskLog.TaskId,
		TaskNodeID: getStringValue(taskLog.TaskNodeId),
		LogType:    c.getLogTypeString(taskLog.LogType),
		LogContent: taskLog.LogContent,
		OperatorID: taskLog.EmployeeId,
		CreateTime: formatTime(&taskLog.CreateTime),
	}
}

// ToTaskLogInfoList 将TaskLog模型列表转换为TaskLogInfo列表
func (c *converter) ToTaskLogInfoList(taskLogs []*task.TaskLog) []types.TaskLogInfo {
	result := make([]types.TaskLogInfo, 0, len(taskLogs))
	for _, taskLog := range taskLogs {
		result = append(result, c.ToTaskLogInfo(taskLog))
	}
	return result
}

// ToNotificationInfo 将Notification模型转换为NotificationInfo类型
func (c *converter) ToNotificationInfo(notification *user_auth.Notification) types.NotificationInfo {
	return types.NotificationInfo{
		ID:          notification.Id,
		EmployeeID:  notification.EmployeeId,
		Title:       notification.Title,
		Content:     notification.Content,
		Type:        int(notification.Type),
		Priority:    int(notification.Priority),
		IsRead:      int(notification.IsRead),
		SenderID:    "", // Notification模型中没有这个字段
		RelatedID:   getStringValue(notification.RelatedId),
		RelatedType: getStringValue(notification.RelatedType),
		ReadTime:    formatNullTime(notification.ReadTime),
		CreateTime:  formatTime(&notification.CreateTime),
		UpdateTime:  formatTime(&notification.UpdateTime),
	}
}

// ToNotificationInfoList 将Notification模型列表转换为NotificationInfo列表
func (c *converter) ToNotificationInfoList(notifications []*user_auth.Notification) []types.NotificationInfo {
	result := make([]types.NotificationInfo, 0, len(notifications))
	for _, notification := range notifications {
		result = append(result, c.ToNotificationInfo(notification))
	}
	return result
}

// ToTaskDetailInfo 将Task模型转换为TaskDetailInfo类型
func (c *converter) ToTaskDetailInfo(task *task.Task) types.TaskDetailInfo {
	return types.TaskDetailInfo{
		TaskInfo: c.ToTaskInfo(task),
		// Nodes:    []types.TaskNodeInfo{}, // 需要单独查询节点信息
		// Logs:     []types.TaskLogInfo{},  // 需要单独查询日志信息
	}
}

// getInt64Value 获取int64值
func (c *converter) getInt64Value(s sql.NullInt64) int64 {
	if !s.Valid {
		return 0
	}
	return s.Int64
}

// getTaskTypeString 将任务类型转换为字符串
func (c *converter) getTaskTypeString(taskType int64) string {
	switch taskType {
	case 0:
		return "单部门任务"
	case 1:
		return "跨部门任务"
	default:
		return "未知类型"
	}
}

// getLogTypeString 将日志类型转换为字符串
func (c *converter) getLogTypeString(logType int64) string {
	switch logType {
	case 0:
		return "创建"
	case 1:
		return "更新"
	case 2:
		return "完成"
	case 3:
		return "交接"
	case 4:
		return "评论"
	default:
		return "未知"
	}
}

// ToPageResponse 构建分页响应
func (c *converter) ToPageResponse(list interface{}, total int, page, pageSize int) types.PageResp {
	return types.PageResp{
		Total: total,
		List:  list,
	}
}

// getStringValue 获取字符串值
func getStringValue(s sql.NullString) string {
	if !s.Valid {
		return ""
	}
	return s.String
}

// formatTime 格式化时间
func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// formatNullTime 格式化空时间
func formatNullTime(t sql.NullTime) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format("2006-01-02 15:04:05")
}

// getFloat64Value 获取浮点数值
func getFloat64Value(f sql.NullFloat64) float64 {
	if !f.Valid {
		return 0
	}
	return f.Float64
}
