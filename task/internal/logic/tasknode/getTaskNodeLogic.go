package tasknode

import (
	"context"
	"database/sql"
	"errors"

	"task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// getStringValue 从sql.NullString获取字符串值
func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

type GetTaskNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTaskNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskNodeLogic {
	return &GetTaskNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTaskNodeLogic) GetTaskNode(req *types.GetTaskNodeRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.TaskNodeID == "" {
		return utils.Response.BusinessError("任务节点ID不能为空"), nil
	}

	// 2. 获取当前用户ID
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 3. 获取任务节点详情
	taskNode, err := l.svcCtx.TaskNodeModel.FindOne(l.ctx, req.TaskNodeID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("任务节点不存在"), nil
		}
		l.Logger.WithContext(l.ctx).Errorf("获取任务节点详情失败: %v", err)
		return nil, err
	}

	// 4. 验证用户权限（只有负责人、执行人或任务创建者可以查看）
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, taskNode.TaskId)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("获取任务信息失败: %v", err)
		return nil, err
	}

	hasPermission := false
	if taskNode.LeaderId == currentUserID ||
		taskNode.ExecutorId == currentUserID ||
		taskInfo.TaskCreator == currentUserID {
		hasPermission = true
	}

	// 检查是否是部门负责人 或者是执行人
	if !hasPermission {
		employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, currentUserID)
		if err == nil {
			department, err := l.svcCtx.DepartmentModel.FindOne(l.ctx, employee.DepartmentId.String)
			if err == nil && department.ManagerId.String == currentUserID {
				hasPermission = true
			}
		}
	}

	if !hasPermission {
		return utils.Response.BusinessError("无权限查看此任务节点"), nil
	}

	// 5. 获取该节点的审批列表
	approvals, err := l.svcCtx.TaskNodeCompletionApprovalModel.FindByTaskNodeId(l.ctx, req.TaskNodeID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("获取审批列表失败: %v", err)
		// 审批列表获取失败不影响节点详情返回，使用空列表
		approvals = []*task.TaskNodeCompletionApproval{}
	}

	// 6. 转换为响应格式
	converter := utils.NewConverter()
	taskNodeInfo := converter.ToTaskNodeInfo(taskNode)

	// 7. 将审批列表转换为响应格式
	approvalList := make([]map[string]interface{}, 0, len(approvals))
	for _, approval := range approvals {
		approvalList = append(approvalList, map[string]interface{}{
			"approvalId":   approval.ApprovalId,
			"taskNodeId":   approval.TaskNodeId,
			"approverId":   approval.ApproverId,
			"approverName": approval.ApproverName,
			"approvalType": approval.ApprovalType, // 0-待审批 1-同意 2-拒绝
			"comment":      getStringValue(approval.Comment),
			"createTime":   approval.CreateTime.Format("2006-01-02 15:04:05"),
			"updateTime":   approval.UpdateTime.Format("2006-01-02 15:04:05"),
		})
	}

	// 8. 将审批列表添加到响应中
	responseData := map[string]interface{}{
		"taskNode":      taskNodeInfo,
		"approvals":     approvalList,
		"approvalCount": len(approvalList),
	}

	return utils.Response.Success(responseData), nil
}
