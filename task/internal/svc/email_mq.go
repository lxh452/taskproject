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
	EventType    string                 `json:"eventType"`
	To           []string               `json:"to"`
	Subject      string                 `json:"subject"`
	Body         string                 `json:"body"`
	IsHTML       bool                   `json:"isHtml"`
	TaskID       string                 `json:"taskId"`
	NodeID       string                 `json:"nodeId"`
	EmployeeID   string                 `json:"employeeId"`
	EmployeeIDs  []string               `json:"employeeIds"`
	RelatedID    string                 `json:"relatedId"`
	TemplateData map[string]interface{} `json:"templateData"`
}

// EmailMQService 邮件消息队列服务
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
		return nil
	}

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

// StartEmailConsumer 启动邮件消费者
func StartEmailConsumer(mqClient *MQClient, queueName string, svcCtx *ServiceContext) error {
	if mqClient == nil {
		return fmt.Errorf("MQClient is nil")
	}

	queue, err := mqClient.GetChannel().QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = mqClient.GetChannel().QueueBind(
		queue.Name,
		"email.*",
		mqClient.GetExchange(),
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	msgs, err := mqClient.GetChannel().Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	logx.Infof("[EmailMQ Consumer] Started consuming messages from queue: %s", queue.Name)

	go func() {
		for msg := range msgs {
			processEmailMessage(context.Background(), msg, svcCtx)
		}
	}()

	return nil
}

// processEmailMessage 处理邮件消息
func processEmailMessage(ctx context.Context, msg amqp.Delivery, svcCtx *ServiceContext) {
	logx.WithContext(ctx).Infof("[EmailMQ Consumer] Received message: %s", msg.Body)

	var event EmailEvent
	err := json.Unmarshal(msg.Body, &event)
	if err != nil {
		logx.WithContext(ctx).Errorf("[EmailMQ Consumer] Failed to unmarshal email event: error=%v, body=%s",
			err, string(msg.Body))
		msg.Nack(false, false)
		return
	}

	logx.WithContext(ctx).Infof("[EmailMQ Consumer] Processing email event: eventType=%s, to=%v, taskId=%s, nodeId=%s",
		event.EventType, event.To, event.TaskID, event.NodeID)

	emails := event.To
	if len(emails) == 0 {
		logx.Infof("[EmailMQ Consumer] To field is empty, resolving recipients from business IDs")
		emails = resolveEmailRecipients(ctx, svcCtx, &event)
		logx.Infof("[EmailMQ Consumer] Resolved %d email addresses: %v", len(emails), emails)
	}

	if len(emails) == 0 {
		logx.Errorf("[EmailMQ Consumer] No valid email addresses found in event: eventType=%s, taskId=%s, nodeId=%s, event=%+v",
			event.EventType, event.TaskID, event.NodeID, event)
		msg.Ack(false)
		return
	}

	if svcCtx.UnifiedEmailService != nil {
		logx.Infof("[EmailMQ Consumer] Using unified email service to send email: eventType=%s", event.EventType)

		templateData := buildTemplateData(&event)

		if event.Subject == "" {
			event.Subject = generateSubjectByEventType(event.EventType, templateData)
		}

		err := svcCtx.UnifiedEmailService.SendTaskNotification(ctx, event.EventType, emails, templateData)
		if err != nil {
			logx.Errorf("[EmailMQ Consumer] Failed to send email using unified service: error=%v, eventType=%s, to=%v",
				err, event.EventType, emails)
			msg.Nack(false, true)
			return
		}

		logx.Infof("[EmailMQ Consumer] Email sent successfully using unified service: eventType=%s, to=%v",
			event.EventType, emails)
	} else {
		logx.Infof("[EmailMQ Consumer] UnifiedEmailService is not available, using legacy email middleware")

		emailMsg := middleware.EmailMessage{
			To:      emails,
			Subject: event.Subject,
			Body:    event.Body,
			IsHTML:  event.IsHTML,
		}

		logx.Infof("[EmailMQ Consumer] Attempting to send email using legacy middleware: subject=%s, to=%v, isHTML=%v",
			event.Subject, emails, event.IsHTML)

		err := svcCtx.EmailMiddleware.SendEmail(ctx, emailMsg)
		if err != nil {
			logx.Errorf("[EmailMQ Consumer] Failed to send email: error=%v, subject=%s, to=%v, eventType=%s, event=%+v",
				err, event.Subject, emails, event.EventType, event)
			msg.Nack(false, true)
			return
		}

		logx.Infof("[EmailMQ Consumer] Email sent successfully using legacy middleware: subject=%s, to=%v, eventType=%s",
			event.Subject, emails, event.EventType)
	}

	msg.Ack(false)
}

// buildTemplateData 构建模板数据
func buildTemplateData(event *EmailEvent) map[string]interface{} {
	data := map[string]interface{}{
		"taskId":     event.TaskID,
		"nodeId":     event.NodeID,
		"employeeId": event.EmployeeID,
		"relatedId":  event.RelatedID,
		"eventType":  event.EventType,
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
	}

	if event.TemplateData != nil {
		for k, v := range event.TemplateData {
			data[k] = v
		}
	}

	return data
}

// generateSubjectByEventType 根据事件类型生成邮件主题
func generateSubjectByEventType(eventType string, data map[string]interface{}) string {
	subject := "系统通知"

	if strings.HasPrefix(eventType, "task.") {
		subject = "任务通知"
		if taskTitle, ok := data["taskTitle"].(string); ok && taskTitle != "" {
			subject = fmt.Sprintf("任务通知 - %s", taskTitle)
		}
	} else if strings.HasPrefix(eventType, "system.") || strings.HasPrefix(eventType, "login.") || strings.HasPrefix(eventType, "register.") {
		subject = "系统通知"
	}

	return subject
}

// resolveEmailRecipients 根据业务ID解析收件人邮箱
func resolveEmailRecipients(ctx context.Context, svcCtx *ServiceContext, event *EmailEvent) []string {
	var emails []string
	emailSet := make(map[string]bool)

	addEmail := func(email string) {
		if email != "" && !emailSet[email] {
			emails = append(emails, email)
			emailSet[email] = true
		}
	}

	for _, email := range event.To {
		addEmail(email)
	}

	if len(event.EmployeeIDs) > 0 {
		for _, employeeID := range event.EmployeeIDs {
			employee, err := svcCtx.EmployeeModel.FindOne(ctx, employeeID)
			if err == nil && employee.Email.Valid && employee.Email.String != "" {
				addEmail(employee.Email.String)
			}
		}
	}

	if event.EmployeeID != "" {
		employee, err := svcCtx.EmployeeModel.FindOne(ctx, event.EmployeeID)
		if err == nil && employee.Email.Valid && employee.Email.String != "" {
			addEmail(employee.Email.String)
		}
	}

	if event.TaskID != "" {
		task, err := svcCtx.TaskModel.FindOne(ctx, event.TaskID)
		if err == nil && task != nil {
			if task.TaskCreator != "" {
				creator, err := svcCtx.EmployeeModel.FindOne(ctx, task.TaskCreator)
				if err == nil && creator.Email.Valid && creator.Email.String != "" {
					addEmail(creator.Email.String)
				}
			}

			if task.LeaderId.Valid && task.LeaderId.String != "" {
				leader, err := svcCtx.EmployeeModel.FindOne(ctx, task.LeaderId.String)
				if err == nil && leader.Email.Valid && leader.Email.String != "" {
					addEmail(leader.Email.String)
				}
			}

			if task.NodeEmployeeIds.Valid && task.NodeEmployeeIds.String != "" {
				nodeEmployeeIDs := strings.Split(task.NodeEmployeeIds.String, ",")
				for _, employeeID := range nodeEmployeeIDs {
					employeeID = strings.TrimSpace(employeeID)
					if employeeID == "" {
						continue
					}
					employee, err := svcCtx.EmployeeModel.FindOne(ctx, employeeID)
					if err == nil && employee.Email.Valid && employee.Email.String != "" {
						addEmail(employee.Email.String)
					}
				}
			}

			nodes, err := svcCtx.TaskNodeModel.FindByTaskID(ctx, event.TaskID)
			if err == nil {
				for _, node := range nodes {
					if node.LeaderId != "" {
						leader, err := svcCtx.EmployeeModel.FindOne(ctx, node.LeaderId)
						if err == nil && leader.Email.Valid && leader.Email.String != "" {
							addEmail(leader.Email.String)
						}
					}
					if node.ExecutorId != "" {
						executorIds := strings.Split(node.ExecutorId, ",")
						for _, eid := range executorIds {
							eid = strings.TrimSpace(eid)
							if eid != "" {
								executor, err := svcCtx.EmployeeModel.FindOne(ctx, eid)
								if err == nil && executor.Email.Valid && executor.Email.String != "" {
									addEmail(executor.Email.String)
								}
							}
						}
					}
				}
			}
		}
	}

	if event.NodeID != "" {
		node, err := svcCtx.TaskNodeModel.FindOne(ctx, event.NodeID)
		if err == nil && node != nil {
			if node.ExecutorId != "" {
				executorIds := strings.Split(node.ExecutorId, ",")
				for _, eid := range executorIds {
					eid = strings.TrimSpace(eid)
					if eid != "" {
						executor, err := svcCtx.EmployeeModel.FindOne(ctx, eid)
						if err == nil && executor.Email.Valid && executor.Email.String != "" {
							addEmail(executor.Email.String)
						}
					}
				}
			}

			if node.LeaderId != "" {
				leader, err := svcCtx.EmployeeModel.FindOne(ctx, node.LeaderId)
				if err == nil && leader.Email.Valid && leader.Email.String != "" {
					addEmail(leader.Email.String)
				}
			}
		}
	}

	if event.RelatedID != "" {
		handover, err := svcCtx.TaskHandoverModel.FindOne(ctx, event.RelatedID)
		if err == nil {
			if handover.FromEmployeeId != "" {
				fromEmployee, err := svcCtx.EmployeeModel.FindOne(ctx, handover.FromEmployeeId)
				if err == nil && fromEmployee.Email.Valid && fromEmployee.Email.String != "" {
					addEmail(fromEmployee.Email.String)
				}
			}
			if handover.ToEmployeeId != "" {
				toEmployee, err := svcCtx.EmployeeModel.FindOne(ctx, handover.ToEmployeeId)
				if err == nil && toEmployee.Email.Valid && toEmployee.Email.String != "" {
					addEmail(toEmployee.Email.String)
				}
			}
		}
	}

	return emails
}
