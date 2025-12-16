// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTaskLogic {
	return &CreateTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTaskLogic) CreateTask(req *types.CreateTaskRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.TaskTitle == "" {
		return utils.Response.BusinessError("任务标题不能为空"), nil
	}
	if req.CompanyID == "" {
		return utils.Response.BusinessError("公司ID不能为空"), nil
	}
	if req.TaskDeadline == "" {
		return utils.Response.BusinessError("任务截止时间不能为空"), nil
	}

	// 2. 解析截止时间
	deadline, err := time.Parse("2006-01-02", req.TaskDeadline)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("输出错误", err.Error())
		return utils.Response.BusinessError("任务截止时间格式错误"), nil
	}

	// 3. 验证公司是否存在
	_, err = l.svcCtx.CompanyModel.FindOne(l.ctx, req.CompanyID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("输出错误", err.Error())
		return utils.Response.BusinessError("公司不存在"), nil
	}
	// 1. 从上下文获取当前员工ID
	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}

	// 4. 验证用户（如果有负责人和节点负责人）
	// 这里需要根据实际需求来确定是否需要验证
	// todo 假设：公司管理员可以创建任务，不需要在负责人列表中

	// 5. 创建任务ID
	taskID := utils.Common.GenerateIDWithPrefix("task")

	//多部门进行拼接
	var departmentIds, nodeEmployeeIDs, responsibleEmployeeIds, attachmentURL string
	if len(req.DepartmentIDs) > 0 {
		departmentIds = strings.Join(req.DepartmentIDs, ",")
	}
	if len(req.NodeEmployeeIDs) > 0 {
		nodeEmployeeIDs = strings.Join(req.NodeEmployeeIDs, ",")
	}
	if len(req.ResponsibleEmployeeIDs) > 0 {
		responsibleEmployeeIds = strings.Join(req.ResponsibleEmployeeIDs, ",")
	}
	if len(req.AttachmentURL) > 0 {
		attachmentURL = strings.Join(req.AttachmentURL, ",")
	}

	// 6. 创建任务
	newTask := &task.Task{
		TaskId:                 taskID,
		CompanyId:              req.CompanyID,
		TaskTitle:              req.TaskTitle,
		TaskDetail:             req.TaskDetail,
		TaskPriority:           int64(req.TaskPriority),
		TaskType:               int64(req.TaskType), // 1: 单部门, 2: 跨部门
		TaskStatus:             0,                   // 待开始
		TaskProgress:           0,                   // 初始进度为0
		TotalNodes:             0,                   // 初始节点数为0
		CompletedNodes:         0,                   // 初始已完成节点数为0
		TotalNodeCount:         0,                   // 初始总节点数为0
		CompletedNodeCount:     0,                   // 初始已完成节点数为0
		TaskCreator:            employeeId,
		TaskStartTime:          time.Now(),
		TaskDeadline:           deadline,
		ResponsibleEmployeeIds: utils.Common.ToSqlNullString(responsibleEmployeeIds),
		NodeEmployeeIds:        utils.Common.ToSqlNullString(nodeEmployeeIDs),
		DepartmentIds:          utils.Common.ToSqlNullString(departmentIds),
		AttachmentUrl:          utils.Common.ToSqlNullString(attachmentURL),
		LeaderId:               utils.Common.ToSqlNullString(employeeId),
	}

	_, err = l.svcCtx.TaskModel.Insert(l.ctx, newTask)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务失败: %v", err)
		return nil, err
	}

	// 这里进行通知（通过消息队列）
	content := fmt.Sprintf("您现在为%s:%s任务的节点负责人，请登录系统进行查看，如无误，请尽快安排人手进行处理", taskID, newTask.TaskTitle)

	// 创建系统通知（通过消息队列）- 通知节点负责人
	if l.svcCtx.NotificationMQService != nil && len(req.NodeEmployeeIDs) > 0 {
		notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
			svc.TaskCreated,
			req.NodeEmployeeIDs,
			taskID,
			svc.NotificationEventOptions{TaskID: taskID},
		)
		notificationEvent.Title = fmt.Sprintf("新任务创建 - %s", req.TaskTitle)
		notificationEvent.Content = content
		notificationEvent.Priority = req.TaskType
		if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
			l.Logger.WithContext(l.ctx).Errorf("发布任务创建通知事件失败: %v", err)
		}
	}

	// 发送邮件给节点负责人（通过消息队列）
	if l.svcCtx.EmailMQService != nil && len(req.NodeEmployeeIDs) > 0 {
		l.Logger.WithContext(l.ctx).Infof("[CreateTask] EmailMQService is available, preparing to send emails to node employees")
		emails := []string{}
		for _, employeeID := range req.NodeEmployeeIDs {
			emp, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, employeeID)
			if err == nil && emp.Email.Valid && emp.Email.String != "" {
				emails = append(emails, emp.Email.String)
				l.Logger.WithContext(l.ctx).Infof("[CreateTask] Found employee email: employeeId=%s, email=%s", employeeID, emp.Email.String)
			} else {
				l.Logger.WithContext(l.ctx).Infof("[CreateTask] Employee email not found or invalid: employeeId=%s, error=%v", employeeID, err)
			}
		}
		if len(emails) > 0 {
			l.Logger.WithContext(l.ctx).Infof("[CreateTask] Sending email to %d recipients: %v", len(emails), emails)
			emailEvent := &svc.EmailEvent{
				EventType: svc.TaskCreated,
				To:        emails,
				Subject:   fmt.Sprintf("新任务创建 - %s", req.TaskTitle),
				Body:      content,
				IsHTML:    true,
				TaskID:    taskID,
			}
			if err := l.svcCtx.EmailMQService.PublishEmailEvent(l.ctx, emailEvent); err != nil {
				l.Logger.WithContext(l.ctx).Errorf("[CreateTask] Failed to publish email event: error=%v, taskId=%s, emails=%v",
					err, taskID, emails)
			} else {
				l.Logger.WithContext(l.ctx).Infof("[CreateTask] Email event published successfully: taskId=%s, emails=%v", taskID, emails)
			}
		} else {
			l.Logger.WithContext(l.ctx).Infof("[CreateTask] No valid email addresses found for node employees: nodeEmployeeIDs=%v", req.NodeEmployeeIDs)
		}
	} else {
		l.Logger.WithContext(l.ctx).Infof("[CreateTask] EmailMQService is not available or no node employees, email notification will be skipped")
	}

	// 7. 创建任务日志
	taskLog := &task.TaskLog{
		LogId:      utils.NewCommon().GenerateIDWithPrefix("task_log"),
		TaskId:     taskID,
		LogType:    1, // 创建类型
		LogContent: "创建任务: " + req.TaskTitle,
		EmployeeId: employeeId,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务日志失败: %v", err)
	}

	// 8. 如果是跨部门任务（TaskType=2），需要通知相关部门负责人
	if req.TaskType == 2 && len(req.DepartmentIDs) > 0 {
		// 这里可以发送通知给相关部门负责人
		// 通知逻辑已经在notification.go中实现
	}
	return utils.Response.Success(map[string]interface{}{
		"taskId":  taskID,
		"message": "任务创建成功",
	}), nil
}
