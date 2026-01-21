package task

import (
	"context"
	"errors"
	"strings"

	taskmodel "task_Project/model/task"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTaskLogic {
	return &GetTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTaskLogic) GetTask(req *types.GetTaskRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.TaskID == "" {
		return utils.Response.BusinessError("task_id_required"), nil
	}

	// 获取当前用户员工ID
	employeeId, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	if !ok || employeeId == "" {
		return nil, errors.New("获取员工信息失败，请重新登录后再试")
	}

	// 2. 获取任务信息
	taskInfo, err := l.svcCtx.TaskModel.FindOne(l.ctx, req.TaskID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("task_not_found"), nil
		}
		return nil, err
	}

	// 3. 获取任务节点
	taskNodes, err := l.svcCtx.TaskNodeModel.FindByTaskID(l.ctx, req.TaskID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("获取任务节点失败: %v", err)
		taskNodes = []*taskmodel.TaskNode{} // 设置为空数组，不影响主流程
	}

	// 4. 判断用户角色和权限
	// 完全访问权限：任务创建者、任务负责人、任务分配者
	isFullAccess := taskInfo.TaskCreator == employeeId ||
		taskInfo.LeaderId.String == employeeId ||
		taskInfo.TaskAssigner.String == employeeId ||
		l.isInEmployeeList(taskInfo.ResponsibleEmployeeIds.String, employeeId)

	// 如果不是完全访问权限，检查是否是节点参与者
	if !isFullAccess {
		// 检查是否参与了任何节点
		isNodeParticipant := false
		var filteredNodes []*taskmodel.TaskNode

		for _, node := range taskNodes {
			// 检查是否是该节点的执行人或负责人
			if l.isInEmployeeList(node.ExecutorId, employeeId) || node.LeaderId == employeeId {
				isNodeParticipant = true
				filteredNodes = append(filteredNodes, node)
			}
		}

		if !isNodeParticipant {
			// 既不是任务负责人，也不是任何节点的参与者，无权查看
			return utils.Response.BusinessError("task_view_denied"), nil
		}

		// 只返回用户参与的节点
		taskNodes = filteredNodes
	}

	// 5. 获取任务日志（如果是节点参与者，只返回相关节点的日志）
	var taskLogs []*taskmodel.TaskLog
	if isFullAccess {
		taskLogs, err = l.svcCtx.TaskLogModel.FindByTaskID(l.ctx, req.TaskID)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("获取任务日志失败: %v", err)
			taskLogs = []*taskmodel.TaskLog{}
		}
	} else {
		// 只获取用户参与节点的日志
		allLogs, err := l.svcCtx.TaskLogModel.FindByTaskID(l.ctx, req.TaskID)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("获取任务日志失败: %v", err)
			allLogs = []*taskmodel.TaskLog{}
		}

		// 构建用户参与的节点ID集合
		nodeIdSet := make(map[string]bool)
		for _, node := range taskNodes {
			nodeIdSet[node.TaskNodeId] = true
		}

		// 过滤日志：只保留任务级别日志和用户参与节点的日志
		for _, log := range allLogs {
			if log.TaskNodeId.String == "" || nodeIdSet[log.TaskNodeId.String] {
				taskLogs = append(taskLogs, log)
			}
		}
	}

	// 6. 转换为响应格式
	converter := utils.NewConverter()
	taskInfoResp := converter.ToTaskInfo(taskInfo)
	// 设置节点数据
	taskInfoResp.Nodes = converter.ToTaskNodeInfoList(taskNodes)

	// 添加权限标识，前端可以根据此字段决定显示哪些操作按钮
	taskDetail := &types.TaskDetailInfo{
		TaskInfo:   taskInfoResp,
		Logs:       converter.ToTaskLogInfoList(taskLogs),
		IsFullAccess: isFullAccess,
	}

	return utils.Response.Success(taskDetail), nil
}

// isInEmployeeList 检查员工ID是否在逗号分隔的员工列表中
func (l *GetTaskLogic) isInEmployeeList(employeeIds string, targetId string) bool {
	if employeeIds == "" {
		return false
	}
	ids := strings.Split(employeeIds, ",")
	for _, id := range ids {
		if strings.TrimSpace(id) == targetId {
			return true
		}
	}
	return false
}
