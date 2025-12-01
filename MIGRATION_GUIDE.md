# 迁移指南：从 notification.go 迁移到消息队列和模板系统

## 概述

本次重构将：
1. **删除** `task/internal/svc/notification.go` 文件
2. **使用** 邮件模板系统（`.tpl` 文件）生成邮件内容
3. **统一** 通过消息队列（RabbitMQ）异步处理所有邮件和通知

## 架构变化

### 之前
```
业务逻辑 → NotificationService → EmailMiddleware.SendEmail (同步)
业务逻辑 → NotificationService → NotificationModel.Insert (同步)
```

### 现在
```
业务逻辑 → EmailService (模板渲染) → EmailMQService → RabbitMQ → 消费者 → EmailMiddleware.SendEmail (异步)
业务逻辑 → NotificationMQService → RabbitMQ → 消费者 → NotificationModel.Insert (异步)
```

## 需要修改的文件

### 1. Logic 层文件（需要修改）

#### task/internal/logic/tasknode/updateTaskNodeLogic.go ✅ 已修改
- 替换 `NotificationService.SendHandoverNotification` 
- 使用 `EmailService.SendHandoverEmail` + `NotificationMQService.PublishNotificationEvent`

#### task/internal/logic/task/updateTaskProgressLogic.go
```go
// 旧代码
notificationService := svc.NewNotificationService(l.svcCtx)
notificationService.SendTaskCompletionNotification(...)

// 新代码
// 发送邮件
if l.svcCtx.EmailService != nil {
    l.svcCtx.EmailService.SendTaskCompletedEmail(...)
}
// 创建通知
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{...}
    l.svcCtx.NotificationMQService.PublishNotificationEvent(...)
}
```

#### task/internal/logic/task/createTaskLogic.go
- 替换所有 `NotificationService` 调用
- 使用 `EmailService` + `NotificationMQService`

#### task/internal/logic/employee/employeeLeaveLogic.go
- 替换 `SendEmployeeLeaveNotification`
- 使用模板渲染 + 消息队列

#### task/internal/logic/handover/*.go (3个文件)
- 替换 `SendHandoverNotification`
- 使用 `EmailService.SendHandoverEmail` + `NotificationMQService`

#### task/internal/logic/task/completeTaskLogic.go
- 替换 `SendHandoverNotification`
- 使用消息队列

#### task/internal/logic/task/deleteTaskLogic.go
- 替换 `SendTaskDeletedNotification`
- 使用消息队列

#### task/internal/logic/tasknode/deleteTaskNodeLogic.go
- 替换 `SendTaskNodeDeletedNotification`
- 使用消息队列

### 2. Scheduler 文件 ✅ 已部分修改

#### task/internal/svc/scheduler.go
- ✅ `SendTaskDeadlineReminder` 已改为消息队列
- 需要修改其他定时任务中的通知发送

### 3. 删除文件

- `task/internal/svc/notification.go` - 删除整个文件

## 迁移模式

### 模式1：发送邮件 + 创建通知

```go
// 1. 发送邮件（通过消息队列）
if l.svcCtx.EmailService != nil {
    err := l.svcCtx.EmailService.SendHandoverEmail(
        ctx,
        employeeEmail,
        employeeName,
        message,
        handoverID,
        taskTitle,
    )
    if err != nil {
        logx.Errorf("发送邮件失败: %v", err)
    }
}

// 2. 创建系统通知（通过消息队列）
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{
        EventType:   "handover.notification",
        EmployeeIDs: []string{employeeID},
        Title:       "任务交接通知",
        Content:     message,
        Type:        3,
        Category:    "handover",
        Priority:    2,
        RelatedID:   handoverID,
        RelatedType: "handover",
    }
    if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
        logx.Errorf("发布通知事件失败: %v", err)
    }
}
```

### 模式2：只发送邮件

```go
if l.svcCtx.EmailService != nil {
    err := l.svcCtx.EmailService.SendTaskDispatchEmail(
        ctx,
        employeeEmail,
        employeeName,
        taskTitle,
        nodeName,
        nodeDetail,
        deadline,
        taskId,
    )
    if err != nil {
        logx.Errorf("发送邮件失败: %v", err)
    }
}
```

### 模式3：只创建通知

```go
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{
        EventType:   "task.detail.updated",
        EmployeeIDs: employeeIDs,
        Title:       "任务详情更新",
        Content:     "任务详情已更新",
        Type:        1,
        Category:    "task",
        Priority:    1,
        RelatedID:   taskId,
        RelatedType: "task",
    }
    if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
        logx.Errorf("发布通知事件失败: %v", err)
    }
}
```

## EmailService 可用方法

- `SendTaskDispatchEmail` - 任务派发邮件
- `SendTaskDeadlineReminderEmail` - 任务截止提醒
- `SendTaskCompletedEmail` - 任务完成邮件
- `SendHandoverEmail` - 交接邮件

## 注意事项

1. **所有操作都是异步的** - 发布到消息队列后立即返回，不等待处理完成
2. **错误处理** - 发布失败只记录日志，不中断主流程
3. **模板数据** - 确保传入的数据格式正确（参考 `email_template.go` 中的数据结构）
4. **消息队列未配置** - 如果 RabbitMQ 未配置，服务会记录警告但不会崩溃

## 完成迁移后

1. 删除 `task/internal/svc/notification.go`
2. 运行测试确保所有功能正常
3. 检查日志确认消息队列正常工作



## 概述

本次重构将：
1. **删除** `task/internal/svc/notification.go` 文件
2. **使用** 邮件模板系统（`.tpl` 文件）生成邮件内容
3. **统一** 通过消息队列（RabbitMQ）异步处理所有邮件和通知

## 架构变化

### 之前
```
业务逻辑 → NotificationService → EmailMiddleware.SendEmail (同步)
业务逻辑 → NotificationService → NotificationModel.Insert (同步)
```

### 现在
```
业务逻辑 → EmailService (模板渲染) → EmailMQService → RabbitMQ → 消费者 → EmailMiddleware.SendEmail (异步)
业务逻辑 → NotificationMQService → RabbitMQ → 消费者 → NotificationModel.Insert (异步)
```

## 需要修改的文件

### 1. Logic 层文件（需要修改）

#### task/internal/logic/tasknode/updateTaskNodeLogic.go ✅ 已修改
- 替换 `NotificationService.SendHandoverNotification` 
- 使用 `EmailService.SendHandoverEmail` + `NotificationMQService.PublishNotificationEvent`

#### task/internal/logic/task/updateTaskProgressLogic.go
```go
// 旧代码
notificationService := svc.NewNotificationService(l.svcCtx)
notificationService.SendTaskCompletionNotification(...)

// 新代码
// 发送邮件
if l.svcCtx.EmailService != nil {
    l.svcCtx.EmailService.SendTaskCompletedEmail(...)
}
// 创建通知
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{...}
    l.svcCtx.NotificationMQService.PublishNotificationEvent(...)
}
```

#### task/internal/logic/task/createTaskLogic.go
- 替换所有 `NotificationService` 调用
- 使用 `EmailService` + `NotificationMQService`

#### task/internal/logic/employee/employeeLeaveLogic.go
- 替换 `SendEmployeeLeaveNotification`
- 使用模板渲染 + 消息队列

#### task/internal/logic/handover/*.go (3个文件)
- 替换 `SendHandoverNotification`
- 使用 `EmailService.SendHandoverEmail` + `NotificationMQService`

#### task/internal/logic/task/completeTaskLogic.go
- 替换 `SendHandoverNotification`
- 使用消息队列

#### task/internal/logic/task/deleteTaskLogic.go
- 替换 `SendTaskDeletedNotification`
- 使用消息队列

#### task/internal/logic/tasknode/deleteTaskNodeLogic.go
- 替换 `SendTaskNodeDeletedNotification`
- 使用消息队列

### 2. Scheduler 文件 ✅ 已部分修改

#### task/internal/svc/scheduler.go
- ✅ `SendTaskDeadlineReminder` 已改为消息队列
- 需要修改其他定时任务中的通知发送

### 3. 删除文件

- `task/internal/svc/notification.go` - 删除整个文件

## 迁移模式

### 模式1：发送邮件 + 创建通知

```go
// 1. 发送邮件（通过消息队列）
if l.svcCtx.EmailService != nil {
    err := l.svcCtx.EmailService.SendHandoverEmail(
        ctx,
        employeeEmail,
        employeeName,
        message,
        handoverID,
        taskTitle,
    )
    if err != nil {
        logx.Errorf("发送邮件失败: %v", err)
    }
}

// 2. 创建系统通知（通过消息队列）
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{
        EventType:   "handover.notification",
        EmployeeIDs: []string{employeeID},
        Title:       "任务交接通知",
        Content:     message,
        Type:        3,
        Category:    "handover",
        Priority:    2,
        RelatedID:   handoverID,
        RelatedType: "handover",
    }
    if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
        logx.Errorf("发布通知事件失败: %v", err)
    }
}
```

### 模式2：只发送邮件

```go
if l.svcCtx.EmailService != nil {
    err := l.svcCtx.EmailService.SendTaskDispatchEmail(
        ctx,
        employeeEmail,
        employeeName,
        taskTitle,
        nodeName,
        nodeDetail,
        deadline,
        taskId,
    )
    if err != nil {
        logx.Errorf("发送邮件失败: %v", err)
    }
}
```

### 模式3：只创建通知

```go
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{
        EventType:   "task.detail.updated",
        EmployeeIDs: employeeIDs,
        Title:       "任务详情更新",
        Content:     "任务详情已更新",
        Type:        1,
        Category:    "task",
        Priority:    1,
        RelatedID:   taskId,
        RelatedType: "task",
    }
    if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
        logx.Errorf("发布通知事件失败: %v", err)
    }
}
```

## EmailService 可用方法

- `SendTaskDispatchEmail` - 任务派发邮件
- `SendTaskDeadlineReminderEmail` - 任务截止提醒
- `SendTaskCompletedEmail` - 任务完成邮件
- `SendHandoverEmail` - 交接邮件

## 注意事项

1. **所有操作都是异步的** - 发布到消息队列后立即返回，不等待处理完成
2. **错误处理** - 发布失败只记录日志，不中断主流程
3. **模板数据** - 确保传入的数据格式正确（参考 `email_template.go` 中的数据结构）
4. **消息队列未配置** - 如果 RabbitMQ 未配置，服务会记录警告但不会崩溃

## 完成迁移后

1. 删除 `task/internal/svc/notification.go`
2. 运行测试确保所有功能正常
3. 检查日志确认消息队列正常工作



## 概述

本次重构将：
1. **删除** `task/internal/svc/notification.go` 文件
2. **使用** 邮件模板系统（`.tpl` 文件）生成邮件内容
3. **统一** 通过消息队列（RabbitMQ）异步处理所有邮件和通知

## 架构变化

### 之前
```
业务逻辑 → NotificationService → EmailMiddleware.SendEmail (同步)
业务逻辑 → NotificationService → NotificationModel.Insert (同步)
```

### 现在
```
业务逻辑 → EmailService (模板渲染) → EmailMQService → RabbitMQ → 消费者 → EmailMiddleware.SendEmail (异步)
业务逻辑 → NotificationMQService → RabbitMQ → 消费者 → NotificationModel.Insert (异步)
```

## 需要修改的文件

### 1. Logic 层文件（需要修改）

#### task/internal/logic/tasknode/updateTaskNodeLogic.go ✅ 已修改
- 替换 `NotificationService.SendHandoverNotification` 
- 使用 `EmailService.SendHandoverEmail` + `NotificationMQService.PublishNotificationEvent`

#### task/internal/logic/task/updateTaskProgressLogic.go
```go
// 旧代码
notificationService := svc.NewNotificationService(l.svcCtx)
notificationService.SendTaskCompletionNotification(...)

// 新代码
// 发送邮件
if l.svcCtx.EmailService != nil {
    l.svcCtx.EmailService.SendTaskCompletedEmail(...)
}
// 创建通知
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{...}
    l.svcCtx.NotificationMQService.PublishNotificationEvent(...)
}
```

#### task/internal/logic/task/createTaskLogic.go
- 替换所有 `NotificationService` 调用
- 使用 `EmailService` + `NotificationMQService`

#### task/internal/logic/employee/employeeLeaveLogic.go
- 替换 `SendEmployeeLeaveNotification`
- 使用模板渲染 + 消息队列

#### task/internal/logic/handover/*.go (3个文件)
- 替换 `SendHandoverNotification`
- 使用 `EmailService.SendHandoverEmail` + `NotificationMQService`

#### task/internal/logic/task/completeTaskLogic.go
- 替换 `SendHandoverNotification`
- 使用消息队列

#### task/internal/logic/task/deleteTaskLogic.go
- 替换 `SendTaskDeletedNotification`
- 使用消息队列

#### task/internal/logic/tasknode/deleteTaskNodeLogic.go
- 替换 `SendTaskNodeDeletedNotification`
- 使用消息队列

### 2. Scheduler 文件 ✅ 已部分修改

#### task/internal/svc/scheduler.go
- ✅ `SendTaskDeadlineReminder` 已改为消息队列
- 需要修改其他定时任务中的通知发送

### 3. 删除文件

- `task/internal/svc/notification.go` - 删除整个文件

## 迁移模式

### 模式1：发送邮件 + 创建通知

```go
// 1. 发送邮件（通过消息队列）
if l.svcCtx.EmailService != nil {
    err := l.svcCtx.EmailService.SendHandoverEmail(
        ctx,
        employeeEmail,
        employeeName,
        message,
        handoverID,
        taskTitle,
    )
    if err != nil {
        logx.Errorf("发送邮件失败: %v", err)
    }
}

// 2. 创建系统通知（通过消息队列）
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{
        EventType:   "handover.notification",
        EmployeeIDs: []string{employeeID},
        Title:       "任务交接通知",
        Content:     message,
        Type:        3,
        Category:    "handover",
        Priority:    2,
        RelatedID:   handoverID,
        RelatedType: "handover",
    }
    if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
        logx.Errorf("发布通知事件失败: %v", err)
    }
}
```

### 模式2：只发送邮件

```go
if l.svcCtx.EmailService != nil {
    err := l.svcCtx.EmailService.SendTaskDispatchEmail(
        ctx,
        employeeEmail,
        employeeName,
        taskTitle,
        nodeName,
        nodeDetail,
        deadline,
        taskId,
    )
    if err != nil {
        logx.Errorf("发送邮件失败: %v", err)
    }
}
```

### 模式3：只创建通知

```go
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{
        EventType:   "task.detail.updated",
        EmployeeIDs: employeeIDs,
        Title:       "任务详情更新",
        Content:     "任务详情已更新",
        Type:        1,
        Category:    "task",
        Priority:    1,
        RelatedID:   taskId,
        RelatedType: "task",
    }
    if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
        logx.Errorf("发布通知事件失败: %v", err)
    }
}
```

## EmailService 可用方法

- `SendTaskDispatchEmail` - 任务派发邮件
- `SendTaskDeadlineReminderEmail` - 任务截止提醒
- `SendTaskCompletedEmail` - 任务完成邮件
- `SendHandoverEmail` - 交接邮件

## 注意事项

1. **所有操作都是异步的** - 发布到消息队列后立即返回，不等待处理完成
2. **错误处理** - 发布失败只记录日志，不中断主流程
3. **模板数据** - 确保传入的数据格式正确（参考 `email_template.go` 中的数据结构）
4. **消息队列未配置** - 如果 RabbitMQ 未配置，服务会记录警告但不会崩溃

## 完成迁移后

1. 删除 `task/internal/svc/notification.go`
2. 运行测试确保所有功能正常
3. 检查日志确认消息队列正常工作



## 概述

本次重构将：
1. **删除** `task/internal/svc/notification.go` 文件
2. **使用** 邮件模板系统（`.tpl` 文件）生成邮件内容
3. **统一** 通过消息队列（RabbitMQ）异步处理所有邮件和通知

## 架构变化

### 之前
```
业务逻辑 → NotificationService → EmailMiddleware.SendEmail (同步)
业务逻辑 → NotificationService → NotificationModel.Insert (同步)
```

### 现在
```
业务逻辑 → EmailService (模板渲染) → EmailMQService → RabbitMQ → 消费者 → EmailMiddleware.SendEmail (异步)
业务逻辑 → NotificationMQService → RabbitMQ → 消费者 → NotificationModel.Insert (异步)
```

## 需要修改的文件

### 1. Logic 层文件（需要修改）

#### task/internal/logic/tasknode/updateTaskNodeLogic.go ✅ 已修改
- 替换 `NotificationService.SendHandoverNotification` 
- 使用 `EmailService.SendHandoverEmail` + `NotificationMQService.PublishNotificationEvent`

#### task/internal/logic/task/updateTaskProgressLogic.go
```go
// 旧代码
notificationService := svc.NewNotificationService(l.svcCtx)
notificationService.SendTaskCompletionNotification(...)

// 新代码
// 发送邮件
if l.svcCtx.EmailService != nil {
    l.svcCtx.EmailService.SendTaskCompletedEmail(...)
}
// 创建通知
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{...}
    l.svcCtx.NotificationMQService.PublishNotificationEvent(...)
}
```

#### task/internal/logic/task/createTaskLogic.go
- 替换所有 `NotificationService` 调用
- 使用 `EmailService` + `NotificationMQService`

#### task/internal/logic/employee/employeeLeaveLogic.go
- 替换 `SendEmployeeLeaveNotification`
- 使用模板渲染 + 消息队列

#### task/internal/logic/handover/*.go (3个文件)
- 替换 `SendHandoverNotification`
- 使用 `EmailService.SendHandoverEmail` + `NotificationMQService`

#### task/internal/logic/task/completeTaskLogic.go
- 替换 `SendHandoverNotification`
- 使用消息队列

#### task/internal/logic/task/deleteTaskLogic.go
- 替换 `SendTaskDeletedNotification`
- 使用消息队列

#### task/internal/logic/tasknode/deleteTaskNodeLogic.go
- 替换 `SendTaskNodeDeletedNotification`
- 使用消息队列

### 2. Scheduler 文件 ✅ 已部分修改

#### task/internal/svc/scheduler.go
- ✅ `SendTaskDeadlineReminder` 已改为消息队列
- 需要修改其他定时任务中的通知发送

### 3. 删除文件

- `task/internal/svc/notification.go` - 删除整个文件

## 迁移模式

### 模式1：发送邮件 + 创建通知

```go
// 1. 发送邮件（通过消息队列）
if l.svcCtx.EmailService != nil {
    err := l.svcCtx.EmailService.SendHandoverEmail(
        ctx,
        employeeEmail,
        employeeName,
        message,
        handoverID,
        taskTitle,
    )
    if err != nil {
        logx.Errorf("发送邮件失败: %v", err)
    }
}

// 2. 创建系统通知（通过消息队列）
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{
        EventType:   "handover.notification",
        EmployeeIDs: []string{employeeID},
        Title:       "任务交接通知",
        Content:     message,
        Type:        3,
        Category:    "handover",
        Priority:    2,
        RelatedID:   handoverID,
        RelatedType: "handover",
    }
    if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
        logx.Errorf("发布通知事件失败: %v", err)
    }
}
```

### 模式2：只发送邮件

```go
if l.svcCtx.EmailService != nil {
    err := l.svcCtx.EmailService.SendTaskDispatchEmail(
        ctx,
        employeeEmail,
        employeeName,
        taskTitle,
        nodeName,
        nodeDetail,
        deadline,
        taskId,
    )
    if err != nil {
        logx.Errorf("发送邮件失败: %v", err)
    }
}
```

### 模式3：只创建通知

```go
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{
        EventType:   "task.detail.updated",
        EmployeeIDs: employeeIDs,
        Title:       "任务详情更新",
        Content:     "任务详情已更新",
        Type:        1,
        Category:    "task",
        Priority:    1,
        RelatedID:   taskId,
        RelatedType: "task",
    }
    if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
        logx.Errorf("发布通知事件失败: %v", err)
    }
}
```

## EmailService 可用方法

- `SendTaskDispatchEmail` - 任务派发邮件
- `SendTaskDeadlineReminderEmail` - 任务截止提醒
- `SendTaskCompletedEmail` - 任务完成邮件
- `SendHandoverEmail` - 交接邮件

## 注意事项

1. **所有操作都是异步的** - 发布到消息队列后立即返回，不等待处理完成
2. **错误处理** - 发布失败只记录日志，不中断主流程
3. **模板数据** - 确保传入的数据格式正确（参考 `email_template.go` 中的数据结构）
4. **消息队列未配置** - 如果 RabbitMQ 未配置，服务会记录警告但不会崩溃

## 完成迁移后

1. 删除 `task/internal/svc/notification.go`
2. 运行测试确保所有功能正常
3. 检查日志确认消息队列正常工作



## 概述

本次重构将：
1. **删除** `task/internal/svc/notification.go` 文件
2. **使用** 邮件模板系统（`.tpl` 文件）生成邮件内容
3. **统一** 通过消息队列（RabbitMQ）异步处理所有邮件和通知

## 架构变化

### 之前
```
业务逻辑 → NotificationService → EmailMiddleware.SendEmail (同步)
业务逻辑 → NotificationService → NotificationModel.Insert (同步)
```

### 现在
```
业务逻辑 → EmailService (模板渲染) → EmailMQService → RabbitMQ → 消费者 → EmailMiddleware.SendEmail (异步)
业务逻辑 → NotificationMQService → RabbitMQ → 消费者 → NotificationModel.Insert (异步)
```

## 需要修改的文件

### 1. Logic 层文件（需要修改）

#### task/internal/logic/tasknode/updateTaskNodeLogic.go ✅ 已修改
- 替换 `NotificationService.SendHandoverNotification` 
- 使用 `EmailService.SendHandoverEmail` + `NotificationMQService.PublishNotificationEvent`

#### task/internal/logic/task/updateTaskProgressLogic.go
```go
// 旧代码
notificationService := svc.NewNotificationService(l.svcCtx)
notificationService.SendTaskCompletionNotification(...)

// 新代码
// 发送邮件
if l.svcCtx.EmailService != nil {
    l.svcCtx.EmailService.SendTaskCompletedEmail(...)
}
// 创建通知
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{...}
    l.svcCtx.NotificationMQService.PublishNotificationEvent(...)
}
```

#### task/internal/logic/task/createTaskLogic.go
- 替换所有 `NotificationService` 调用
- 使用 `EmailService` + `NotificationMQService`

#### task/internal/logic/employee/employeeLeaveLogic.go
- 替换 `SendEmployeeLeaveNotification`
- 使用模板渲染 + 消息队列

#### task/internal/logic/handover/*.go (3个文件)
- 替换 `SendHandoverNotification`
- 使用 `EmailService.SendHandoverEmail` + `NotificationMQService`

#### task/internal/logic/task/completeTaskLogic.go
- 替换 `SendHandoverNotification`
- 使用消息队列

#### task/internal/logic/task/deleteTaskLogic.go
- 替换 `SendTaskDeletedNotification`
- 使用消息队列

#### task/internal/logic/tasknode/deleteTaskNodeLogic.go
- 替换 `SendTaskNodeDeletedNotification`
- 使用消息队列

### 2. Scheduler 文件 ✅ 已部分修改

#### task/internal/svc/scheduler.go
- ✅ `SendTaskDeadlineReminder` 已改为消息队列
- 需要修改其他定时任务中的通知发送

### 3. 删除文件

- `task/internal/svc/notification.go` - 删除整个文件

## 迁移模式

### 模式1：发送邮件 + 创建通知

```go
// 1. 发送邮件（通过消息队列）
if l.svcCtx.EmailService != nil {
    err := l.svcCtx.EmailService.SendHandoverEmail(
        ctx,
        employeeEmail,
        employeeName,
        message,
        handoverID,
        taskTitle,
    )
    if err != nil {
        logx.Errorf("发送邮件失败: %v", err)
    }
}

// 2. 创建系统通知（通过消息队列）
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{
        EventType:   "handover.notification",
        EmployeeIDs: []string{employeeID},
        Title:       "任务交接通知",
        Content:     message,
        Type:        3,
        Category:    "handover",
        Priority:    2,
        RelatedID:   handoverID,
        RelatedType: "handover",
    }
    if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
        logx.Errorf("发布通知事件失败: %v", err)
    }
}
```

### 模式2：只发送邮件

```go
if l.svcCtx.EmailService != nil {
    err := l.svcCtx.EmailService.SendTaskDispatchEmail(
        ctx,
        employeeEmail,
        employeeName,
        taskTitle,
        nodeName,
        nodeDetail,
        deadline,
        taskId,
    )
    if err != nil {
        logx.Errorf("发送邮件失败: %v", err)
    }
}
```

### 模式3：只创建通知

```go
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{
        EventType:   "task.detail.updated",
        EmployeeIDs: employeeIDs,
        Title:       "任务详情更新",
        Content:     "任务详情已更新",
        Type:        1,
        Category:    "task",
        Priority:    1,
        RelatedID:   taskId,
        RelatedType: "task",
    }
    if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
        logx.Errorf("发布通知事件失败: %v", err)
    }
}
```

## EmailService 可用方法

- `SendTaskDispatchEmail` - 任务派发邮件
- `SendTaskDeadlineReminderEmail` - 任务截止提醒
- `SendTaskCompletedEmail` - 任务完成邮件
- `SendHandoverEmail` - 交接邮件

## 注意事项

1. **所有操作都是异步的** - 发布到消息队列后立即返回，不等待处理完成
2. **错误处理** - 发布失败只记录日志，不中断主流程
3. **模板数据** - 确保传入的数据格式正确（参考 `email_template.go` 中的数据结构）
4. **消息队列未配置** - 如果 RabbitMQ 未配置，服务会记录警告但不会崩溃

## 完成迁移后

1. 删除 `task/internal/svc/notification.go`
2. 运行测试确保所有功能正常
3. 检查日志确认消息队列正常工作



## 概述

本次重构将：
1. **删除** `task/internal/svc/notification.go` 文件
2. **使用** 邮件模板系统（`.tpl` 文件）生成邮件内容
3. **统一** 通过消息队列（RabbitMQ）异步处理所有邮件和通知

## 架构变化

### 之前
```
业务逻辑 → NotificationService → EmailMiddleware.SendEmail (同步)
业务逻辑 → NotificationService → NotificationModel.Insert (同步)
```

### 现在
```
业务逻辑 → EmailService (模板渲染) → EmailMQService → RabbitMQ → 消费者 → EmailMiddleware.SendEmail (异步)
业务逻辑 → NotificationMQService → RabbitMQ → 消费者 → NotificationModel.Insert (异步)
```

## 需要修改的文件

### 1. Logic 层文件（需要修改）

#### task/internal/logic/tasknode/updateTaskNodeLogic.go ✅ 已修改
- 替换 `NotificationService.SendHandoverNotification` 
- 使用 `EmailService.SendHandoverEmail` + `NotificationMQService.PublishNotificationEvent`

#### task/internal/logic/task/updateTaskProgressLogic.go
```go
// 旧代码
notificationService := svc.NewNotificationService(l.svcCtx)
notificationService.SendTaskCompletionNotification(...)

// 新代码
// 发送邮件
if l.svcCtx.EmailService != nil {
    l.svcCtx.EmailService.SendTaskCompletedEmail(...)
}
// 创建通知
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{...}
    l.svcCtx.NotificationMQService.PublishNotificationEvent(...)
}
```

#### task/internal/logic/task/createTaskLogic.go
- 替换所有 `NotificationService` 调用
- 使用 `EmailService` + `NotificationMQService`

#### task/internal/logic/employee/employeeLeaveLogic.go
- 替换 `SendEmployeeLeaveNotification`
- 使用模板渲染 + 消息队列

#### task/internal/logic/handover/*.go (3个文件)
- 替换 `SendHandoverNotification`
- 使用 `EmailService.SendHandoverEmail` + `NotificationMQService`

#### task/internal/logic/task/completeTaskLogic.go
- 替换 `SendHandoverNotification`
- 使用消息队列

#### task/internal/logic/task/deleteTaskLogic.go
- 替换 `SendTaskDeletedNotification`
- 使用消息队列

#### task/internal/logic/tasknode/deleteTaskNodeLogic.go
- 替换 `SendTaskNodeDeletedNotification`
- 使用消息队列

### 2. Scheduler 文件 ✅ 已部分修改

#### task/internal/svc/scheduler.go
- ✅ `SendTaskDeadlineReminder` 已改为消息队列
- 需要修改其他定时任务中的通知发送

### 3. 删除文件

- `task/internal/svc/notification.go` - 删除整个文件

## 迁移模式

### 模式1：发送邮件 + 创建通知

```go
// 1. 发送邮件（通过消息队列）
if l.svcCtx.EmailService != nil {
    err := l.svcCtx.EmailService.SendHandoverEmail(
        ctx,
        employeeEmail,
        employeeName,
        message,
        handoverID,
        taskTitle,
    )
    if err != nil {
        logx.Errorf("发送邮件失败: %v", err)
    }
}

// 2. 创建系统通知（通过消息队列）
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{
        EventType:   "handover.notification",
        EmployeeIDs: []string{employeeID},
        Title:       "任务交接通知",
        Content:     message,
        Type:        3,
        Category:    "handover",
        Priority:    2,
        RelatedID:   handoverID,
        RelatedType: "handover",
    }
    if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
        logx.Errorf("发布通知事件失败: %v", err)
    }
}
```

### 模式2：只发送邮件

```go
if l.svcCtx.EmailService != nil {
    err := l.svcCtx.EmailService.SendTaskDispatchEmail(
        ctx,
        employeeEmail,
        employeeName,
        taskTitle,
        nodeName,
        nodeDetail,
        deadline,
        taskId,
    )
    if err != nil {
        logx.Errorf("发送邮件失败: %v", err)
    }
}
```

### 模式3：只创建通知

```go
if l.svcCtx.NotificationMQService != nil {
    event := &svc.NotificationEvent{
        EventType:   "task.detail.updated",
        EmployeeIDs: employeeIDs,
        Title:       "任务详情更新",
        Content:     "任务详情已更新",
        Type:        1,
        Category:    "task",
        Priority:    1,
        RelatedID:   taskId,
        RelatedType: "task",
    }
    if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
        logx.Errorf("发布通知事件失败: %v", err)
    }
}
```

## EmailService 可用方法

- `SendTaskDispatchEmail` - 任务派发邮件
- `SendTaskDeadlineReminderEmail` - 任务截止提醒
- `SendTaskCompletedEmail` - 任务完成邮件
- `SendHandoverEmail` - 交接邮件

## 注意事项

1. **所有操作都是异步的** - 发布到消息队列后立即返回，不等待处理完成
2. **错误处理** - 发布失败只记录日志，不中断主流程
3. **模板数据** - 确保传入的数据格式正确（参考 `email_template.go` 中的数据结构）
4. **消息队列未配置** - 如果 RabbitMQ 未配置，服务会记录警告但不会崩溃

## 完成迁移后

1. 删除 `task/internal/svc/notification.go`
2. 运行测试确保所有功能正常
3. 检查日志确认消息队列正常工作


