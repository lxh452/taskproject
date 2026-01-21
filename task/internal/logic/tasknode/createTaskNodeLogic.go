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
	"task_Project/model/user"
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
	// 先查看自己是否通过校验
	// 获取当前用户Id
	empID, ok := utils.Common.GetCurrentEmployeeID(l.ctx)
	emp, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, empID)
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		l.Logger.WithContext(l.ctx).Errorf("找不到该员工，reason：%v", err)
		return nil, err
	}
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}
	// 首先查看总任务是否存在 并且查看是否为空
	currentTask, err := l.svcCtx.TaskModel.FindOne(l.ctx, req.TaskID)
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		l.Logger.WithContext(l.ctx).Errorf("找不到该总任务，reason：%v", err)
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
		if empID == v {
			flag = true
		}
	}
	if empID == currentTask.TaskCreator {
		flag = true
	}

	//遍历后发现都是空的，及该登录用户不在该存储中，非法用户
	if !flag {
		return utils.Response.NotFoundError("非法用户"), nil
	}
	// 检查权限：不能以下犯上（但允许给自己安排任务）
	if err := l.checkAssignmentPermission(emp, req.ExecutorIDs); err != nil {
		return utils.Response.BusinessError(err.Error()), nil
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
		LeaderId:       empID,
		Progress:       0,
		NodePriority:   req.NodePriority,
		DeleteTime:     sql.NullTime{Valid: false}, // 删除时间，创建时为空
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
	_, err = l.svcCtx.TaskNodeModel.InsertTask(l.ctx, node)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务节点失败：%v", err)
		return nil, err
	}

	// 更新任务的 node_employee_ids（去重）
	if len(req.ExecutorIDs) > 0 {
		// 获取当前任务的 node_employee_ids
		currentNodeEmployeeIds := ""
		if currentTask.NodeEmployeeIds.Valid {
			currentNodeEmployeeIds = currentTask.NodeEmployeeIds.String
		}

		// 创建一个 map 用于去重
		employeeIdSet := make(map[string]bool)

		// 添加现有的员工ID
		if currentNodeEmployeeIds != "" {
			existingIds := strings.Split(currentNodeEmployeeIds, ",")
			for _, id := range existingIds {
				trimmedId := strings.TrimSpace(id)
				if trimmedId != "" {
					employeeIdSet[trimmedId] = true
				}
			}
		}

		// 添加新的执行人ID
		for _, executorId := range req.ExecutorIDs {
			trimmedId := strings.TrimSpace(executorId)
			if trimmedId != "" {
				employeeIdSet[trimmedId] = true
			}
		}

		// 将 map 转换回逗号分隔的字符串
		var updatedEmployeeIds []string
		for id := range employeeIdSet {
			updatedEmployeeIds = append(updatedEmployeeIds, id)
		}
		newNodeEmployeeIds := strings.Join(updatedEmployeeIds, ",")

		// 更新任务的 node_employee_ids
		err = l.svcCtx.TaskModel.UpdateNodeEmployeeIds(l.ctx, req.TaskID, newNodeEmployeeIds)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务的节点员工列表失败：%v", err)
			// 不返回错误，因为节点已经创建成功，这只是辅助信息更新失败
		} else {
			l.Logger.WithContext(l.ctx).Infof("成功更新任务 %s 的节点员工列表: %s", req.TaskID, newNodeEmployeeIds)
		}
	}

	// 更新任务的总节点数
	totalNodeCount, err := l.svcCtx.TaskNodeModel.GetTaskNodeCountByTask(l.ctx, req.TaskID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("获取任务节点总数失败：%v", err)
	} else {
		completedNodeCount, err := l.svcCtx.TaskNodeModel.GetCompletedNodeCountByTask(l.ctx, req.TaskID)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("获取已完成节点数失败：%v", err)
			completedNodeCount = 0
		}
		err = l.svcCtx.TaskModel.UpdateNodeCount(l.ctx, req.TaskID, totalNodeCount, completedNodeCount)
		if err != nil {
			l.Logger.WithContext(l.ctx).Errorf("更新任务节点统计失败：%v", err)
		}
	}

	if len(req.ExecutorIDs) > 0 {
		content := fmt.Sprintf("您被分配为任务节点 %s 的执行人，请登录系统查看详情并及时处理", req.NodeName)
		title := fmt.Sprintf("任务节点分配 - %s", req.NodeName)

		l.Logger.WithContext(l.ctx).Infof("准备为执行人创建通知: executorIDs=%v, nodeID=%s", req.ExecutorIDs, nodeID)

		// 直接批量创建通知（最可靠的方式）
		notificationLogic := notification.NewCreateNotificationLogic(context.Background(), l.svcCtx)
		notificationLogic.CreateNotificationBatch(&notification.BatchCreateNotificationRequest{
			EmployeeIDs: req.ExecutorIDs,
			Title:       title,
			Content:     content,
			Type:        1, // 任务类型
			Priority:    int(req.NodePriority),
			RelatedID:   nodeID,
			RelatedType: "task",
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
		l.Logger.WithContext(l.ctx).Info("没有执行人，跳过通知发送")
	}
	// 创建任务日志
	taskLog := &task.TaskLog{
		LogId:      utils.NewCommon().GenId("task_log"),
		TaskId:     req.TaskID,
		LogType:    2, //更新内容
		EmployeeId: empID,
		TaskNodeId: utils.Common.ToSqlNullString(nodeID),
		CreateTime: time.Now(),
	}
	_, err = l.svcCtx.TaskLogModel.Insert(l.ctx, taskLog)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建任务失败")
		return utils.Response.BusinessError("task_log_error"), nil
	}
	return utils.Response.Success(node), nil
}

// checkAssignmentPermission 检查任务分配权限
// 规则：
// 1. 创始人可以给所有人分配任务
// 2. 不能以下犯上（下级不能给上级安排任务）
// 3. 允许给自己安排任务
func (l *CreateTaskNodeLogic) checkAssignmentPermission(assigner *user.Employee, executorIDs []string) error {
	// 检查是否是创始人
	isFounder := l.isFounder(assigner)
	if isFounder {
		// 创始人可以给所有人分配任务，无需检查
		return nil
	}

	// 获取分配者的职位级别和部门优先级
	assignerLevel := 0
	assignerDeptPriority := 0
	if assigner.PositionId.Valid {
		pos, err := l.svcCtx.PositionModel.FindOne(l.ctx, assigner.PositionId.String)
		if err == nil && pos != nil {
			assignerLevel = int(pos.PositionLevel)
		}
	}
	if assigner.DepartmentId.Valid {
		dept, err := l.svcCtx.DepartmentModel.FindOne(l.ctx, assigner.DepartmentId.String)
		if err == nil && dept != nil {
			assignerDeptPriority = int(dept.DepartmentPriority)
		}
	}

	// 检查每个执行人
	for _, executorID := range executorIDs {
		executorID = strings.TrimSpace(executorID)
		if executorID == "" {
			continue
		}

		// 允许给自己安排任务
		if executorID == assigner.Id {
			continue
		}

		// 获取执行人信息
		executor, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, executorID)
		if err != nil {
			return fmt.Errorf("执行人 %s 不存在", executorID)
		}

		// 检查执行人是否是创始人（不能给创始人安排任务，除非分配者也是创始人）
		if l.isFounder(executor) {
			return fmt.Errorf("不能给公司创始人安排任务")
		}

		// 获取执行人的职位级别和部门优先级
		executorLevel := 0
		executorDeptPriority := 0
		if executor.PositionId.Valid {
			pos, err := l.svcCtx.PositionModel.FindOne(l.ctx, executor.PositionId.String)
			if err == nil && pos != nil {
				executorLevel = int(pos.PositionLevel)
			}
		}
		if executor.DepartmentId.Valid {
			dept, err := l.svcCtx.DepartmentModel.FindOne(l.ctx, executor.DepartmentId.String)
			if err == nil && dept != nil {
				executorDeptPriority = int(dept.DepartmentPriority)
			}
		}

		// 检查是否以下犯上
		// 如果执行人的职位级别更高，或者部门优先级更高，则不能分配
		if executorLevel > assignerLevel {
			return fmt.Errorf("不能给职位级别更高的员工安排任务")
		}
		if executorLevel == assignerLevel && executorDeptPriority > assignerDeptPriority {
			return fmt.Errorf("不能给部门优先级更高的员工安排任务")
		}
	}

	return nil
}

// isFounder 检查员工是否是创始人
func (l *CreateTaskNodeLogic) isFounder(employee *user.Employee) bool {
	// 检查职位代码是否为 FOUNDER
	if employee.PositionId.Valid {
		pos, err := l.svcCtx.PositionModel.FindOne(l.ctx, employee.PositionId.String)
		if err == nil && pos != nil && pos.PositionCode.Valid {
			if pos.PositionCode.String == "FOUNDER" {
				return true
			}
		}
	}
	// 检查是否是公司Owner
	company, err := l.svcCtx.CompanyModel.FindOne(l.ctx, employee.CompanyId)
	if err == nil && company != nil && company.Owner == employee.UserId {
		return true
	}
	return false
}
