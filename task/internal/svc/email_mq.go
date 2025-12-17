package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"task_Project/task/internal/middleware"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logx"
)

// EmailEvent 邮件事件消息结构
type EmailEvent struct {
	EventType string   `json:"eventType"` // 事件类型：task.completed, task.updated, task.node.executor.changed 等
	To        []string `json:"to"`        // 收件人列表（邮箱地址，如果为空则根据业务ID查询）
	Subject   string   `json:"subject"`   // 邮件主题（可选，如果为空则根据事件类型生成）
	Body      string   `json:"body"`      // 邮件内容（可选，如果为空则根据事件类型和模板生成）
	IsHTML    bool     `json:"isHtml"`    // 是否为HTML格式
	// 业务相关字段（用于查询收件人）
	TaskID      string   `json:"taskId"`      // 任务ID（用于查询任务相关人员）
	NodeID      string   `json:"nodeId"`      // 节点ID（用于查询节点相关人员）
	EmployeeID  string   `json:"employeeId"`  // 员工ID（直接指定收件人）
	EmployeeIDs []string `json:"employeeIds"` // 员工ID列表（用于指定多个收件人）
	RelatedID   string   `json:"relatedId"`   // 关联ID（如交接ID，用于查询关联数据）
	// 模板数据（用于生成邮件内容）
	TemplateData map[string]interface{} `json:"templateData"` // 模板数据
}

// EmailMQService 邮件消息队列服务（用于发布邮件事件）
type EmailMQService struct {
	mqClient        *MQClient
	emailMiddleware *middleware.EmailMiddleware
}

// NewEmailMQService 创建邮件消息队列服务
func NewEmailMQService(mqClient *MQClient, emailMiddleware *middleware.EmailMiddleware) *EmailMQService {
	return &EmailMQService{
		mqClient:        mqClient,
		emailMiddleware: emailMiddleware,
	}
}

// PublishEmailEvent 发布邮件事件到消息队列
func (s *EmailMQService) PublishEmailEvent(ctx context.Context, event *EmailEvent) error {
	logx.WithContext(ctx).Infof("[EmailMQ] Attempting to publish email event: eventType=%s, to=%v, taskId=%s, nodeId=%s",
		event.EventType, event.To, event.TaskID, event.NodeID)

	if s.mqClient == nil {
		logx.WithContext(ctx).Errorf("[EmailMQ] MQClient is nil, email event will be ignored: eventType=%s, to=%v",
			event.EventType, event.To)
		return nil // 如果 MQ 未初始化，静默失败，不影响主流程
	}

	// 路由键格式：email.{eventType}
	routingKey := fmt.Sprintf("email.%s", event.EventType)
	logx.WithContext(ctx).Infof("[EmailMQ] Publishing to routing key: %s", routingKey)

	err := s.mqClient.Publish(routingKey, event)
	if err != nil {
		logx.WithContext(ctx).Errorf("[EmailMQ] Failed to publish email event: routingKey=%s, error=%v, event=%+v",
			routingKey, err, event)
		return err
	}

	logx.WithContext(ctx).Infof("[EmailMQ] Successfully published email event: eventType=%s, routingKey=%s, to=%d recipients (%v)",
		event.EventType, routingKey, len(event.To), event.To)
	return nil
}

// StartEmailConsumer 启动邮件消费者（处理邮件发送）
func StartEmailConsumer(mqClient *MQClient, queueName string, svcCtx *ServiceContext) error {
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

	// 绑定队列到交换机，监听所有邮件事件
	err = mqClient.GetChannel().QueueBind(
		queue.Name,             // 队列名
		"email.*",              // 路由键模式：匹配所有 email.* 开头的消息
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
			handleEmailMessage(msg, svcCtx)
		}
	}()

	logx.Infof("[EmailMQ Consumer] Email consumer started successfully: queue=%s, exchange=%s",
		queueName, mqClient.GetExchange())
	return nil
}

// handleEmailMessage 处理邮件消息
func handleEmailMessage(msg amqp.Delivery, svcCtx *ServiceContext) {
	ctx := context.Background()

	logx.Infof("[EmailMQ Consumer] Received message: routingKey=%s, messageId=%s, bodySize=%d",
		msg.RoutingKey, msg.MessageId, len(msg.Body))

	var event EmailEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		logx.Errorf("[EmailMQ Consumer] Failed to unmarshal email event: error=%v, body=%s", err, string(msg.Body))
		msg.Nack(false, false) // 拒绝消息，不重新入队
		return
	}

	logx.Infof("[EmailMQ Consumer] Parsed email event: eventType=%s, to=%v, taskId=%s, nodeId=%s, hasSubject=%v, hasBody=%v",
		event.EventType, event.To, event.TaskID, event.NodeID, event.Subject != "", event.Body != "")

	// 如果 To 为空，根据业务ID查询收件人邮箱
	emails := event.To
	if len(emails) == 0 {
		logx.Infof("[EmailMQ Consumer] To field is empty, resolving recipients from business IDs")
		emails = resolveEmailRecipients(ctx, svcCtx, &event)
		logx.Infof("[EmailMQ Consumer] Resolved %d email addresses: %v", len(emails), emails)
	}

	if len(emails) == 0 {
		logx.Errorf("[EmailMQ Consumer] No valid email addresses found in event: eventType=%s, taskId=%s, nodeId=%s, event=%+v",
			event.EventType, event.TaskID, event.NodeID, event)
		msg.Ack(false) // 确认消息，避免重复处理
		return
	}

	// 如果 Subject 或 Body 为空，根据事件类型和模板生成
	if event.Subject == "" || event.Body == "" {
		logx.Infof("[EmailMQ Consumer] Generating email content: eventType=%s", event.EventType)
		subject, body := generateEmailContent(ctx, svcCtx, &event)
		// 如果模板生成失败（返回空），不发送邮件
		if subject == "" || body == "" {
			logx.Infof("[EmailMQ Consumer] Email content generation failed or template service unavailable, skipping email: eventType=%s, subject=%s, bodyLength=%d",
				event.EventType, subject, len(body))
			msg.Ack(false) // 确认消息，避免重复处理
			return
		}
		if event.Subject == "" {
			event.Subject = subject
		}
		if event.Body == "" {
			event.Body = body
		}
		logx.Infof("[EmailMQ Consumer] Generated email content: subject=%s, bodyLength=%d", event.Subject, len(event.Body))
	}

	// 构建邮件消息
	emailMsg := middleware.EmailMessage{
		To:      emails,
		Subject: event.Subject,
		Body:    event.Body,
		IsHTML:  event.IsHTML,
	}

	logx.Infof("[EmailMQ Consumer] Attempting to send email: subject=%s, to=%v, isHTML=%v",
		event.Subject, emails, event.IsHTML)

	// 发送邮件
	err := svcCtx.EmailMiddleware.SendEmail(ctx, emailMsg)
	if err != nil {
		logx.Errorf("[EmailMQ Consumer] Failed to send email: error=%v, subject=%s, to=%v, eventType=%s, event=%+v",
			err, event.Subject, emails, event.EventType, event)
		// 邮件发送失败，可以选择重新入队或记录到死信队列
		// 这里先拒绝消息，让 RabbitMQ 重新投递（需要配置重试机制）
		msg.Nack(false, true) // 重新入队
		return
	}

	logx.Infof("[EmailMQ Consumer] Email sent successfully: subject=%s, to=%v, eventType=%s",
		event.Subject, emails, event.EventType)
	// 手动确认消息
	msg.Ack(false)
}

// resolveEmailRecipients 根据业务ID解析收件人邮箱
func resolveEmailRecipients(ctx context.Context, svcCtx *ServiceContext, event *EmailEvent) []string {
	emails := []string{}

	switch event.EventType {
	case "task.created":
		// 任务创建时，如果已经指定了To字段，直接使用；否则根据EmployeeIDs或TaskID查询
		if len(event.To) > 0 {
			// 已经指定了收件人，直接使用
			emails = event.To
		} else if len(event.EmployeeIDs) > 0 {
			// 根据EmployeeIDs查询邮箱
			for _, employeeID := range event.EmployeeIDs {
				employee, err := svcCtx.EmployeeModel.FindOne(ctx, employeeID)
				if err == nil && employee.Email.Valid && employee.Email.String != "" {
					emails = append(emails, employee.Email.String)
				}
			}
		} else if event.TaskID != "" {
			// 根据TaskID查询节点负责人
			taskInfo, err := svcCtx.TaskModel.FindOne(ctx, event.TaskID)
			if err == nil && taskInfo.NodeEmployeeIds.Valid && taskInfo.NodeEmployeeIds.String != "" {
				nodeEmployeeIDs := strings.Split(taskInfo.NodeEmployeeIds.String, ",")
				emailSet := make(map[string]bool)
				for _, employeeID := range nodeEmployeeIDs {
					if employeeID == "" {
						continue
					}
					employee, err := svcCtx.EmployeeModel.FindOne(ctx, employeeID)
					if err == nil && employee.Email.Valid && employee.Email.String != "" {
						if !emailSet[employee.Email.String] {
							emails = append(emails, employee.Email.String)
							emailSet[employee.Email.String] = true
						}
					}
				}
			}
		}
	case "task.completed", "task.updated":
		// 根据 TaskID 查询任务创建者和所有节点负责人/执行人的邮箱
		if event.TaskID != "" {
			taskInfo, err := svcCtx.TaskModel.FindOne(ctx, event.TaskID)
			if err == nil {
				// 查询任务创建者
				if taskInfo.TaskCreator != "" {
					creator, err := svcCtx.EmployeeModel.FindOne(ctx, taskInfo.TaskCreator)
					if err == nil && creator.Email.Valid && creator.Email.String != "" {
						emails = append(emails, creator.Email.String)
					}
				}
				// 查询所有节点
				nodes, err := svcCtx.TaskNodeModel.FindByTaskID(ctx, event.TaskID)
				if err == nil {
					emailSet := make(map[string]bool)
					for _, node := range nodes {
						if node.LeaderId != "" {
							leader, err := svcCtx.EmployeeModel.FindOne(ctx, node.LeaderId)
							if err == nil && leader.Email.Valid && leader.Email.String != "" {
								if !emailSet[leader.Email.String] {
									emails = append(emails, leader.Email.String)
									emailSet[leader.Email.String] = true
								}
							}
						}
						if node.ExecutorId != "" {
							executor, err := svcCtx.EmployeeModel.FindOne(ctx, node.ExecutorId)
							if err == nil && executor.Email.Valid && executor.Email.String != "" {
								if !emailSet[executor.Email.String] {
									emails = append(emails, executor.Email.String)
									emailSet[executor.Email.String] = true
								}
							}
						}
					}
				}
			}
		}
	case "task.node.executor.changed":
		// 根据 NodeID 查询新执行人邮箱
		if event.NodeID != "" {
			node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
			if err == nil && node.ExecutorId != "" {
				executor, err := svcCtx.EmployeeModel.FindOne(ctx, node.ExecutorId)
				if err == nil && executor.Email.Valid && executor.Email.String != "" {
					emails = append(emails, executor.Email.String)
				}
			}
		}
	case "task.deadline.reminder", "task.slow.progress", "task.node.executor.left":
		// 根据 NodeID 查询节点负责人或执行人邮箱
		if event.NodeID != "" {
			node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
			if err == nil {
				// 对于截止提醒和进度缓慢，通知执行人
				if event.EventType == "task.deadline.reminder" || event.EventType == "task.slow.progress" {
					if node.ExecutorId != "" {
						executor, err := svcCtx.EmployeeModel.FindOne(ctx, node.ExecutorId)
						if err == nil && executor.Email.Valid && executor.Email.String != "" {
							emails = append(emails, executor.Email.String)
						}
					}
					// 进度缓慢也通知负责人
					if event.EventType == "task.slow.progress" && node.LeaderId != "" {
						leader, err := svcCtx.EmployeeModel.FindOne(ctx, node.LeaderId)
						if err == nil && leader.Email.Valid && leader.Email.String != "" {
							emails = append(emails, leader.Email.String)
						}
					}
				} else if event.EventType == "task.node.executor.left" {
					// 执行人离职，通知负责人
					if event.EmployeeID != "" {
						leader, err := svcCtx.EmployeeModel.FindOne(ctx, event.EmployeeID)
						if err == nil && leader.Email.Valid && leader.Email.String != "" {
							emails = append(emails, leader.Email.String)
						}
					} else if node.LeaderId != "" {
						leader, err := svcCtx.EmployeeModel.FindOne(ctx, node.LeaderId)
						if err == nil && leader.Email.Valid && leader.Email.String != "" {
							emails = append(emails, leader.Email.String)
						}
					}
				}
			}
		}
	case "daily.report.reminder", "employee.direct":
		// 直接指定员工ID
		if event.EmployeeID != "" {
			employee, err := svcCtx.EmployeeModel.FindOne(ctx, event.EmployeeID)
			if err == nil && employee.Email.Valid && employee.Email.String != "" {
				emails = append(emails, employee.Email.String)
			}
		}
	case "handover.notification":
		// 根据 RelatedID 查询交接记录，获取发起人和接收人邮箱
		if event.RelatedID != "" {
			handover, err := svcCtx.TaskHandoverModel.FindOne(ctx, event.RelatedID)
			if err == nil {
				// 如果指定了 EmployeeIDs，优先使用指定的
				if len(event.EmployeeIDs) > 0 {
					for _, employeeID := range event.EmployeeIDs {
						employee, err := svcCtx.EmployeeModel.FindOne(ctx, employeeID)
						if err == nil && employee.Email.Valid && employee.Email.String != "" {
							emails = append(emails, employee.Email.String)
						}
					}
				} else {
					// 否则根据交接记录查询
					if handover.FromEmployeeId != "" {
						fromEmployee, err := svcCtx.EmployeeModel.FindOne(ctx, handover.FromEmployeeId)
						if err == nil && fromEmployee.Email.Valid && fromEmployee.Email.String != "" {
							emails = append(emails, fromEmployee.Email.String)
						}
					}
					if handover.ToEmployeeId != "" {
						toEmployee, err := svcCtx.EmployeeModel.FindOne(ctx, handover.ToEmployeeId)
						if err == nil && toEmployee.Email.Valid && toEmployee.Email.String != "" {
							emails = append(emails, toEmployee.Email.String)
						}
					}
				}
			}
		}
	}

	return emails
}

// generateEmailContent 根据事件类型生成邮件内容
func generateEmailContent(ctx context.Context, svcCtx *ServiceContext, event *EmailEvent) (subject, body string) {
	if svcCtx.EmailService == nil {
		return event.Subject, event.Body
	}

	switch event.EventType {
	case "task.created":
		// 任务创建邮件内容生成
		if event.Subject != "" && event.Body != "" {
			// 如果已经指定了主题和内容，直接使用
			return event.Subject, event.Body
		}
		// 否则使用模板生成
		if event.TaskID != "" && svcCtx.EmailTemplateService != nil {
			taskInfo, err := svcCtx.TaskModel.FindOne(ctx, event.TaskID)
			if err == nil {
				data := TaskCreatedData{
					TaskTitle: taskInfo.TaskTitle,
					TaskID:    event.TaskID,
					Year:      time.Now().Year(),
				}
				body, err := svcCtx.EmailTemplateService.RenderTemplate("task_created", data)
				if err == nil {
					return "新任务创建通知", body
				}
			}
		}
		// 模板不可用，使用默认内容
		if event.Body != "" {
			return event.Subject, event.Body
		}
		return "新任务创建通知", "您有新的任务需要处理，请登录系统查看详情。"
	case "task.completed":
		if event.TaskID != "" {
			taskInfo, err := svcCtx.TaskModel.FindOne(ctx, event.TaskID)
			if err == nil {
				creator, err := svcCtx.EmployeeModel.FindOne(ctx, taskInfo.TaskCreator)
				if err == nil && creator.Email.Valid && creator.Email.String != "" {
					message := fmt.Sprintf("任务 %s 已完成", taskInfo.TaskTitle)
					err := svcCtx.EmailService.SendHandoverEmail(ctx, creator.Email.String, creator.RealName, message, event.TaskID, taskInfo.TaskTitle)
					if err == nil {
						// 邮件已通过 EmailService 发送，这里返回空，避免重复发送
						return "", ""
					}
				}
			}
		}
	case "task.updated":
		if event.TaskID != "" {
			taskInfo, err := svcCtx.TaskModel.FindOne(ctx, event.TaskID)
			if err == nil {
				// 使用模板生成邮件内容
				if svcCtx.EmailTemplateService != nil {
					updateTime := time.Now().Format("2006-01-02 15:04:05")
					data := TaskUpdatedData{
						TaskTitle:  taskInfo.TaskTitle,
						TaskID:     event.TaskID,
						UpdateTime: updateTime,
						UpdateNote: "",
						Year:       time.Now().Year(),
					}
					body, err := svcCtx.EmailTemplateService.RenderTemplate("task_updated", data)
					if err == nil {
						return "任务更新通知", body
					}
				}
				// 模板服务不可用，不发送邮件
				return "", ""
			}
		}
	case "task.node.executor.changed":
		if event.NodeID != "" {
			node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
			if err == nil && node.ExecutorId != "" {
				executor, err := svcCtx.EmployeeModel.FindOne(ctx, node.ExecutorId)
				if err == nil && executor.Email.Valid && executor.Email.String != "" {
					err := svcCtx.EmailService.SendHandoverEmail(ctx, executor.Email.String, executor.RealName, "您已被分配为新的任务节点执行人", event.NodeID, "")
					if err == nil {
						// 邮件已通过 EmailService 发送，这里返回空，避免重复发送
						return "", ""
					}
				}
			}
		}
	case "task.deadline.reminder":
		if event.NodeID != "" {
			node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
			if err == nil {
				deadline := node.NodeDeadline.Format("2006-01-02 15:04:05")
				if svcCtx.EmailService != nil {
					executor, err := svcCtx.EmployeeModel.FindOne(ctx, node.ExecutorId)
					if err == nil && executor.Email.Valid && executor.Email.String != "" {
						err := svcCtx.EmailService.SendTaskDeadlineReminderEmail(ctx, executor.Email.String, node.NodeName, deadline, int(node.Progress))
						if err == nil {
							return "", ""
						}
					}
				}
				// 如果 EmailService 不可用，使用模板生成内容
				if svcCtx.EmailTemplateService != nil {
					data := TaskDeadlineReminderData{
						NodeName: node.NodeName,
						Deadline: deadline,
						Progress: int(node.Progress),
						Year:     time.Now().Year(),
					}
					body, err := svcCtx.EmailTemplateService.RenderTemplate("task_deadline_reminder", data)
					if err == nil {
						return "任务截止时间提醒", body
					}
				}
				// 模板服务不可用，不发送邮件
				return "", ""
			}
		}
	case "task.slow.progress":
		if event.NodeID != "" {
			node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
			if err == nil {
				// 使用模板生成邮件内容
				if svcCtx.EmailTemplateService != nil {
					data := TaskSlowProgressData{
						NodeName: node.NodeName,
						Progress: int(node.Progress),
						Deadline: node.NodeDeadline.Format("2006-01-02 15:04:05"),
						Year:     time.Now().Year(),
					}
					body, err := svcCtx.EmailTemplateService.RenderTemplate("task_slow_progress", data)
					if err == nil {
						return "任务进度缓慢提醒", body
					}
				}
				// 模板服务不可用，不发送邮件
				return "", ""
			}
		}
	case "daily.report.reminder":
		if event.EmployeeID != "" {
			employee, err := svcCtx.EmployeeModel.FindOne(ctx, event.EmployeeID)
			if err == nil {
				// 使用模板生成邮件内容
				if svcCtx.EmailTemplateService != nil {
					data := DailyReportReminderData{
						EmployeeName: employee.RealName,
						Year:         time.Now().Year(),
					}
					body, err := svcCtx.EmailTemplateService.RenderTemplate("daily_report_reminder", data)
					if err == nil {
						return "每日工作报告提醒", body
					}
				}
				// 模板服务不可用，不发送邮件
				return "", ""
			}
		}
	case "task.node.executor.left":
		if event.NodeID != "" {
			node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
			if err == nil {
				taskInfo, err := svcCtx.TaskModel.FindOne(ctx, node.TaskId)
				if err == nil {
					// 使用模板生成邮件内容
					if svcCtx.EmailTemplateService != nil {
						leftEmployeeName := ""
						if node.ExecutorId != "" {
							leftEmployee, err := svcCtx.EmployeeModel.FindOne(ctx, node.ExecutorId)
							if err == nil {
								leftEmployeeName = leftEmployee.RealName
							}
						}
						data := TaskNodeExecutorLeftData{
							TaskTitle:        taskInfo.TaskTitle,
							NodeName:         node.NodeName,
							NodeDetail:       node.NodeDetail.String,
							LeftEmployeeName: leftEmployeeName,
							Year:             time.Now().Year(),
						}
						body, err := svcCtx.EmailTemplateService.RenderTemplate("task_node_executor_left", data)
						if err == nil {
							return fmt.Sprintf("任务节点执行人离职通知 - %s", taskInfo.TaskTitle), body
						}
					}
					// 模板服务不可用，不发送邮件
					return "", ""
				}
			}
		}
	case "handover.notification":
		// 根据 RelatedID 查询交接记录，生成邮件内容
		if event.RelatedID != "" {
			handover, err := svcCtx.TaskHandoverModel.FindOne(ctx, event.RelatedID)
			if err == nil {
				taskTitle := ""
				if handover.TaskId != "" {
					taskInfo, err := svcCtx.TaskModel.FindOne(ctx, handover.TaskId)
					if err == nil {
						taskTitle = taskInfo.TaskTitle
					}
				}
				// 使用模板生成邮件内容
				if svcCtx.EmailTemplateService != nil {
					// 根据接收人和发起人确定消息内容
					message := "任务交接通知"
					if len(event.EmployeeIDs) > 0 {
						if event.EmployeeIDs[0] == handover.ToEmployeeId {
							// 发给接收人
							message = fmt.Sprintf("您收到了一个任务交接请求：%s", taskTitle)
						} else if event.EmployeeIDs[0] == handover.FromEmployeeId {
							// 发给发起人
							if handover.HandoverStatus == 2 {
								message = "任务交接完成"
							} else if handover.HandoverStatus == 3 {
								message = "任务交接已拒绝"
							} else if handover.HandoverStatus == 4 {
								message = "交接已确认"
							}
						}
					}
					// 获取员工姓名
					employeeName := ""
					if len(event.EmployeeIDs) > 0 {
						employee, err := svcCtx.EmployeeModel.FindOne(ctx, event.EmployeeIDs[0])
						if err == nil {
							employeeName = employee.RealName
						}
					}
					data := HandoverData{
						EmployeeName: employeeName,
						Message:      message,
						HandoverID:   event.RelatedID,
						TaskTitle:    taskTitle,
						Year:         time.Now().Year(),
					}
					body, err := svcCtx.EmailTemplateService.RenderTemplate("handover", data)
					if err == nil {
						return "任务交接通知", body
					}
				}
				// 模板服务不可用，不发送邮件
				return "", ""
			}
		}
	}

	// 如果无法通过模板生成，返回原始值
	return event.Subject, event.Body
}
