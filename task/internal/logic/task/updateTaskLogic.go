package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTaskLogic {
	return &UpdateTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTaskLogic) UpdateTask(req *types.UpdateTaskRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.TaskID == "" {
		return utils.Response.BusinessError("task_id_required"), nil
	}

	// 2. 获取当前用户ID
	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}
	// 3. 获取任务信息
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, req.TaskID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("task_not_found"), nil
		}
		return nil, err
	}

	// 4. 验证用户是否有权限更新任务（任务创建者或任务负责人都可以）
	leaderId := ""
	if taskInfo.LeaderId.Valid {
		leaderId = taskInfo.LeaderId.String
	}
	if taskInfo.TaskCreator != employeeId && leaderId != employeeId {
		return utils.Response.BusinessError("task_update_denied"), nil
	}

	// 5. 验证任务状态
	if taskInfo.TaskStatus == 3 { // 已完成
		return utils.Response.BusinessError("task_already_completed"), nil
	}

	// 6. 构建更新数据
	updateData := make(map[string]interface{})

	if req.TaskTitle != "" {
		updateData["task_title"] = req.TaskTitle
	}
	if req.TaskDescription != "" {
		updateData["task_description"] = req.TaskDescription
	}
	if req.Status > 0 {
		updateData["task_status"] = req.Status
	}
	if req.Deadline != "" {
		// 支持多种日期格式
		deadline, parseErr := parseDeadline(req.Deadline)
		if parseErr != nil {
			return utils.Response.BusinessError("task_deadline_format"), nil
		}
		updateData["task_deadline"] = deadline
	}

	updateData["update_time"] = time.Now()

	// 7. 更新任务
	updatedTask := *taskInfo
	if req.TaskTitle != "" {
		updatedTask.TaskTitle = req.TaskTitle
	}
	if req.TaskDescription != "" {
		updatedTask.TaskDetail = req.TaskDescription
	}
	if req.Status > 0 {
		updatedTask.TaskStatus = int64(req.Status)
	}
	if req.Deadline != "" {
		deadline, parseErr := parseDeadline(req.Deadline)
		if parseErr != nil {
			return utils.Response.BusinessError("task_deadline_format"), nil
		}
		updatedTask.TaskDeadline = deadline
	}
	updatedTask.UpdateTime = time.Now()

	err = l.svcCtx.TaskModel.Update(l.ctx, &updatedTask)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("更新任务失败: %v", err)
		return nil, err
	}

	// 8. 创建任务日志
	taskLog := &task.TaskLog{
		LogId:      utils.Common.GenerateID(),
		TaskId:     req.TaskID,
		LogType:    2, // 更新类型
		LogContent: fmt.Sprintf("任务信息已更新：%s", req.UpdateNote),
		EmployeeId: employeeId,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务日志失败: %v", err)
	}

	return utils.Response.Success(map[string]interface{}{
		"taskId":  req.TaskID,
		"message": "任务更新成功",
	}), nil
}

// parseDeadline 解析多种格式的截止时间
func parseDeadline(dateStr string) (time.Time, error) {
	// 尝试多种日期格式
	formats := []string{
		time.RFC3339,           // 2006-01-02T15:04:05Z07:00
		"2006-01-02T15:04:05Z", // ISO 8601
		"2006-01-02T15:04:05",  // ISO 8601 without timezone
		"2006-01-02 15:04:05",  // 常见格式
		"2006-01-02",           // 纯日期
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析日期格式: %s", dateStr)
}
