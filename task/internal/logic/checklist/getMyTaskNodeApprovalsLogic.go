package checklist

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyTaskNodeApprovalsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyTaskNodeApprovalsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyTaskNodeApprovalsLogic {
	return &GetMyTaskNodeApprovalsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyTaskNodeApprovalsLogic) GetMyTaskNodeApprovals(req *types.PageReq) (resp *types.BaseResponse, err error) {
	// 1. 获取当前用户ID
	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return utils.Response.BusinessError("employee_not_found"), nil
	}

	// 2. 设置分页参数
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 3. 查询任务节点完成审批列表（只返回待审批状态）
	approvals, total, err := l.svcCtx.HandoverApprovalModel.FindTaskNodeApprovalsByApprover(l.ctx, employeeId, int(page), int(pageSize))
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询任务节点完成审批失败: %v", err)
		return nil, err
	}

	// 4. 构建返回数据
	list := make([]map[string]interface{}, 0, len(approvals))
	for _, approval := range approvals {
		// 获取任务节点信息
		var taskNodeName, taskTitle string
		var taskId string
		var nodeProgress int64
		var nodeStatus int64
		var nodeStatusText string
		if approval.TaskNodeId.Valid && approval.TaskNodeId.String != "" {
			taskNode, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, approval.TaskNodeId.String)
			if err == nil {
				taskNodeName = taskNode.NodeName
				taskId = taskNode.TaskId
				nodeProgress = taskNode.Progress
				nodeStatus = taskNode.NodeStatus
				// 节点状态文本：0-未开始，1-进行中，2-已完成，3-已逾期
				switch nodeStatus {
				case 0:
					nodeStatusText = "未开始"
				case 1:
					nodeStatusText = "进行中"
				case 2:
					nodeStatusText = "已完成"
				case 3:
					nodeStatusText = "已逾期"
				default:
					nodeStatusText = "未知"
				}

				// 获取任务信息
				task, err := l.svcCtx.TaskModel.FindOne(l.ctx, taskNode.TaskId)
				if err == nil {
					taskTitle = task.TaskTitle
				}
			}
		}

		// 获取提交人信息
		var submitterName string
		var submitterId string
		if approval.TaskNodeId.Valid && approval.TaskNodeId.String != "" {
			taskNode, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, approval.TaskNodeId.String)
			if err == nil && taskNode.ExecutorId != "" {
				// 取第一个执行人作为提交人
				executorIds := taskNode.ExecutorId
				if idx := len(executorIds); idx > 0 {
					firstExecutorId := executorIds
					if commaIdx := findFirstComma(executorIds); commaIdx > 0 {
						firstExecutorId = executorIds[:commaIdx]
					}
					submitterId = firstExecutorId
					executor, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, firstExecutorId)
					if err == nil {
						submitterName = executor.RealName
					}
				}
			}
		}

		// 审批状态文本：0-待审批，1-已通过，2-已拒绝
		var approvalStatusText string
		switch approval.ApprovalType {
		case 0:
			approvalStatusText = "待审批"
		case 1:
			approvalStatusText = "已通过"
		case 2:
			approvalStatusText = "已拒绝"
		default:
			approvalStatusText = "未知"
		}

		item := map[string]interface{}{
			"id":                 approval.ApprovalId,
			"approvalId":         approval.ApprovalId,
			"taskNodeId":         approval.TaskNodeId.String,
			"taskNodeName":       taskNodeName,
			"taskId":             taskId,
			"taskTitle":          taskTitle,
			"submitterId":        submitterId,
			"submitterName":      submitterName,
			"approverId":         approval.ApproverId,
			"approverName":       approval.ApproverName,
			"approvalType":       approval.ApprovalType,
			"approvalStatusText": approvalStatusText,
			"nodeProgress":       nodeProgress,
			"nodeStatus":         nodeStatus,
			"nodeStatusText":     nodeStatusText,
			"comment":            approval.Comment.String,
			"createTime":         approval.CreateTime.Format("2006-01-02 15:04:05"),
		}
		if approval.UpdateTime.Valid {
			item["updateTime"] = approval.UpdateTime.Time.Format("2006-01-02 15:04:05")
		}
		list = append(list, item)
	}

	return utils.Response.Success(map[string]interface{}{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}), nil
}

// findFirstComma 查找字符串中第一个逗号的位置
func findFirstComma(s string) int {
	for i, c := range s {
		if c == ',' {
			return i
		}
	}
	return -1
}
