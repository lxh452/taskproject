package task

import (
	"context"
	"errors"

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

	// 4. 获取任务日志
	taskLogs, err := l.svcCtx.TaskLogModel.FindByTaskID(l.ctx, req.TaskID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("获取任务日志失败: %v", err)
		taskLogs = []*taskmodel.TaskLog{} // 设置为空数组，不影响主流程
	}

	// 5. 转换为响应格式
	converter := utils.NewConverter()
	taskInfoResp := converter.ToTaskInfo(taskInfo)
	// 设置节点数据
	taskInfoResp.Nodes = converter.ToTaskNodeInfoList(taskNodes)

	taskDetail := &types.TaskDetailInfo{
		TaskInfo: taskInfoResp,
		Logs:     converter.ToTaskLogInfoList(taskLogs),
	}

	return utils.Response.Success(taskDetail), nil
}
