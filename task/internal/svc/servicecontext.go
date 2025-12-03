// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"fmt"
	"task_Project/model/company"
	"task_Project/model/role"
	"task_Project/model/task"
	"task_Project/model/upload"
	"task_Project/model/user"
	"task_Project/model/user_auth"
	"task_Project/task/internal/config"
	"task_Project/task/internal/middleware"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

	// 中间件
	JWTMiddleware   *middleware.JWTMiddleware
	EmailMiddleware *middleware.EmailMiddleware
	SMSMiddleware   *middleware.SMSMiddleware

	// 事务服务
	TransactionService *TransactionService
	TransactionHelper  *TransactionHelper

	// 用户相关模型
	UserModel     user.UserModel
	EmployeeModel user.EmployeeModel

	// 公司相关模型
	CompanyModel    company.CompanyModel
	DepartmentModel company.DepartmentModel
	PositionModel   company.PositionModel

	// 角色相关模型
	RoleModel         role.RoleModel
	PositionRoleModel role.PositionRoleModel

	// 任务相关模型
	TaskModel             task.TaskModel
	TaskNodeModel         task.TaskNodeModel
	TaskLogModel          task.TaskLogModel
	TaskHandoverModel     task.TaskHandoverModel
	HandoverApprovalModel task.HandoverApprovalModel
	TaskChecklistModel    task.TaskChecklistModel

	// 通知相关模型
	NotificationModel user_auth.NotificationModel

	// 权限相关模型
	UserPermissionModel user_auth.UserPermissionModel

	// MongoDB 相关模型
	MongoURL               string                        // MongoDB 连接 URL
	MongoDB                string                        // MongoDB 数据库名
	UploadFileModel        upload.Upload_fileModel       // 文件上传模型
	TaskProjectDetailModel task.Task_project_detailModel // 任务详情模型

	// RabbitMQ 相关
	MQClient              *MQClient              // RabbitMQ 客户端
	NotificationMQService *NotificationMQService // 通知消息队列服务
	EmailMQService        *EmailMQService        // 邮件消息队列服务

	// 邮件模板和服务
	EmailTemplateService *EmailTemplateService // 邮件模板服务
	EmailService         *EmailService         // 邮件服务

	Scheduler *SchedulerService
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化JWT中间件
	jwtMiddleware := middleware.NewJWTMiddleware(middleware.JWTConfig{
		SecretKey:   c.JWT.SecretKey,
		ExpireTime:  c.JWT.ExpireTime,
		RefreshTime: c.JWT.RefreshTime,
		Issuer:      c.JWT.Issuer,
		Audience:    c.JWT.Audience,
	})

	// 初始化邮件中间件
	emailMiddleware := middleware.NewEmailMiddleware(middleware.EmailConfig{
		Host:     c.Email.Host,
		Port:     c.Email.Port,
		Username: c.Email.Username,
		Password: c.Email.Password,
		From:     c.Email.From,
		UseTLS:   c.Email.UseTLS,
	})

	// 初始化短信中间件
	smsMiddleware := middleware.NewSMSMiddleware(middleware.SMSConfig{
		Provider:   c.SMS.Provider,
		AccessKey:  c.SMS.AccessKey,
		SecretKey:  c.SMS.SecretKey,
		SignName:   c.SMS.SignName,
		TemplateID: c.SMS.TemplateID,
		Endpoint:   c.SMS.Endpoint,
		Region:     c.SMS.Region,
	})
	// 构建 MongoDB 连接 URL
	// 格式: mongodb://[username:password@]host[:port]/[database][?authSource=admin]
	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=%s",
		c.Mongo.Username,
		c.Mongo.Password,
		c.Mongo.Host,
		c.Mongo.Port,
		c.Mongo.Database,
		c.Mongo.AuthSource,
	)

	// 初始化 MongoDB model
	uploadFileModel := upload.NewUpload_fileModel(mongoURL, c.Mongo.Database, "file_upload")
	taskProjectDetailModel := task.NewTask_project_detailModel(mongoURL, c.Mongo.Database, "task_project_detail")
	// 初始化数据库连接
	conn := sqlx.NewMysql(c.MySQL.DataSource)

	// 初始化事务服务
	transactionService := NewTransactionService(conn)
	transactionHelper := NewTransactionHelper(conn)

	// 初始化所有模型
	userModel := user.NewUserModel(conn)
	employeeModel := user.NewEmployeeModel(conn)
	companyModel := company.NewCompanyModel(conn)
	departmentModel := company.NewDepartmentModel(conn)
	positionModel := company.NewPositionModel(conn)
	roleModel := role.NewRoleModel(conn)
	positionRoleModel := role.NewPositionRoleModel(conn)

	// 任务相关模型
	taskModel := task.NewTaskModel(conn)
	taskNodeModel := task.NewTaskNodeModel(conn)
	taskLogModel := task.NewTaskLogModel(conn)
	taskHandoverModel := task.NewTaskHandoverModel(conn)
	handoverApprovalModel := task.NewHandoverApprovalModel(conn)
	taskChecklistModel := task.NewTaskChecklistModel(conn)

	// 初始化 RabbitMQ
	var mqClient *MQClient
	var notificationMQService *NotificationMQService
	var emailMQService *EmailMQService
	if c.RabbitMQ.URL != "" {
		logx.Infof("[ServiceContext] Initializing RabbitMQ: url=%s, exchange=%s", c.RabbitMQ.URL, c.RabbitMQ.Exchange)
		mq, err := NewMQClient(c.RabbitMQ.URL, c.RabbitMQ.Exchange)
		if err != nil {
			logx.Errorf("[ServiceContext] Failed to initialize RabbitMQ: error=%v, url=%s, exchange=%s, notifications and emails will be disabled",
				err, c.RabbitMQ.URL, c.RabbitMQ.Exchange)
		} else {
			logx.Infof("[ServiceContext] RabbitMQ client initialized successfully: exchange=%s", c.RabbitMQ.Exchange)
			mqClient = mq
			notificationMQService = NewNotificationMQService(mq)
			emailMQService = NewEmailMQService(mq, emailMiddleware)
			logx.Infof("[ServiceContext] EmailMQService and NotificationMQService created successfully")

			// 注意：消费者需要 ServiceContext，将在 ServiceContext 完全初始化后启动
		}
	} else {
		logx.Infof("[ServiceContext] RabbitMQ URL is empty, message queue services will be disabled")
	}

	// 初始化邮件模板服务
	emailTemplateService, err := NewEmailTemplateService()
	if err != nil {
		logx.Errorf("Failed to initialize email template service: %v", err)
		emailTemplateService = nil
	}

	// 初始化邮件服务
	var emailService *EmailService
	if emailTemplateService != nil {
		emailService = NewEmailService(emailTemplateService, emailMQService, emailMiddleware, c.System.BaseURL)
	}

	s := &ServiceContext{
		Config:          c,
		JWTMiddleware:   jwtMiddleware,
		EmailMiddleware: emailMiddleware,
		SMSMiddleware:   smsMiddleware,

		// 事务服务
		TransactionService: transactionService,
		TransactionHelper:  transactionHelper,

		// 用户相关模型
		UserModel:     userModel,
		EmployeeModel: employeeModel,

		// 公司相关模型
		CompanyModel:    companyModel,
		DepartmentModel: departmentModel,
		PositionModel:   positionModel,

		// 角色相关模型
		RoleModel:         roleModel,
		PositionRoleModel: positionRoleModel,

		// 任务相关模型
		TaskModel:             taskModel,
		TaskNodeModel:         taskNodeModel,
		TaskLogModel:          taskLogModel,
		TaskHandoverModel:     taskHandoverModel,
		HandoverApprovalModel: handoverApprovalModel,
		TaskChecklistModel:    taskChecklistModel,

		// 通知相关模型
		NotificationModel: user_auth.NewNotificationModel(conn),

		// 权限相关模型
		UserPermissionModel: user_auth.NewUserPermissionModel(conn),

		// MongoDB 相关
		MongoURL:               mongoURL,
		MongoDB:                c.Mongo.Database,
		UploadFileModel:        uploadFileModel,
		TaskProjectDetailModel: taskProjectDetailModel,

		// RabbitMQ 相关
		MQClient:              mqClient,
		NotificationMQService: notificationMQService,
		EmailMQService:        emailMQService,

		// 邮件模板和服务
		EmailTemplateService: emailTemplateService,
		EmailService:         emailService,
	}
	s.Scheduler = NewSchedulerService(s)

	// 启动消息队列消费者（在 ServiceContext 完全初始化后）
	if mqClient != nil {
		logx.Infof("[ServiceContext] Starting message queue consumers...")

		// 启动邮件消费者
		if err := StartEmailConsumer(mqClient, "email_queue", s); err != nil {
			logx.Errorf("[ServiceContext] Failed to start email consumer: error=%v", err)
		} else {
			logx.Infof("[ServiceContext] Email consumer started successfully")
		}

		// 启动通知消费者
		if err := StartNotificationConsumer(mqClient, "notification_queue", s); err != nil {
			logx.Errorf("[ServiceContext] Failed to start notification consumer: error=%v", err)
		} else {
			logx.Infof("[ServiceContext] Notification consumer started successfully")
		}
	} else {
		logx.Infof("[ServiceContext] MQClient is nil, consumers will not be started")
	}

	return s
}
