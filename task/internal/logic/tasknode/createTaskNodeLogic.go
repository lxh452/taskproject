// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tasknode

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
	"task_Project/model/task"
	"task_Project/task/internal/utils"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTaskNodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建任务节点
func NewCreateTaskNodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTaskNodeLogic {
	return &CreateTaskNodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTaskNodeLogic) CreateTaskNode(req *types.CreateTaskNodeRequest) (resp *types.BaseResponse, err error) {
	resp = new(types.BaseResponse)
	// 先查看自己是否通过校验
	// 获取当前用户Id
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}
	// 首先查看总任务是否存在 并且查看是否为空
	currentTask, err := l.svcCtx.TaskModel.FindOne(l.ctx, req.TaskID)
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		l.Logger.Errorf("找不到该总任务，reason：%v", err)
		return nil, err
	}
	if currentTask == nil {
		return resp, nil
	}
	// 然后查阅自己是否存在该任务节点中 是否是该任务的节点的负责人，否则无法创建节点
	// 将字符串分割的节点负责人列表进行
	nodeIDs := strings.Split(currentTask.NodeEmployeeIds.String, ",")
	var flag bool
	for _, v := range nodeIDs {
		if userID == v {
			flag = true
		}
	}
	//遍历后发现都是空的，及该登录用户不在该存储中，非法用户
	if !flag {
		return utils.Response.NotFoundError("非法用户"), nil
	}
	executorIds := strings.Join(req.ExecutorIDs, ",")
	node := &task.TaskNode{
		TaskNodeId:    utils.Common.GenId("node"),
		TaskId:        req.TaskID,
		DepartmentId:  req.DepartmentID,
		NodeName:      req.NodeName,
		NodeDetail:    utils.Common.ToSqlNullString(req.NodeDetail),
		EstimatedDays: req.EstimatedDays,
		NodeStatus:    0,
		ExecutorId:    executorIds,
		LeaderId:      userID,
		Progress:      0,
		NodePriority:  req.NodePriority,
		CreateTime:    time.Now(),
	}
	_, err = l.svcCtx.TaskNodeModel.Insert(l.ctx, node)
	if err != nil {
		l.Logger.Errorf("创建任务节点失败：%v", err)
		return nil, err
	}
	return utils.Response.Success(node), nil
}
