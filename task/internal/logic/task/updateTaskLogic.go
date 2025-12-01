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
		return utils.Response.BusinessError("任务ID不能为空"), nil
	}

	// 2. 获取当前用户ID
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 3. 获取任务信息
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, req.TaskID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("任务不存在"), nil
		}
		return nil, err
	}

	// 4. 验证用户是否有权限更新任务
	if taskInfo.TaskCreator != currentUserID {
		return utils.Response.BusinessError("无权限更新此任务"), nil
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
		deadline, err := time.Parse("2006-01-02", req.Deadline)
		if err != nil {
			return utils.Response.BusinessError("任务截止时间格式错误"), nil
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
		deadline, err := time.Parse("2006-01-02", req.Deadline)
		if err != nil {
			return utils.Response.BusinessError("任务截止时间格式错误"), nil
		}
		updatedTask.TaskDeadline = deadline
	}
	updatedTask.UpdateTime = time.Now()

	err = l.svcCtx.TaskModel.Update(l.ctx, &updatedTask)
	if err != nil {
		l.Logger.Errorf("更新任务失败: %v", err)
		return nil, err
	}

	// 8. 创建任务日志
	taskLog := &task.TaskLog{
		LogId:      utils.Common.GenerateID(),
		TaskId:     req.TaskID,
		LogType:    2, // 更新类型
		LogContent: fmt.Sprintf("任务信息已更新：%s", req.UpdateNote),
		EmployeeId: currentUserID,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.Errorf("创建任务日志失败: %v", err)
	}

	return utils.Response.Success(map[string]interface{}{
		"taskId":  req.TaskID,
		"message": "任务更新成功",
	}), nil
}
