package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"task_Project/model/user_auth"
	"task_Project/task/internal/utils"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// 枚举需要通知的事务
const (
	// 任务相关
	TaskCreated                = "task.created"
	TaskUpdated                = "task.updated"
	TaskDeleted                = "task.deleted"
	TaskCompleted              = "task.completed"
	TaskNodeExecutorChanged    = "task.node.executor.changed"
	TaskNodeCreated            = "task.node.created"
	TaskNodeDeleted            = "task.node.deleted"
	TaskNodeCompleted          = "task.node.completed"
	TaskNodeCompletionApproval = "task.node.completion.approval"
	TaskDeadlineReminder       = "task.deadline.reminder"
	TaskSlowProgress           = "task.slow.progress"
	TaskNodeExecutorLeft       = "task.node.executor.left"

	// 员工相关
	EmployeeCreated = "employee.created"
	EmployeeLeave   = "employee.leave"

	// 部门相关
	DepartmentCreated = "department.created"

	// 交接相关
	HandoverNotification = "handover.notification"
)

// NotificationEvent 通知事件消息结构
type NotificationEvent struct {
	EventType   string   `json:"eventType"`   // 事件类型：task.created, task.updated, task.completed, task.node.completed 等
	EmployeeIDs []string `json:"employeeIds"` // 需要通知的员工ID列表（如果为空则根据业务ID查询）
	Title       string   `json:"title"`       // 通知标题（可选，如果为空则根据事件类型生成）
	Content     string   `json:"content"`     // 通知内容（可选，如果为空则根据事件类型生成）
	Type        int      `json:"type"`        // 通知类型
	Category    string   `json:"category"`    // 通知分类
	Priority    int      `json:"priority"`    // 优先级
	RelatedID   string   `json:"relatedId"`   // 关联的业务ID（如任务ID、节点ID）
	RelatedType string   `json:"relatedType"` // 关联的业务类型（如 task, tasknode）
	// 业务相关字段（用于查询需要通知的员工）
	TaskID string `json:"taskId"` // 任务ID（用于查询任务相关人员）
	NodeID string `json:"nodeId"` // 节点ID（用于查询节点相关人员）
}

// NotificationMQService 通知消息队列服务（用于发布通知事件）
type NotificationMQService struct {
	mqClient *MQClient
}

// NewNotificationMQService 创建通知消息队列服务
func NewNotificationMQService(mqClient *MQClient) *NotificationMQService {
	return &NotificationMQService{
		mqClient: mqClient,
	}
}

// PublishNotificationEvent 发布通知事件到消息队列
func (s *NotificationMQService) PublishNotificationEvent(ctx context.Context, event *NotificationEvent) error {
	logx.WithContext(ctx).Infof("[NotificationMQ] PublishNotificationEvent called: eventType=%s, employeeIds=%v, relatedId=%s, taskId=%s, nodeId=%s",
		event.EventType, event.EmployeeIDs, event.RelatedID, event.TaskID, event.NodeID)

	if s == nil {
		logx.WithContext(ctx).Error("[NotificationMQ] NotificationMQService is nil, notification event will be ignored")
		return nil
	}

	if s.mqClient == nil {
		logx.WithContext(ctx).Error("[NotificationMQ] MQClient is nil, notification event will be ignored")
		return nil // 如果 MQ 未初始化，静默失败，不影响主流程
	}

	// 路由键格式：notification.{eventType}
	routingKey := fmt.Sprintf("notification.%s", event.EventType)
	logx.WithContext(ctx).Infof("[NotificationMQ] Publishing to routingKey: %s", routingKey)

	err := s.mqClient.Publish(routingKey, event)
	if err != nil {
		logx.WithContext(ctx).Errorf("[NotificationMQ] Failed to publish notification event: %v", err)
		return err
	}

	logx.WithContext(ctx).Infof("[NotificationMQ] Successfully published notification event: %s for %d employees", event.EventType, len(event.EmployeeIDs))
	return nil
}

// StartNotificationConsumer 启动通知消费者（处理通知创建）
func StartNotificationConsumer(mqClient *MQClient, queueName string, svcCtx *ServiceContext) error {
	if mqClient == nil {
		return fmt.Errorf("MQClient is nil")
	}

	// 声明队列
	queue, err := mqClient.GetChannel().QueueDeclare(
		queueName, // 队列名称
		true,      // 持久化
		false,     // 自动删除
		false,     // 排他
		false,     // 不等待
		nil,       // 参数
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// 绑定队列到交换机，监听所有通知事件
	err = mqClient.GetChannel().QueueBind(
		queue.Name,             // 队列名
		"notification.*",       // 路由键模式：匹配所有 notification.* 开头的消息
		mqClient.GetExchange(), // 交换机
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// 消费消息
	msgs, err := mqClient.GetChannel().Consume(
		queue.Name, // 队列
		"",         // 消费者标签
		false,      // 自动确认（改为 false，手动确认）
		false,      // 排他
		false,      // 不等待
		false,      // 无额外参数
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// 启动 goroutine 处理消息
	go func() {
		for msg := range msgs {
			handleNotificationMessage(msg, svcCtx)
		}
	}()

	logx.Infof("Notification consumer started, queue: %s", queueName)
	return nil
}

// handleNotificationMessage 处理通知消息
func handleNotificationMessage(msg amqp.Delivery, svcCtx *ServiceContext) {
	ctx := context.Background()

	logx.Infof("[NotificationMQ Consumer] Received message: routingKey=%s, body=%s", msg.RoutingKey, string(msg.Body))

	var event NotificationEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		logx.Errorf("[NotificationMQ Consumer] Failed to unmarshal notification event: %v", err)
		msg.Nack(false, false) // 拒绝消息，不重新入队
		return
	}

	logx.Infof("[NotificationMQ Consumer] Parsed event: eventType=%s, employeeIds=%v, relatedId=%s", event.EventType, event.EmployeeIDs, event.RelatedID)

	// 如果 EmployeeIDs 为空，根据业务ID查询需要通知的员工
	employeeIDs := event.EmployeeIDs
	if len(employeeIDs) == 0 {
		logx.Infof("[NotificationMQ Consumer] EmployeeIDs is empty, resolving recipients...")
		employeeIDs = resolveNotificationRecipients(ctx, svcCtx, &event)
		logx.Infof("[NotificationMQ Consumer] Resolved %d recipients", len(employeeIDs))
	}

	// 如果 Title 或 Content 为空，根据事件类型生成
	if event.Title == "" || event.Content == "" {
		logx.Infof("[NotificationMQ Consumer] Title or Content is empty, generating content...")
		title, content := generateNotificationContent(ctx, svcCtx, &event)
		if event.Title == "" {
			event.Title = title
		}
		if event.Content == "" {
			event.Content = content
		}
		logx.Infof("[NotificationMQ Consumer] Generated title=%s, content=%s", event.Title, event.Content)
	}

	// 为每个员工创建通知
	successCount := 0
	for _, employeeID := range employeeIDs {
		// 确保使用员工主键 Id（通知表使用员工主键存储）
		actualEmployeeID := employeeID
		// 验证员工是否存在
		emp, err := svcCtx.EmployeeModel.FindOne(ctx, employeeID)
		if err != nil {
			logx.Errorf("[NotificationMQ Consumer] Failed to find employee by id %s: %v", employeeID, err)
			continue
		}
		// 使用员工主键 Id
		actualEmployeeID = emp.Id
		logx.Infof("[NotificationMQ Consumer] Creating notification for employee: %s (ID: %s)", emp.RealName, actualEmployeeID)

		notification := &user_auth.Notification{
			Id:          utils.Common.GenId("notification"),
			EmployeeId:  actualEmployeeID,
			Title:       event.Title,
			Content:     event.Content,
			Type:        int64(event.Type),
			Category:    sql.NullString{String: event.Category, Valid: event.Category != ""},
			IsRead:      0,
			ReadTime:    sql.NullTime{}, // 未读时为空
			Priority:    int64(event.Priority),
			SenderId:    sql.NullString{String: "system", Valid: true}, // 系统发送
			RelatedId:   sql.NullString{String: event.RelatedID, Valid: event.RelatedID != ""},
			RelatedType: sql.NullString{String: event.RelatedType, Valid: event.RelatedType != ""},
			CreateTime:  time.Now(),
			UpdateTime:  time.Now(),
		}

		_, insertErr := svcCtx.NotificationModel.Insert(ctx, notification)
		if insertErr != nil {
			logx.Errorf("[NotificationMQ Consumer] Failed to create notification for employee %s: %v", actualEmployeeID, insertErr)
			continue
		}
		logx.Infof("[NotificationMQ Consumer] Created notification for employee %s, notificationId=%s", actualEmployeeID, notification.Id)
		successCount++
	}

	logx.Infof("[NotificationMQ Consumer] Created %d/%d notifications for event: %s", successCount, len(employeeIDs), event.EventType)

	// 手动确认消息
	msg.Ack(false)
}

// resolveNotificationRecipients 根据业务ID解析需要通知的员工
func resolveNotificationRecipients(ctx context.Context, svcCtx *ServiceContext, event *NotificationEvent) []string {
	employeeIDSet := make(map[string]bool)

	switch event.EventType {
	case TaskCompleted, TaskUpdated, TaskCreated, TaskDeleted:
		// 根据 TaskID 查询任务创建者和所有节点负责人/执行人
		if event.TaskID != "" {
			taskInfo, err := svcCtx.TaskModel.FindOne(ctx, event.TaskID)
			if err == nil {
				if taskInfo.TaskCreator != "" {
					employeeIDSet[taskInfo.TaskCreator] = true
				}
				// 查询所有节点
				nodes, err := svcCtx.TaskNodeModel.FindByTaskID(ctx, event.TaskID)
				if err == nil {
					for _, node := range nodes {
						if node.LeaderId != "" {
							employeeIDSet[node.LeaderId] = true
						}
						if node.ExecutorId != "" {
							employeeIDSet[node.ExecutorId] = true
						}
					}
				}
			}
		}
	case TaskNodeExecutorChanged, TaskNodeDeleted, TaskNodeCompleted, TaskDeadlineReminder, TaskSlowProgress, TaskNodeExecutorLeft:
		// 根据 NodeID 查询相关人员
		if event.NodeID != "" {
			node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
			if err == nil {
				if event.EventType == "task.deadline.reminder" {
					// 截止提醒只通知执行人
					if node.ExecutorId != "" {
						employeeIDSet[node.ExecutorId] = true
					}
				} else if event.EventType == "task.slow.progress" {
					// 进度缓慢通知负责人和执行人
					if node.LeaderId != "" {
						employeeIDSet[node.LeaderId] = true
					}
					if node.ExecutorId != "" {
						employeeIDSet[node.ExecutorId] = true
					}
				} else if event.EventType == "task.node.executor.left" {
					// 执行人离职，通知负责人
					if len(event.EmployeeIDs) > 0 {
						employeeIDSet[event.EmployeeIDs[0]] = true
					} else if node.LeaderId != "" {
						employeeIDSet[node.LeaderId] = true
					}
				} else {
					// 其他情况通知负责人和执行人
					if node.LeaderId != "" {
						employeeIDSet[node.LeaderId] = true
					}
					if node.ExecutorId != "" {
						employeeIDSet[node.ExecutorId] = true
					}
				}
			}
		}
	case TaskNodeCreated:
		// 根据 TaskID 查询任务创建者和节点负责人
		if event.TaskID != "" {
			taskInfo, err := svcCtx.TaskModel.FindOne(ctx, event.TaskID)
			if err == nil {
				if taskInfo.TaskCreator != "" {
					employeeIDSet[taskInfo.TaskCreator] = true
				}
			}
			// 如果指定了 EmployeeIDs，也加入
			for _, id := range event.EmployeeIDs {
				employeeIDSet[id] = true
			}
		}
	case HandoverNotification:
		// 交接通知，根据 RelatedID 查询交接记录
		if event.RelatedID != "" {
			handover, err := svcCtx.TaskHandoverModel.FindOne(ctx, event.RelatedID)
			if err == nil {
				if handover.ToEmployeeId != "" {
					employeeIDSet[handover.ToEmployeeId] = true
				}
				if handover.FromEmployeeId != "" {
					employeeIDSet[handover.FromEmployeeId] = true
				}
			}
		}
		// 如果指定了 EmployeeIDs，也加入
		for _, id := range event.EmployeeIDs {
			employeeIDSet[id] = true
		}
	case EmployeeLeave:
		// 员工离职通知，使用指定的 EmployeeIDs
		for _, id := range event.EmployeeIDs {
			employeeIDSet[id] = true
		}
	}

	// 转换为切片
	employeeIDs := make([]string, 0, len(employeeIDSet))
	for id := range employeeIDSet {
		employeeIDs = append(employeeIDs, id)
	}
	return employeeIDs
}

// generateNotificationContent 根据事件类型生成通知内容
func generateNotificationContent(ctx context.Context, svcCtx *ServiceContext, event *NotificationEvent) (title, content string) {
	switch event.EventType {
	case TaskCompleted:
		if event.TaskID != "" {
			taskInfo, err := svcCtx.TaskModel.FindOne(ctx, event.TaskID)
			if err == nil {
				title = "任务完成通知"
				content = fmt.Sprintf("任务 %s 已完成", taskInfo.TaskTitle)
			}
		}
	case TaskUpdated:
		if event.TaskID != "" {
			taskInfo, err := svcCtx.TaskModel.FindOne(ctx, event.TaskID)
			if err == nil {
				title = "任务更新通知"
				content = fmt.Sprintf("任务 %s 已更新", taskInfo.TaskTitle)
			}
		}
	case TaskDeleted:
		if event.TaskID != "" {
			taskInfo, err := svcCtx.TaskModel.FindOne(ctx, event.TaskID)
			if err == nil {
				title = "任务删除通知"
				content = fmt.Sprintf("任务 %s 已被删除", taskInfo.TaskTitle)
			}
		}
	case TaskNodeExecutorChanged:
		if event.NodeID != "" {
			node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
			if err == nil {
				title = "任务交接通知"
				content = fmt.Sprintf("您已被分配为任务节点 %s 的执行人", node.NodeName)
			}
		}
	case TaskNodeDeleted:
		if event.NodeID != "" {
			node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
			if err == nil {
				title = "任务节点删除通知"
				content = fmt.Sprintf("任务节点 %s 已被删除", node.NodeName)
			}
		}
	case TaskNodeCompleted:
		if event.NodeID != "" {
			node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
			if err == nil {
				title = "任务完成通知"
				content = fmt.Sprintf("任务节点 %s 已完成", node.NodeName)
			}
		}
	case TaskNodeCreated:
		if event.TaskID != "" {
			taskInfo, err := svcCtx.TaskModel.FindOne(ctx, event.TaskID)
			if err == nil {
				title = taskInfo.TaskTitle
				content = fmt.Sprintf("您现在为%s:%s任务的节点负责人，请登录系统进行查看，如无误，请尽快安排人手进行处理", event.TaskID, taskInfo.TaskTitle)
			}
		}
	case TaskDeadlineReminder:
		if event.NodeID != "" {
			node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
			if err == nil {
				title = "任务截止时间提醒"
				content = fmt.Sprintf("任务节点 %s 即将到期，请及时完成", node.NodeName)
			}
		}
	case TaskSlowProgress:
		if event.NodeID != "" {
			node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
			if err == nil {
				title = "任务进度缓慢提醒"
				content = fmt.Sprintf("任务节点 %s 进度可能较慢，当前进度：%d%%", node.NodeName, node.Progress)
			}
		}
	case TaskNodeExecutorLeft:
		if event.NodeID != "" {
			node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
			if err == nil {
				title = "任务节点执行人离职通知"
				content = fmt.Sprintf("任务节点 %s 的执行人已离职，系统已自动处理", node.NodeName)
			}
		}
	case EmployeeLeave:
		if len(event.EmployeeIDs) > 0 {
			employee, err := svcCtx.EmployeeModel.FindOne(ctx, event.EmployeeIDs[0])
			if err == nil {
				title = "员工离职通知"
				content = fmt.Sprintf("员工 %s 已离职", employee.RealName)
			}
		}
	case HandoverNotification:
		if event.RelatedID != "" {
			handover, err := svcCtx.TaskHandoverModel.FindOne(ctx, event.RelatedID)
			if err == nil {
				title = "任务交接通知"
				if handover.TaskId != "" {
					taskInfo, err := svcCtx.TaskModel.FindOne(ctx, handover.TaskId)
					if err == nil {
						// 根据交接状态和接收人/发起人确定内容
						if len(event.EmployeeIDs) > 0 && event.EmployeeIDs[0] == handover.ToEmployeeId {
							// 发给接收人
							content = fmt.Sprintf("您收到了一个任务交接请求：%s", taskInfo.TaskTitle)
						} else if len(event.EmployeeIDs) > 0 && event.EmployeeIDs[0] == handover.FromEmployeeId {
							// 发给发起人
							if handover.HandoverStatus == 2 {
								content = "任务交接完成"
							} else if handover.HandoverStatus == 3 {
								content = "任务交接已拒绝"
							} else {
								content = "交接已确认"
							}
						} else {
							content = fmt.Sprintf("您收到了一个任务交接通知：%s", taskInfo.TaskTitle)
						}
					} else {
						content = "您收到了一个任务交接通知"
					}
				} else {
					content = "您收到了一个任务交接通知"
				}
			}
		}
	default:
		title = event.Title
		content = event.Content
	}
	return
}

// NotificationEventOptions 扩展参数，用于传入可选的 TaskID、NodeID 等
type NotificationEventOptions struct {
	TaskID string // 任务ID（可选）
	NodeID string // 节点ID（可选）
}

// 创建通知实体类，extras 为可选扩展参数（最多传一个）
func (s *NotificationMQService) NewNotificationEvent(eventType string, employeeIds []string, relatedID string, extras ...NotificationEventOptions) *NotificationEvent {
	var category string
	switch eventType {
	case EmployeeLeave:
		category = "employee"
	case HandoverNotification:
		category = "handover"
	default:
		category = "task"
	}

	// 解析扩展参数
	var taskID, nodeID string
	if len(extras) > 0 {
		taskID = extras[0].TaskID
		nodeID = extras[0].NodeID
	}

	// 这里构造一个用于描述通知元数据的 map，方便后续扩展，也可以用于拼接默认内容
	meta := map[string]interface{}{
		"eventType": eventType,
	}
	if relatedID != "" {
		meta["relatedId"] = relatedID
	}
	if len(employeeIds) > 0 {
		meta["employeeCount"] = len(employeeIds)
	}
	if taskID != "" {
		meta["taskId"] = taskID
	}
	if nodeID != "" {
		meta["nodeId"] = nodeID
	}

	// 通用的标题和内容（对于未在 generateNotificationContent 中单独处理的事件类型会生效）
	title := ""
	switch eventType {
	case TaskCreated:
		title = "任务创建通知"
	case TaskUpdated:
		title = "任务更新通知"
	case TaskCompleted:
		title = "任务完成通知"
	case TaskDeleted:
		title = "任务删除通知"
	case TaskNodeCreated:
		title = "任务节点创建通知"
	case TaskNodeCompleted:
		title = "任务节点完成通知"
	case TaskNodeCompletionApproval:
		title = "任务节点完成审批"
	case TaskDeadlineReminder:
		title = "任务截止时间提醒"
	case TaskSlowProgress:
		title = "任务进度缓慢提醒"
	case TaskNodeExecutorLeft:
		title = "任务节点执行人离职通知"
	case EmployeeLeave:
		title = "员工离职通知"
	case HandoverNotification:
		title = "任务交接通知"
	default:
		title = "系统通知"
	}

	// 使用 meta 生成一个简易的内容描述，便于在前端调试或展示更多上下文
	metaBytes, _ := json.Marshal(meta)
	content := fmt.Sprintf("事件：%s，详情：%s", eventType, string(metaBytes))

	// 根据事件类型大致推断关联类型
	relatedType := ""
	switch eventType {
	case HandoverNotification:
		relatedType = "handover"
	case EmployeeLeave:
		relatedType = "employee"
	default:
		if len(eventType) >= 5 && eventType[:5] == "task." {
			relatedType = "task"
		}
	}

	return &NotificationEvent{
		EventType:   eventType,
		EmployeeIDs: employeeIds,
		Title:       title,
		Content:     content,
		Type:        0,
		Category:    category,
		Priority:    0,
		RelatedID:   relatedID,
		RelatedType: relatedType,
		TaskID:      taskID,
		NodeID:      nodeID,
	}
}
