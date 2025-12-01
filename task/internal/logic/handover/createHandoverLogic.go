package handover

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
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateHandoverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateHandoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateHandoverLogic {
	return &CreateHandoverLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 交接状态定义:
// 0 = 待接收人确认
// 1 = 待上级审批
// 2 = 已通过
// 3 = 已拒绝
// 4 = 已完成

func (l *CreateHandoverLogic) CreateHandover(req *types.CreateHandoverRequest) (resp *types.BaseResponse, err error) {
	// 打印请求参数用于调试
	l.Logger.Infof("创建交接请求参数: TaskID=%s, FromEmployeeID=%s, ToEmployeeID=%s", req.TaskID, req.FromEmployeeID, req.ToEmployeeID)

	// 1. 参数验证
	if req.TaskID == "" {
		return utils.Response.ValidationError("任务ID不能为空"), nil
	}
	if req.FromEmployeeID == "" {
		return utils.Response.ValidationError("发起人ID不能为空"), nil
	}
	if req.ToEmployeeID == "" {
		return utils.Response.ValidationError("接收人ID不能为空"), nil
	}
	if req.HandoverReason == "" {
		return utils.Response.ValidationError("交接原因不能为空"), nil
	}

	// 2. 验证任务是否存在
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, req.TaskID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.ValidationError("任务不存在"), nil
		}
		return nil, err
	}

	// 3. 验证任务状态（只有进行中的任务才能交接）
	// 任务状态：0-未开始，1-进行中，2-已完成，3-逾期完成
	// 如果任务状态是0（未开始），检查是否已到开始日期
	isInProgress := taskInfo.TaskStatus == 1
	if taskInfo.TaskStatus == 0 && !taskInfo.TaskStartTime.IsZero() {
		// 如果已到开始日期，视为进行中
		if !time.Now().Before(taskInfo.TaskStartTime) {
			isInProgress = true
		}
	}
	if !isInProgress {
		return utils.Response.ValidationError("只有进行中的任务才能交接"), nil
	}

	// 4. 验证发起人是否有权发起交接（任务创建者、负责人或节点执行人）
	nodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, req.TaskID)
	if err != nil {
		return nil, err
	}

	// 检查是否是任务创建者
	isCreator := taskInfo.TaskCreator == req.FromEmployeeID

	// 检查是否是任务负责人（通过responsible_employee_ids字段）
	isResponsible := false
	if taskInfo.ResponsibleEmployeeIds.Valid && taskInfo.ResponsibleEmployeeIds.String != "" {
		responsibleIds := strings.Split(taskInfo.ResponsibleEmployeeIds.String, ",")
		for _, id := range responsibleIds {
			if strings.TrimSpace(id) == req.FromEmployeeID {
				isResponsible = true
				break
			}
		}
	}

	// 检查是否是进行中节点的执行人
	isExecutor := false
	for _, node := range nodes {
		if node.ExecutorId == req.FromEmployeeID && node.NodeStatus == 2 {
			isExecutor = true
			break
		}
	}

	l.Logger.Infof("权限检查: isCreator=%v, isResponsible=%v, isExecutor=%v, TaskCreator=%s", isCreator, isResponsible, isExecutor, taskInfo.TaskCreator)
	if !isCreator && !isResponsible && !isExecutor {
		return utils.Response.ValidationError("只有任务的创建者、负责人或执行人才能发起交接"), nil
	}

	// 5. 验证接收人是否存在且在职
	l.Logger.Infof("查找接收人: ToEmployeeID=%s", req.ToEmployeeID)
	toEmployee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, req.ToEmployeeID)
	if err != nil {
		l.Logger.Errorf("查找接收人失败: %v", err)
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.ValidationError("接收人不存在"), nil
		}
		return nil, err
	}
	l.Logger.Infof("找到接收人: ID=%s, Name=%s, Status=%d", toEmployee.Id, toEmployee.RealName, toEmployee.Status)
	if toEmployee.Status != 1 {
		return utils.Response.ValidationError("接收人不在职，无法接收任务"), nil
	}

	// 6. 检查是否已有待处理的交接请求
	handovers, err := l.svcCtx.TaskHandoverModel.FindByTaskID(l.ctx, req.TaskID)
	if err == nil {
		for _, h := range handovers {
			// 状态0或1都是待处理状态
			if (h.HandoverStatus == 0 || h.HandoverStatus == 1) && h.ToEmployeeId == req.ToEmployeeID {
				return utils.Response.ValidationError("该任务对该接收人已有待处理的交接请求"), nil
			}
		}
	}

	// 7. 获取发起人信息，找到上级作为审批人
	fromEmployee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, req.FromEmployeeID)
	if err != nil {
		return nil, err
	}

	// 获取审批人（优先使用传入的，否则查找部门经理）
	var approverID string
	if req.ApproverID != "" {
		approverID = req.ApproverID
	} else if fromEmployee.DepartmentId.Valid && fromEmployee.DepartmentId.String != "" {
		// 查找部门经理作为审批人
		dept, deptErr := l.svcCtx.DepartmentModel.FindOne(l.ctx, fromEmployee.DepartmentId.String)
		if deptErr == nil && dept.ManagerId.Valid && dept.ManagerId.String != "" {
			approverID = dept.ManagerId.String
		}
	}

	// 8. 创建交接记录，状态为0（待接收人确认）
	handoverID := utils.Common.GenerateIDWithPrefix("handover")
	newHandover := &task.TaskHandover{
		HandoverId:     handoverID,
		TaskId:         req.TaskID,
		FromEmployeeId: req.FromEmployeeID,
		ToEmployeeId:   req.ToEmployeeID,
		HandoverReason: utils.Common.ToSqlNullString(req.HandoverReason),
		HandoverNote:   utils.Common.ToSqlNullString(req.HandoverNote),
		HandoverStatus: 0, // 待接收人确认
		ApproverId:     utils.Common.ToSqlNullString(approverID),
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}

	_, err = l.svcCtx.TaskHandoverModel.Insert(l.ctx, newHandover)
	if err != nil {
		l.Logger.Errorf("创建交接记录失败: %v", err)
		return nil, err
	}

	// 9. 创建任务日志
	taskLog := &task.TaskLog{
		LogId:      utils.Common.GenerateID(),
		TaskId:     req.TaskID,
		LogType:    5, // 交接请求
		LogContent: fmt.Sprintf("发起任务交接: %s -> %s, 原因: %s", req.FromEmployeeID, req.ToEmployeeID, req.HandoverReason),
		EmployeeId: req.FromEmployeeID,
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.Errorf("创建任务日志失败: %v", err)
	}

	// 10. 发送通知给接收人（通过消息队列）
	if l.svcCtx.NotificationMQService != nil {
		notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
			svc.HandoverNotification,
			[]string{req.ToEmployeeID},
			handoverID,
			svc.NotificationEventOptions{TaskID: req.TaskID},
		)
		notificationEvent.Title = "任务交接请求"
		notificationEvent.Content = fmt.Sprintf("您收到了一个任务交接请求：%s，原因：%s，请确认是否接收", taskInfo.TaskTitle, req.HandoverReason)
		notificationEvent.Priority = 2
		if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
			l.Logger.Errorf("发布交接通知事件失败: %v", err)
		}
	}

	// 发布邮件事件
	if l.svcCtx.EmailMQService != nil {
		emailEvent := &svc.EmailEvent{
			EventType:   svc.HandoverNotification,
			EmployeeIDs: []string{req.ToEmployeeID},
			RelatedID:   handoverID,
		}
		if err := l.svcCtx.EmailMQService.PublishEmailEvent(l.ctx, emailEvent); err != nil {
			l.Logger.Errorf("发布交接邮件事件失败: %v", err)
		}
	}

	return utils.Response.Success(map[string]interface{}{
		"handoverId": handoverID,
		"approverId": approverID,
		"message":    "交接请求已创建，等待接收人确认",
	}), nil
}
