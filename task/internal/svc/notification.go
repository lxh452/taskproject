package svc

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"task_Project/model/task"
	"task_Project/model/user_auth"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

// NotificationService 通知服务
type NotificationService struct {
	svcCtx *ServiceContext
}

// NewNotificationService 创建通知服务
func NewNotificationService(svcCtx *ServiceContext) *NotificationService {
	return &NotificationService{
		svcCtx: svcCtx,
	}
}

// todo 全部重新修改
// SendTaskDispatchNotification 发送任务派发通知
func (n *NotificationService) SendTaskDispatchNotification(ctx context.Context, taskNode *task.TaskNode, employeeID string) error {
	// 获取员工信息
	employee, err := n.svcCtx.EmployeeModel.FindOne(ctx, employeeID)
	if err != nil {
		logx.Errorf("查询员工失败: %v", err)
		return err
	}

	// 获取任务信息
	taskInfo, err := n.svcCtx.TaskModel.FindOne(ctx, taskNode.TaskId)
	if err != nil {
		return err
	}

	// 构建邮件内容
	subject := "任务派发通知"
	body := fmt.Sprintf(`
		<h2>任务派发通知</h2>
		<p>您好 %s，</p>
		<p>您已被分配执行以下任务节点：</p>
		<ul>
			<li><strong>任务名称：</strong>%s</li>
			<li><strong>节点名称：</strong>%s</li>
			<li><strong>节点详情：</strong>%s</li>
			<li><strong>截止时间：</strong>%s</li>
		</ul>
		<p>请及时查看并开始执行任务。</p>
		<p>祝工作顺利！</p>
	`, employee.RealName, taskInfo.TaskTitle, taskNode.NodeName,
		getStringValue(taskNode.NodeDetail), taskNode.NodeDeadline.Format("2006-01-02 15:04:05"))

	// 发送邮件通知
	err = n.SendEmailNotification(employee.Email.String, subject, body)
	if err != nil {
		logx.Errorf("发送邮件通知失败: %v", err)
	}

	// 发送系统通知
	notification := &user_auth.Notification{
		Id:          utils.NewCommon().GenerateID(),
		EmployeeId:  employeeID,
		Title:       subject,
		Content:     fmt.Sprintf("您有一个新的任务需要处理：%s", taskNode.NodeName),
		Type:        1, // 任务通知
		Priority:    2, // 中等优先级
		IsRead:      0, // 未读
		RelatedId:   sql.NullString{String: taskNode.TaskId, Valid: true},
		RelatedType: sql.NullString{String: "task", Valid: true},
		SenderId:    sql.NullString{String: "system", Valid: true},
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	_, err = n.svcCtx.NotificationModel.Insert(ctx, notification)
	if err != nil {
		logx.Errorf("创建系统通知失败: %v", err)
	}

	return nil
}

// SendEmployeeLeaveNotification 发送员工离职通知
func (n *NotificationService) SendEmployeeLeaveNotification(ctx context.Context, employeeID string) error {
	// 获取员工信息
	employee, err := n.svcCtx.EmployeeModel.FindOne(ctx, employeeID)
	if err != nil {
		logx.Errorf("查询员工失败: %v", err)
		return err
	}

	// 获取员工当前负责的任务节点
	taskNodes, err := n.getEmployeeCurrentTaskNodes(ctx, employeeID)
	if err != nil {
		logx.Errorf("查询员工任务节点失败: %v", err)
		return err
	}

	if len(taskNodes) == 0 {
		return nil // 没有需要交接的任务
	}

	// 发送通知给部门负责人
	department, err := n.svcCtx.DepartmentModel.FindOne(ctx, employee.DepartmentId.String)
	if err != nil {
		logx.Errorf("查询部门失败: %v", err)
		return err
	}

	if department.ManagerId.Valid && department.ManagerId.String != "" {
		subject := "员工离职任务交接通知"
		body := fmt.Sprintf(`
			<h2>员工离职任务交接通知</h2>
			<p>您好，</p>
			<p>员工 %s 已离职，其负责的以下任务需要重新分配：</p>
			<ul>
		`, employee.RealName)

		for _, taskNode := range taskNodes {
			body += fmt.Sprintf("<li>%s</li>", taskNode.NodeName)
		}

		body += `
			</ul>
			<p>请及时登录系统进行任务重新分配。</p>
		`

		err = n.SendEmailNotification(department.ManagerId.String, subject, body)
		if err != nil {
			logx.Errorf("发送邮件通知失败: %v", err)
		}
	}

	return nil
}

// SendTaskDeadlineReminder 发送任务截止时间提醒
func (n *NotificationService) SendTaskDeadlineReminder(ctx context.Context, taskNode *task.TaskNode) error {
	// 获取执行人信息
	executor, err := n.svcCtx.EmployeeModel.FindOne(ctx, taskNode.ExecutorId)
	if err != nil {
		logx.Errorf("查询执行人失败: %v", err)
		return err
	}

	// 获取负责人信息
	_, err = n.svcCtx.EmployeeModel.FindOne(ctx, taskNode.LeaderId)
	if err != nil {
		logx.Errorf("查询负责人失败: %v", err)
		return err
	}

	// 发送邮件通知
	emailBody := fmt.Sprintf(`
		<html>
		<body>
			<h2>任务截止时间提醒</h2>
			<p>您好，</p>
			<p>您负责的任务即将到期：</p>
			<ul>
				<li>任务节点：%s</li>
				<li>截止时间：%s</li>
				<li>当前进度：%d%%</li>
			</ul>
			<p>请及时完成或调整计划。</p>
		</body>
		</html>
	`, taskNode.NodeName, taskNode.NodeDeadline.Format("2006-01-02 15:04:05"), taskNode.Progress)

	err = n.SendEmailNotification(executor.Email.String, "任务截止时间提醒", emailBody)
	if err != nil {
		logx.Errorf("发送邮件通知失败: %v", err)
	}

	notification := &user_auth.Notification{
		Id:          utils.NewCommon().GenerateID(),
		EmployeeId:  taskNode.ExecutorId,
		Title:       "任务截止时间提醒",
		Content:     fmt.Sprintf("任务节点 %s 即将到期", taskNode.NodeName),
		Type:        2, // 截止提醒
		Priority:    3, // 高优先级
		IsRead:      0, // 未读
		RelatedId:   sql.NullString{String: taskNode.TaskId, Valid: true},
		RelatedType: sql.NullString{String: "task", Valid: true},
		SenderId:    sql.NullString{String: "system", Valid: true},
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	_, err = n.svcCtx.NotificationModel.Insert(ctx, notification)
	if err != nil {
		logx.Errorf("创建系统通知失败: %v", err)
	}

	return nil
}

// SendDailyReportReminder 发送每日工作报告提醒
func (n *NotificationService) SendDailyReportReminder(ctx context.Context, employeeID string) error {
	// 获取员工信息
	employee, err := n.svcCtx.EmployeeModel.FindOne(ctx, employeeID)
	if err != nil {
		logx.Errorf("查询员工失败: %v", err)
		return err
	}

	// 发送邮件通知
	emailBody := fmt.Sprintf(`
		<html>
		<body>
			<h2>每日工作报告提醒</h2>
			<p>您好 %s，</p>
			<p>请及时提交今日的工作报告。</p>
			<p>祝工作顺利！</p>
		</body>
		</html>
	`, employee.RealName)

	err = n.SendEmailNotification(employee.Email.String, "每日工作报告提醒", emailBody)
	if err != nil {
		logx.Errorf("发送邮件通知失败: %v", err)
	}

	return nil
}

// SendSlowProgressNotification 发送进度缓慢通知
func (n *NotificationService) SendSlowProgressNotification(ctx context.Context, taskNode *task.TaskNode) error {
	// 获取负责人信息
	leader, err := n.svcCtx.EmployeeModel.FindOne(ctx, taskNode.LeaderId)
	if err != nil {
		logx.Errorf("查询负责人失败: %v", err)
		return err
	}

	// 发送邮件通知
	emailBody := fmt.Sprintf(`
		<html>
		<body>
			<h2>任务进度缓慢提醒</h2>
			<p>您好，</p>
			<p>以下任务进度可能较慢：</p>
			<ul>
				<li>任务节点：%s</li>
				<li>当前进度：%d%%</li>
				<li>预计完成时间：%s</li>
			</ul>
			<p>请确认是否需要增加人员或调整计划。</p>
		</body>
		</html>
	`, taskNode.NodeName, taskNode.Progress, taskNode.NodeDeadline.Format("2006-01-02 15:04:05"))

	err = n.SendEmailNotification(leader.Email.String, "任务进度缓慢提醒", emailBody)
	if err != nil {
		logx.Errorf("发送邮件通知失败: %v", err)
	}

	return nil
}

// SendTaskCompletionNotification 发送任务完成通知
func (n *NotificationService) SendTaskCompletionNotification(ctx context.Context, taskNode *task.TaskNode) error {
	// 获取任务信息
	task, err := n.svcCtx.TaskModel.FindOne(ctx, taskNode.TaskId)
	if err != nil {
		logx.Errorf("查询任务失败: %v", err)
		return err
	}

	// 获取节点负责人信息
	leader, err := n.svcCtx.EmployeeModel.FindOne(ctx, taskNode.LeaderId)
	if err != nil {
		logx.Errorf("查询节点负责人失败: %v", err)
		return err
	}

	// 发送邮件通知
	emailBody := fmt.Sprintf(`
		<html>
		<body>
			<h2>任务完成通知</h2>
			<p>您好，</p>
			<p>任务节点已完成：</p>
			<ul>
				<li>任务标题：%s</li>
				<li>节点名称：%s</li>
				<li>完成时间：%s</li>
			</ul>
			<p>请及时查看任务进度。</p>
		</body>
		</html>
	`, task.TaskTitle, taskNode.NodeName, time.Now().Format("2006-01-02 15:04:05"))

	err = n.SendEmailNotification(leader.Email.String, "任务完成通知", emailBody)
	if err != nil {
		logx.Errorf("发送邮件通知失败: %v", err)
	}

	// 发送系统通知
	notification := &user_auth.Notification{
		Id:          utils.NewCommon().GenerateID(),
		EmployeeId:  taskNode.LeaderId,
		Title:       "任务完成通知",
		Content:     fmt.Sprintf("任务节点 %s 已完成", taskNode.NodeName),
		Type:        3, // 任务完成
		Priority:    2, // 中等优先级
		IsRead:      0, // 未读
		RelatedId:   sql.NullString{String: taskNode.TaskId, Valid: true},
		RelatedType: sql.NullString{String: "task", Valid: true},
		SenderId:    sql.NullString{String: "system", Valid: true},
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	_, err = n.svcCtx.NotificationModel.Insert(ctx, notification)
	if err != nil {
		logx.Errorf("创建系统通知失败: %v", err)
	}

	return nil
}

// SendCrossDepartmentTaskNotification 发送跨部门任务通知
func (n *NotificationService) SendCrossDepartmentTaskNotification(managerID, taskID, taskTitle, departmentName string) error {
	// 获取部门负责人信息
	manager, err := n.svcCtx.EmployeeModel.FindOne(context.Background(), managerID)
	if err != nil {
		logx.Errorf("查询部门负责人失败: %v", err)
		return err
	}

	// 发送邮件通知
	emailBody := fmt.Sprintf(`
		<html>
		<body>
			<h2>跨部门任务通知</h2>
			<p>您好，</p>
			<p>您有一个新的跨部门任务需要处理：</p>
			<ul>
				<li>任务标题：%s</li>
				<li>任务ID：%s</li>
				<li>负责部门：%s</li>
				<li>创建时间：%s</li>
			</ul>
			<p>请及时登录系统创建任务节点并分配执行人员。</p>
			<p>任务链接：<a href="/task/detail/%s">查看任务详情</a></p>
		</body>
		</html>
	`, taskTitle, taskID, departmentName, time.Now().Format("2006-01-02 15:04:05"), taskID)

	err = n.SendEmailNotification(manager.Email.String, "跨部门任务通知", emailBody)
	if err != nil {
		logx.Errorf("发送邮件通知失败: %v", err)
	}

	// 发送系统通知
	notification := &user_auth.Notification{
		Id:          utils.NewCommon().GenerateID(),
		EmployeeId:  managerID,
		Title:       "跨部门任务通知",
		Content:     fmt.Sprintf("您有一个新的跨部门任务需要处理：%s", taskTitle),
		Type:        1, // 任务通知
		Priority:    3, // 高优先级
		IsRead:      0, // 未读
		RelatedId:   sql.NullString{String: taskID, Valid: true},
		RelatedType: sql.NullString{String: "task", Valid: true},
		SenderId:    sql.NullString{String: "system", Valid: true},
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	_, err = n.svcCtx.NotificationModel.Insert(context.Background(), notification)
	if err != nil {
		logx.Errorf("创建系统通知失败: %v", err)
	}

	return nil
}

// SendTaskNodeCreatedNotification 发送任务节点创建通知
func (n *NotificationService) SendTaskNodeCreatedNotification(leaderID, taskID, nodeID, nodeName string) error {
	// 获取负责人信息
	leader, err := n.svcCtx.EmployeeModel.FindOne(context.Background(), leaderID)
	if err != nil {
		logx.Errorf("查询负责人失败: %v", err)
		return err
	}

	// 发送邮件通知
	emailBody := fmt.Sprintf(`
		<html>
		<body>
			<h2>任务节点创建通知</h2>
			<p>您好，</p>
			<p>您有一个新的任务节点需要处理：</p>
			<ul>
				<li>节点名称：%s</li>
				<li>节点ID：%s</li>
				<li>任务ID：%s</li>
				<li>创建时间：%s</li>
			</ul>
			<p>请及时登录系统查看任务详情并开始执行。</p>
			<p>任务链接：<a href="/task/detail/%s">查看任务详情</a></p>
		</body>
		</html>
	`, nodeName, nodeID, taskID, time.Now().Format("2006-01-02 15:04:05"), taskID)

	err = n.SendEmailNotification(leader.Email.String, "任务节点创建通知", emailBody)
	if err != nil {
		logx.Errorf("发送邮件通知失败: %v", err)
	}

	// 发送系统通知
	notification := &user_auth.Notification{
		Id:          utils.NewCommon().GenerateID(),
		EmployeeId:  leaderID,
		Title:       "任务节点创建通知",
		Content:     fmt.Sprintf("您有一个新的任务节点需要处理：%s", nodeName),
		Type:        1, // 任务通知
		Priority:    2, // 中等优先级
		IsRead:      0, // 未读
		RelatedId:   sql.NullString{String: taskID, Valid: true},
		RelatedType: sql.NullString{String: "task", Valid: true},
		SenderId:    sql.NullString{String: "system", Valid: true},
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	_, err = n.svcCtx.NotificationModel.Insert(context.Background(), notification)
	if err != nil {
		logx.Errorf("创建系统通知失败: %v", err)
	}

	return nil
}

// SendEmailNotification 发送邮件通知
func (n *NotificationService) SendEmailNotification(to, subject, body string) error {
	// TODO: 实现邮件发送逻辑
	// 这里可以集成SMTP服务或其他邮件服务
	logx.Infof("发送邮件通知: to=%s, subject=%s", to, subject)
	return nil
}

// SendSMSNotification 发送短信通知
func (n *NotificationService) SendSMSNotification(phone, content string) error {
	// TODO: 实现短信发送逻辑
	// 这里可以集成短信服务
	logx.Infof("发送短信通知: phone=%s, content=%s", phone, content)
	return nil
}

// SendSystemNotification 发送系统通知
func (n *NotificationService) SendSystemNotification(employeeID, title, content string) error {
	notification := &user_auth.Notification{
		Id:          utils.NewCommon().GenerateID(),
		EmployeeId:  employeeID,
		Title:       title,
		Content:     content,
		Type:        0, // 系统通知
		Priority:    1, // 低优先级
		IsRead:      0, // 未读
		RelatedId:   sql.NullString{String: "", Valid: false},
		RelatedType: sql.NullString{String: "", Valid: false},
		SenderId:    sql.NullString{String: "system", Valid: true},
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	_, err := n.svcCtx.NotificationModel.Insert(context.Background(), notification)
	return err
}

// 获取员工当前负责的任务节点
func (n *NotificationService) getEmployeeCurrentTaskNodes(ctx context.Context, employeeID string) ([]*task.TaskNode, error) {
	// 查询员工当前负责的任务节点
	// 假设有FindByExecutorAndStatus方法
	return []*task.TaskNode{}, nil // TODO: 实现具体查询逻辑
}

// getStringValue 获取字符串值
func getStringValue(s sql.NullString) string {
	if !s.Valid {
		return ""
	}
	return s.String
}
