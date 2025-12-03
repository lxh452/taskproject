package handover

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	taskModel "task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ApproveHandoverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApproveHandoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveHandoverLogic {
	return &ApproveHandoverLogic{
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

// ApproveHandover 上级审批交接（第二步）
func (l *ApproveHandoverLogic) ApproveHandover(req *types.ApproveHandoverRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.HandoverID == "" {
		return utils.Response.ValidationError("交接ID不能为空"), nil
	}

	// 2. 获取当前用户ID并查找员工信息
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 获取当前员工信息
	currentEmployee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, currentUserID)
	if err != nil {
		l.Logger.Errorf("获取当前员工信息失败: %v", err)
		return utils.Response.ValidationError("用户未绑定员工信息"), nil
	}
	currentEmployeeID := currentEmployee.Id
	approverName := currentEmployee.RealName

	// 3. 获取交接记录
	handover, err := l.svcCtx.TaskHandoverModel.FindOne(l.ctx, req.HandoverID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.ValidationError("交接记录不存在"), nil
		}
		return nil, err
	}

	// 4. 验证交接状态（只有待上级审批的才能审批）
	if handover.HandoverStatus != 1 {
		return utils.Response.ValidationError("只有待上级审批的交接才能进行审批"), nil
	}

	// 5. 验证是否有审批权限
	// 审批人可以是：指定的审批人、发起人的部门经理、或者有管理员权限的人
	hasApprovalPermission := false

	// 检查是否是指定的审批人
	if handover.ApproverId.Valid && handover.ApproverId.String == currentEmployeeID {
		hasApprovalPermission = true
	}

	// 如果没有指定审批人，检查是否是发起人的部门经理
	if !hasApprovalPermission {
		fromEmployee, fromErr := l.svcCtx.EmployeeModel.FindOne(l.ctx, handover.FromEmployeeId)
		if fromErr == nil && fromEmployee.DepartmentId.Valid && fromEmployee.DepartmentId.String != "" {
			dept, deptErr := l.svcCtx.DepartmentModel.FindOne(l.ctx, fromEmployee.DepartmentId.String)
			if deptErr == nil && dept.ManagerId.Valid && dept.ManagerId.String == currentEmployeeID {
				hasApprovalPermission = true
			}
		}
	}

	// 检查是否是接收人的部门经理
	if !hasApprovalPermission {
		toEmployee, toErr := l.svcCtx.EmployeeModel.FindOne(l.ctx, handover.ToEmployeeId)
		if toErr == nil && toEmployee.DepartmentId.Valid && toEmployee.DepartmentId.String != "" {
			dept, deptErr := l.svcCtx.DepartmentModel.FindOne(l.ctx, toEmployee.DepartmentId.String)
			if deptErr == nil && dept.ManagerId.Valid && dept.ManagerId.String == currentEmployeeID {
				hasApprovalPermission = true
			}
		}
	}

	if !hasApprovalPermission {
		return utils.Response.ValidationError("您没有审批此交接的权限，只有指定审批人或相关部门经理可以审批"), nil
	}

	// 6. 处理指定交接人（对于离职申请或ToEmployeeId为空的情况）
	// 如果审批通过且指定了交接人，更新ToEmployeeId
	if req.Approved == 1 && req.ToEmployeeID != "" {
		// 验证指定的交接人是否存在
		toEmployee, toErr := l.svcCtx.EmployeeModel.FindOne(l.ctx, req.ToEmployeeID)
		if toErr != nil {
			return utils.Response.ValidationError("指定的交接人不存在"), nil
		}
		if toEmployee.Status == 0 {
			return utils.Response.ValidationError("指定的交接人已离职，无法交接"), nil
		}
		// 更新交接人
		handover.ToEmployeeId = req.ToEmployeeID
		l.Logger.Infof("审批时指定交接人: %s -> %s", handover.FromEmployeeId, req.ToEmployeeID)
	}

	// 7. 更新交接状态
	var newStatus int64
	var statusText string
	var approvalType int64
	if req.Approved == 1 {
		// 如果ToEmployeeId仍然为空，不允许通过
		if handover.ToEmployeeId == "" {
			return utils.Response.ValidationError("审批通过前必须指定交接人"), nil
		}
		newStatus = 2 // 已通过
		statusText = "已通过"
		approvalType = 1 // 同意
	} else {
		newStatus = 3 // 已拒绝
		statusText = "已拒绝"
		approvalType = 2 // 拒绝
		// 如果拒绝时建议了交接人，记录在Comment中
		if req.ToEmployeeID != "" {
			suggestedEmployee, sugErr := l.svcCtx.EmployeeModel.FindOne(l.ctx, req.ToEmployeeID)
			if sugErr == nil {
				if req.Comment != "" {
					req.Comment = fmt.Sprintf("%s（建议交接给：%s）", req.Comment, suggestedEmployee.RealName)
				} else {
					req.Comment = fmt.Sprintf("建议交接给：%s", suggestedEmployee.RealName)
				}
			}
		}
	}

	// 更新状态和审批时间
	handover.HandoverStatus = newStatus
	handover.ApproveTime = sql.NullTime{Time: time.Now(), Valid: true}
	handover.UpdateTime = time.Now()
	err = l.svcCtx.TaskHandoverModel.Update(l.ctx, handover)
	if err != nil {
		return nil, err
	}

	// 8. 插入审批记录到数据库
	comment := req.Comment
	if comment == "" && approvalType == 1 {
		comment = "上级审批通过"
	}
	approvalRecord := &taskModel.HandoverApproval{
		ApprovalId:   utils.Common.GenerateIDWithPrefix("approval"),
		HandoverId:   req.HandoverID,
		ApprovalStep: 2, // 第二步：上级审批
		ApproverId:   currentEmployeeID,
		ApproverName: approverName,
		ApprovalType: approvalType,
		Comment:      sql.NullString{String: comment, Valid: comment != ""},
		CreateTime:   time.Now(),
	}
	_, err = l.svcCtx.HandoverApprovalModel.Insert(l.ctx, approvalRecord)
	if err != nil {
		l.Logger.Errorf("插入审批记录失败: %v", err)
	}

	// 9. 如果通过，更新任务和任务节点的相关人员
	if newStatus == 2 {
		if handover.TaskId != "" {
			// 普通任务交接：更新指定任务的相关人员
			l.Logger.Infof("开始更新任务相关人员: 从 %s 转移到 %s", handover.FromEmployeeId, handover.ToEmployeeId)

			// 9.1 更新任务的负责人
			taskInfo, taskErr := l.svcCtx.TaskModel.FindOne(l.ctx, handover.TaskId)
			if taskErr == nil {
				needUpdateTask := false

				// 检查并更新任务创建者
				if taskInfo.TaskCreator == handover.FromEmployeeId {
					taskInfo.TaskCreator = handover.ToEmployeeId
					needUpdateTask = true
					l.Logger.Infof("更新任务创建者: %s -> %s", handover.FromEmployeeId, handover.ToEmployeeId)
				}

				// 检查并更新任务负责人列表
				if taskInfo.ResponsibleEmployeeIds.Valid && taskInfo.ResponsibleEmployeeIds.String != "" {
					oldIds := taskInfo.ResponsibleEmployeeIds.String
					newIds := strings.ReplaceAll(oldIds, handover.FromEmployeeId, handover.ToEmployeeId)
					if oldIds != newIds {
						taskInfo.ResponsibleEmployeeIds = sql.NullString{String: newIds, Valid: true}
						needUpdateTask = true
						l.Logger.Infof("更新任务负责人: %s -> %s", oldIds, newIds)
					}
				}

				// 检查并更新任务节点员工列表
				if taskInfo.NodeEmployeeIds.Valid && taskInfo.NodeEmployeeIds.String != "" {
					oldIds := taskInfo.NodeEmployeeIds.String
					newIds := strings.ReplaceAll(oldIds, handover.FromEmployeeId, handover.ToEmployeeId)
					if oldIds != newIds {
						taskInfo.NodeEmployeeIds = sql.NullString{String: newIds, Valid: true}
						needUpdateTask = true
						l.Logger.Infof("更新任务节点员工: %s -> %s", oldIds, newIds)
					}
				}

				if needUpdateTask {
					taskInfo.UpdateTime = time.Now()
					if updateErr := l.svcCtx.TaskModel.Update(l.ctx, taskInfo); updateErr != nil {
						l.Logger.Errorf("更新任务信息失败: %v", updateErr)
					} else {
						l.Logger.Infof("任务信息更新成功")
					}
				}
			}

			// 9.2 更新任务节点的执行人和负责人
			nodes, nodeErr := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, handover.TaskId)
			if nodeErr != nil {
				l.Logger.Errorf("获取任务节点失败: %v", nodeErr)
			} else {
				for _, node := range nodes {
					needUpdateNode := false

					// 更新执行人
					if node.ExecutorId == handover.FromEmployeeId {
						node.ExecutorId = handover.ToEmployeeId
						needUpdateNode = true
						l.Logger.Infof("更新节点 %s 执行人: %s -> %s", node.TaskNodeId, handover.FromEmployeeId, handover.ToEmployeeId)
					}

					// 更新负责人
					if node.LeaderId == handover.FromEmployeeId {
						node.LeaderId = handover.ToEmployeeId
						needUpdateNode = true
						l.Logger.Infof("更新节点 %s 负责人: %s -> %s", node.TaskNodeId, handover.FromEmployeeId, handover.ToEmployeeId)
					}

					if needUpdateNode {
						node.UpdateTime = time.Now()
						if updateErr := l.svcCtx.TaskNodeModel.Update(l.ctx, node); updateErr != nil {
							l.Logger.Errorf("更新任务节点失败: %v", updateErr)
						}
					}
				}
			}

			l.Logger.Infof("任务相关人员更新完成")
		} else {
			// 离职申请：处理离职员工的所有任务节点交接
			l.Logger.Infof("开始处理离职员工任务交接: 从 %s 转移到 %s", handover.FromEmployeeId, handover.ToEmployeeId)

			// 9.3 查找离职员工负责的所有任务节点（作为执行人和负责人）
			executorNodes, _, executorErr := l.svcCtx.TaskNodeModel.FindByExecutor(l.ctx, handover.FromEmployeeId, 1, 1000)
			leaderNodes, _, leaderErr := l.svcCtx.TaskNodeModel.FindByLeader(l.ctx, handover.FromEmployeeId, 1, 1000)

			// 合并节点列表（去重）
			nodeMap := make(map[string]*taskModel.TaskNode)
			if executorErr == nil {
				for _, node := range executorNodes {
					nodeMap[node.TaskNodeId] = node
				}
			}
			if leaderErr == nil {
				for _, node := range leaderNodes {
					nodeMap[node.TaskNodeId] = node
				}
			}

			if len(nodeMap) > 0 {
				updatedCount := 0
				for _, node := range nodeMap {
					needUpdateNode := false

					// 更新执行人（支持多执行人，使用字符串替换）
					if node.ExecutorId != "" && strings.Contains(node.ExecutorId, handover.FromEmployeeId) {
						// 处理多执行人的情况
						executorIds := strings.Split(node.ExecutorId, ",")
						newExecutorIds := make([]string, 0, len(executorIds))
						for _, id := range executorIds {
							if strings.TrimSpace(id) != handover.FromEmployeeId {
								newExecutorIds = append(newExecutorIds, strings.TrimSpace(id))
							}
						}
						// 添加新的交接人
						newExecutorIds = append(newExecutorIds, handover.ToEmployeeId)
						node.ExecutorId = strings.Join(newExecutorIds, ",")
						needUpdateNode = true
						l.Logger.Infof("更新节点 %s 执行人: %s -> %s", node.TaskNodeId, node.ExecutorId, handover.ToEmployeeId)
					} else if node.ExecutorId == handover.FromEmployeeId {
						// 单执行人情况
						node.ExecutorId = handover.ToEmployeeId
						needUpdateNode = true
						l.Logger.Infof("更新节点 %s 执行人: %s -> %s", node.TaskNodeId, handover.FromEmployeeId, handover.ToEmployeeId)
					}

					// 更新负责人（支持多负责人）
					if node.LeaderId != "" && strings.Contains(node.LeaderId, handover.FromEmployeeId) {
						// 处理多负责人的情况
						leaderIds := strings.Split(node.LeaderId, ",")
						newLeaderIds := make([]string, 0, len(leaderIds))
						for _, id := range leaderIds {
							if strings.TrimSpace(id) != handover.FromEmployeeId {
								newLeaderIds = append(newLeaderIds, strings.TrimSpace(id))
							}
						}
						// 添加新的交接人
						newLeaderIds = append(newLeaderIds, handover.ToEmployeeId)
						node.LeaderId = strings.Join(newLeaderIds, ",")
						needUpdateNode = true
						l.Logger.Infof("更新节点 %s 负责人: %s -> %s", node.TaskNodeId, node.LeaderId, handover.ToEmployeeId)
					} else if node.LeaderId == handover.FromEmployeeId {
						// 单负责人情况
						node.LeaderId = handover.ToEmployeeId
						needUpdateNode = true
						l.Logger.Infof("更新节点 %s 负责人: %s -> %s", node.TaskNodeId, handover.FromEmployeeId, handover.ToEmployeeId)
					}

					if needUpdateNode {
						node.UpdateTime = time.Now()
						if updateErr := l.svcCtx.TaskNodeModel.Update(l.ctx, node); updateErr != nil {
							l.Logger.Errorf("更新任务节点失败: %v", updateErr)
						} else {
							updatedCount++
						}
					}
				}
				l.Logger.Infof("离职员工任务节点交接完成，共更新 %d 个节点", updatedCount)
			} else {
				l.Logger.Infof("离职员工没有需要交接的任务节点")
			}

			// 9.4 更新员工状态为离职
			updateData := map[string]interface{}{
				"status":     0, // 离职
				"leave_date": time.Now(),
			}
			err = l.svcCtx.EmployeeModel.SelectiveUpdate(l.ctx, handover.FromEmployeeId, updateData)
			if err != nil {
				l.Logger.Errorf("更新员工离职状态失败: %v", err)
			} else {
				l.Logger.Infof("员工 %s 状态已更新为离职", handover.FromEmployeeId)
			}
		}
	}

	// 10. 创建任务日志（离职申请可能没有任务，需要判断）
	if handover.TaskId != "" {
		taskLog := &taskModel.TaskLog{
			LogId:   utils.Common.GenerateID(),
			TaskId:  handover.TaskId,
			LogType: 6, // 交接审批
			LogContent: fmt.Sprintf("上级审批交接: %s -> %s, 审批结果: %s, 审批人: %s",
				handover.FromEmployeeId, handover.ToEmployeeId, statusText, approverName),
			EmployeeId: currentEmployeeID,
			CreateTime: time.Now(),
		}
		_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
		if err != nil {
			l.Logger.Errorf("创建任务日志失败: %v", err)
		}
	}

	// 10. 发送通知给发起人和接收人
	notifyEmployees := []string{handover.FromEmployeeId, handover.ToEmployeeId}
	l.Logger.Infof("准备发送审批结果通知, NotificationMQService=%v, notifyEmployees=%v", l.svcCtx.NotificationMQService != nil, notifyEmployees)
	if l.svcCtx.NotificationMQService != nil {
		notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
			svc.HandoverNotification,
			notifyEmployees,
			req.HandoverID,
			svc.NotificationEventOptions{TaskID: handover.TaskId},
		)
		notificationEvent.Title = "交接审批结果"
		if newStatus == 2 {
			notificationEvent.Content = "您的交接申请已通过上级审批，交接生效"
		} else {
			notificationEvent.Content = fmt.Sprintf("您的交接申请被上级拒绝，原因：%s", req.Comment)
		}
		notificationEvent.Priority = 2
		l.Logger.Infof("发布审批结果通知: title=%s, content=%s", notificationEvent.Title, notificationEvent.Content)
		if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(l.ctx, notificationEvent); err != nil {
			l.Logger.Errorf("发布通知事件失败: %v", err)
		} else {
			l.Logger.Infof("审批结果通知发布成功")
		}
	} else {
		l.Logger.Errorf("NotificationMQService 为空，无法发送审批结果通知")
	}

	// 发布邮件事件
	if l.svcCtx.EmailMQService != nil {
		emailEvent := &svc.EmailEvent{
			EventType:   svc.HandoverNotification,
			EmployeeIDs: notifyEmployees,
			RelatedID:   req.HandoverID,
		}
		if err := l.svcCtx.EmailMQService.PublishEmailEvent(l.ctx, emailEvent); err != nil {
			l.Logger.Errorf("发布邮件事件失败: %v", err)
		}
	}

	return utils.Response.Success(map[string]interface{}{
		"handoverId": req.HandoverID,
		"status":     newStatus,
		"statusText": statusText,
		"message":    fmt.Sprintf("交接审批完成，状态: %s", statusText),
	}), nil
}
