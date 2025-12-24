package handover

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHandoverListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetHandoverListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHandoverListLogic {
	return &GetHandoverListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHandoverListLogic) GetHandoverList(req *types.HandoverListRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	validator := utils.NewValidator()
	page, pageSize, errs := validator.ValidatePageParams(req.Page, req.PageSize)
	if len(errs) > 0 {
		return utils.Response.ValidationError(errs[0]), nil
	}

	// 2. 获取当前用户ID
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 3. 获取员工信息
	employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, currentUserID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询员工失败: %v", err)
		return utils.Response.ValidationError("用户未绑定员工信息"), nil
	}

	// 4. 查询与当前员工相关的所有交接（作为发起人、接收人或审批人）
	// 使用员工主键ID（employee.Id）而不是工号（employee.EmployeeId）
	l.Logger.WithContext(l.ctx).Infof("查询交接列表: employeeId=%s, page=%d, pageSize=%d", employee.Id, page, pageSize)
	handovers, total, err := l.svcCtx.TaskHandoverModel.FindByEmployeeInvolved(l.ctx, employee.Id, page, pageSize)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询交接列表失败: employeeId=%s, error=%v", employee.Id, err)
		return utils.Response.ValidationError("查询交接列表失败: " + err.Error()), nil
	}
	l.Logger.WithContext(l.ctx).Infof("查询到 %d 条交接记录, 总数: %d", len(handovers), total)

	// 5. 状态过滤（如果指定了状态）
	var filteredHandovers = handovers
	if req.Status > 0 {
		filteredHandovers = nil
		for _, handover := range handovers {
			if handover.HandoverStatus == int64(req.Status) {
				filteredHandovers = append(filteredHandovers, handover)
			}
		}
		total = int64(len(filteredHandovers))
	}

	// 6. 转换为响应格式，包含更多详细信息
	var handoverInfos []interface{}
	for _, handover := range filteredHandovers {
		// 获取任务信息（离职申请的TaskId为空，不需要查询任务）
		taskTitle := ""
		if handover.TaskId != "" {
			if task, taskErr := l.svcCtx.TaskModel.FindOne(l.ctx, handover.TaskId); taskErr == nil {
				taskTitle = task.TaskTitle
			}
		} else {
			// 离职申请没有关联任务
			taskTitle = "离职审批"
		}

		// 获取发起人姓名
		fromEmployeeName := ""
		if fromEmp, fromErr := l.svcCtx.EmployeeModel.FindOne(l.ctx, handover.FromEmployeeId); fromErr == nil {
			fromEmployeeName = fromEmp.RealName
		}

		// 获取接收人姓名
		toEmployeeName := ""
		if toEmp, toErr := l.svcCtx.EmployeeModel.FindOne(l.ctx, handover.ToEmployeeId); toErr == nil {
			toEmployeeName = toEmp.RealName
		}

		// 获取审批人姓名
		approverName := ""
		if handover.ApproverId.Valid && handover.ApproverId.String != "" {
			if approver, approverErr := l.svcCtx.EmployeeModel.FindOne(l.ctx, handover.ApproverId.String); approverErr == nil {
				approverName = approver.RealName
			}
		}

		// 处理可空字段
		approverId := ""
		if handover.ApproverId.Valid {
			approverId = handover.ApproverId.String
		}
		handoverReason := ""
		if handover.HandoverReason.Valid {
			handoverReason = handover.HandoverReason.String
		}
		handoverNote := ""
		if handover.HandoverNote.Valid {
			handoverNote = handover.HandoverNote.String
		}

		handoverInfo := map[string]interface{}{
			"handoverId":       handover.HandoverId,
			"taskId":           handover.TaskId,
			"taskTitle":        taskTitle,
			"fromEmployeeId":   handover.FromEmployeeId,
			"fromEmployeeName": fromEmployeeName,
			"toEmployeeId":     handover.ToEmployeeId,
			"toEmployeeName":   toEmployeeName,
			"approverId":       approverId,
			"approverName":     approverName,
			"handoverReason":   handoverReason,
			"handoverNote":     handoverNote,
			"handoverStatus":   handover.HandoverStatus,
			"approvalType":     "handover", // 标记为交接审批
			"createTime":       handover.CreateTime.Format("2006-01-02 15:04:05"),
			"updateTime":       handover.UpdateTime.Format("2006-01-02 15:04:05"),
		}

		if handover.ApproveTime.Valid {
			handoverInfo["approveTime"] = handover.ApproveTime.Time.Format("2006-01-02 15:04:05")
		}

		handoverInfos = append(handoverInfos, handoverInfo)
	}

	// 7. 查询任务节点完成审批记录（作为审批人）
	nodeApprovals, nodeTotal, err := l.svcCtx.HandoverApprovalModel.FindTaskNodeApprovalsByApprover(l.ctx, employee.Id, page, pageSize)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询任务节点审批列表失败: employeeId=%s, error=%v", employee.Id, err)
	} else {
		l.Logger.WithContext(l.ctx).Infof("查询到 %d 条任务节点审批记录, 总数: %d", len(nodeApprovals), nodeTotal)

		// 转换任务节点审批记录
		for _, approval := range nodeApprovals {
			// 状态过滤
			if req.Status > 0 {
				// 将审批类型映射到交接状态: 0-待审批->1, 1-同意->2, 2-拒绝->3
				mappedStatus := approval.ApprovalType + 1
				if mappedStatus != int64(req.Status) {
					continue
				}
			}

			// 获取任务节点信息
			taskNodeId := ""
			if approval.TaskNodeId.Valid {
				taskNodeId = approval.TaskNodeId.String
			}

			nodeName := ""
			taskId := ""
			taskTitle := ""
			fromEmployeeId := ""
			fromEmployeeName := ""

			if taskNodeId != "" {
				if taskNode, nodeErr := l.svcCtx.TaskNodeModel.FindOne(l.ctx, taskNodeId); nodeErr == nil {
					nodeName = taskNode.NodeName
					taskId = taskNode.TaskId
					fromEmployeeId = taskNode.ExecutorId

					// 获取任务标题
					if task, taskErr := l.svcCtx.TaskModel.FindOne(l.ctx, taskNode.TaskId); taskErr == nil {
						taskTitle = task.TaskTitle
					}

					// 获取执行人姓名
					if fromEmployeeId != "" {
						if fromEmp, fromErr := l.svcCtx.EmployeeModel.FindOne(l.ctx, fromEmployeeId); fromErr == nil {
							fromEmployeeName = fromEmp.RealName
						}
					}
				}
			}

			// 将审批类型映射到交接状态: 0-待审批->1, 1-同意->2, 2-拒绝->3
			handoverStatus := approval.ApprovalType + 1

			nodeApprovalInfo := map[string]interface{}{
				"handoverId":       approval.ApprovalId, // 使用审批ID作为handoverId
				"taskId":           taskId,
				"taskNodeId":       taskNodeId,
				"taskTitle":        taskTitle + " - " + nodeName, // 任务标题 + 节点名称
				"fromEmployeeId":   fromEmployeeId,
				"fromEmployeeName": fromEmployeeName,
				"toEmployeeId":     approval.ApproverId,
				"toEmployeeName":   approval.ApproverName,
				"approverId":       approval.ApproverId,
				"approverName":     approval.ApproverName,
				"handoverReason":   "任务节点完成审批",
				"handoverNote":     "",
				"handoverStatus":   handoverStatus,
				"approvalType":     "node_completion", // 标记为任务节点完成审批
				"createTime":       approval.CreateTime.Format("2006-01-02 15:04:05"),
			}

			if approval.UpdateTime.Valid {
				nodeApprovalInfo["updateTime"] = approval.UpdateTime.Time.Format("2006-01-02 15:04:05")
			}

			handoverInfos = append(handoverInfos, nodeApprovalInfo)
		}

		// 更新总数
		total += nodeTotal
	}

	// 8. 构建分页响应
	converter := utils.NewConverter()
	pageResponse := converter.ToPageResponse(handoverInfos, int(total), page, pageSize)

	return utils.Response.Success(pageResponse), nil
}
