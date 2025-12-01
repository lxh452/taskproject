// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tasknode

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"task_Project/model/task"
	"task_Project/task/internal/logic/notification"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	fmt.Println("用户", req)
	// 先查看自己是否通过校验
	// 获取当前用户Id
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	emp, err := l.svcCtx.EmployeeModel.FindOneByUserId(l.ctx, userID)
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		l.Logger.Errorf("找不到该员工，reason：%v", err)
		return nil, err
	}
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
		if emp.Id == v {
			flag = true
		}
	}

	//遍历后发现都是空的，及该登录用户不在该存储中，非法用户
	if !flag {
		return utils.Response.NotFoundError("非法用户"), nil
	}
	executorIds := strings.Join(req.ExecutorIDs, ",")

	// 解析时间字段
	var nodeDeadline, nodeStartTime time.Time
	if req.NodeDeadline != "" {
		parsedDeadline, err := time.Parse("2006-01-02 15:04:05", req.NodeDeadline)
		if err != nil {
			// 尝试其他格式
			parsedDeadline, err = time.Parse("2006-01-02", req.NodeDeadline)
			if err == nil {
				nodeDeadline = parsedDeadline
			}
		} else {
			nodeDeadline = parsedDeadline
		}
	}
	if req.NodeStartTime != "" {
		parsedStartTime, err := time.Parse("2006-01-02 15:04:05", req.NodeStartTime)
		if err != nil {
			// 尝试其他格式
			parsedStartTime, err = time.Parse("2006-01-02", req.NodeStartTime)
			if err == nil {
				nodeStartTime = parsedStartTime
			}
		} else {
			nodeStartTime = parsedStartTime
		}
	}

	nodeID := utils.Common.GenId("node")
	node := &task.TaskNode{
		TaskNodeId:     nodeID,
		TaskId:         req.TaskID,
		DepartmentId:   req.DepartmentID,
		NodeName:       req.NodeName,
		NodeDetail:     utils.Common.ToSqlNullString(req.NodeDetail),
		ExNodeIds:      "", // 前置节点ID，创建时为空
		NodeDeadline:   nodeDeadline,
		NodeStartTime:  nodeStartTime,
		EstimatedDays:  req.EstimatedDays,
		ActualDays:     sql.NullInt64{Valid: false}, // 实际完成天数，创建时为空
		NodeStatus:     0,
		NodeFinishTime: sql.NullTime{Valid: false}, // 节点完成时间，创建时为空
		ExecutorId:     executorIds,
		LeaderId:       emp.Id, // 使用员工ID，不是用户ID
		Progress:       0,
		NodePriority:   req.NodePriority,
		DeleteTime:     sql.NullTime{Valid: false}, // 删除时间，创建时为空
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
	_, err = l.svcCtx.TaskNodeModel.InsertTask(l.ctx, node)
	if err != nil {
		l.Logger.Errorf("创建任务节点失败：%v", err)
		return nil, err
	}

	// 发送通知给执行人（异步处理，不阻塞主进程）
	l.Logger.Infof("========== 通知发送检查 ==========")
	l.Logger.Infof("req.ExecutorIDs = %v, 长度 = %d", req.ExecutorIDs, len(req.ExecutorIDs))

	if len(req.ExecutorIDs) > 0 {
		content := fmt.Sprintf("您被分配为任务节点 %s 的执行人，请登录系统查看详情并及时处理", req.NodeName)
		title := fmt.Sprintf("任务节点分配 - %s", req.NodeName)

		l.Logger.Infof("准备为执行人创建通知: executorIDs=%v, nodeID=%s", req.ExecutorIDs, nodeID)

		// 直接批量创建通知（最可靠的方式）
		notificationLogic := notification.NewCreateNotificationLogic(context.Background(), l.svcCtx)
		notificationLogic.CreateNotificationBatch(&notification.BatchCreateNotificationRequest{
			EmployeeIDs: req.ExecutorIDs,
			Title:       title,
			Content:     content,
			Type:        1, // 任务类型
			Priority:    int(req.NodePriority),
			RelatedID:   nodeID,
			RelatedType: "任务节点分配",
		})

		// 发布邮件事件（如果邮件服务可用）
		if l.svcCtx.EmailMQService != nil {
			go func() {
				ctx := context.Background()
				deadlineStr := ""
				if !nodeDeadline.IsZero() {
					deadlineStr = nodeDeadline.Format("2006-01-02")
				}
				emailEvent := &svc.EmailEvent{
					EventType:   svc.TaskNodeCreated,
					EmployeeIDs: req.ExecutorIDs,
					Subject:     title,
					Body:        fmt.Sprintf("%s\n\n截止时间：%s", content, deadlineStr),
					IsHTML:      false,
					TaskID:      req.TaskID,
					NodeID:      nodeID,
				}
				if err := l.svcCtx.EmailMQService.PublishEmailEvent(ctx, emailEvent); err != nil {
					logx.Errorf("发布任务节点邮件事件失败: %v", err)
				}
			}()
		}
	} else {
		l.Logger.Info("没有执行人，跳过通知发送")
	}

	return utils.Response.Success(node), nil
}
