// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"
	"fmt"
	"time"

	"task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTaskDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新任务详情（添加文件）
func NewUpdateTaskDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTaskDetailLogic {
	return &UpdateTaskDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTaskDetailLogic) UpdateTaskDetail(req *types.UpdateTaskDetailRequest) (resp *types.BaseResponse, err error) {
	// 先查看该员工是否合法
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}
	//遍历查看员工信息是否存在，如果不存在则返回错误
	employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, userID)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("员工信息不存在: %v", err)
		return utils.Response.BusinessError("员工信息不存在"), nil
	}
	if employee == nil {
		logx.WithContext(l.ctx).Errorf("员工信息不存在: %v", err)
		return utils.Response.BusinessError("员工信息不存在,请重新登录"), nil
	}
	//遍历查看任务信息是否存在，如果不存在则返回错误
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, req.TaskId)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("任务信息不存在: %v", err)
		return utils.Response.BusinessError("任务信息不存在"), nil
	}
	if taskInfo == nil {
		logx.WithContext(l.ctx).Errorf("任务信息不存在: %v", err)
		return utils.Response.BusinessError("任务信息不存在,请重新登录"), nil
	}
	for _, file := range req.File {
		l.svcCtx.TaskProjectDetailModel.Insert(l.ctx, &task.Task_project_detail{
			TaskID:   req.TaskId,
			FileID:   file.FileID,
			FileName: file.FileName,
			FileURL:  file.FileURL,
			FileType: file.FileType,
			FileSize: file.FileSize,
			CreateAt: time.Now(),
			UpdateAt: time.Now(),
		})
	}

	// 收集任务相关人员
	employeeIDSet := make(map[string]bool)
	// 查询任务所有节点的负责人和执行人
	nodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, req.TaskId)
	if err == nil {
		for _, node := range nodes {
			if node.LeaderId != "" {
				employeeIDSet[node.LeaderId] = true
			}
			if node.ExecutorId != "" {
				employeeIDSet[node.ExecutorId] = true
			}
		}
	}
	employeeIDs := make([]string, 0, len(employeeIDSet))
	for id := range employeeIDSet {
		employeeIDs = append(employeeIDs, id)
	}

	// 发布通知事件到消息队列
	if l.svcCtx.NotificationMQService != nil && len(employeeIDs) > 0 {
		notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
			svc.TaskUpdated,
			employeeIDs,
			req.TaskId,
			svc.NotificationEventOptions{TaskID: req.TaskId},
		)
		notificationEvent.Title = fmt.Sprintf("任务详情更新 - %s", taskInfo.TaskTitle)
		notificationEvent.Content = fmt.Sprintf("任务 %s 的详情已更新，新增了 %d 个文件附件", taskInfo.TaskTitle, len(req.File))
		notificationEvent.Priority = 1
		if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
			logx.WithContext(l.ctx).Errorf("发布任务更新通知事件失败: %v", err)
		}
	}

	// 发布邮件事件到消息队列
	if l.svcCtx.EmailMQService != nil {
		emailEvent := &svc.EmailEvent{
			EventType: svc.TaskUpdated,
			TaskID:    req.TaskId,
		}
		if err := l.svcCtx.EmailMQService.PublishEmailEvent(l.ctx, emailEvent); err != nil {
			logx.WithContext(l.ctx).Errorf("发布任务更新邮件事件失败: %v", err)
		}
	}

	return utils.Response.Success("任务详情更新成功"), nil
}
