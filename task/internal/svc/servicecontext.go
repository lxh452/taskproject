// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"fmt"
	adminModel "task_Project/model/admin"
	"task_Project/model/company"
	"task_Project/model/role"
	"task_Project/model/task"
	"task_Project/model/upload"
	"task_Project/model/user"
	"task_Project/model/user_auth"
	"task_Project/task/internal/config"
	"task_Project/task/internal/middleware"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config

	// 中间件
	JWTMiddleware       *middleware.JWTMiddleware
	AdminAuthMiddleware *middleware.AdminAuthMiddleware
	EmailMiddleware     *middleware.EmailMiddleware
	SMSMiddleware       *middleware.SMSMiddleware

	// Redis 客户端（用于Token存储和验证）
	RedisClient *redis.Redis

	// 事务服务
	TransactionService *TransactionService
	TransactionHelper  *TransactionHelper

	// 用户相关模型
	UserModel     user.UserModel
	EmployeeModel user.EmployeeModel

	// 管理员相关模型
	AdminModel       adminModel.AdminModel
	LoginRecordModel adminModel.LoginRecordModel
	SystemLogModel   adminModel.SystemLogModel

	// 系统日志服务
	SystemLogService *SystemLogService

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

	// 加入公司相关
	JoinApplicationModel user.JoinApplicationModel
	InviteCodeService    *InviteCodeService

	// MongoDB 相关模型
	MongoURL               string                         // MongoDB 连接 URL
	MongoDB                string                         // MongoDB 数据库名
	UploadFileModel        upload.Upload_fileModel        // 文件上传模型
	TaskProjectDetailModel task.Task_project_detailModel  // 任务详情模型
	TaskCommentModel       task.Task_commentModel         // 任务评论模型(MongoDB)
	AttachmentCommentModel upload.Attachment_commentModel // 附件评论标注模型(MongoDB)

	// RabbitMQ 相关
	MQClient              *MQClient              // RabbitMQ 客户端
	NotificationMQService *NotificationMQService // 通知消息队列服务
	EmailMQService        *EmailMQService        // 邮件消息队列服务

	// 邮件模板和服务
	EmailTemplateService *EmailTemplateService // 邮件模板服务
	EmailService         *EmailService         // 邮件服务

	// 文件存储服务（COS存储）
	FileStorageService FileStorageInterface

	// SQL执行服务
	SQLExecutorService *SQLExecutorService

	Scheduler *SchedulerService

	// GLM AI服务（智能任务派发）
	GLMService *GLMService
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
		Enabled:  c.Email.Enabled,
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

	// 初始化 Redis 客户端
	redisAddr := fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
	redisClient := redis.MustNewRedis(redis.RedisConf{
		Host: redisAddr,
		Pass: c.Redis.Password,
		Type: "node",
	})
	logx.Infof("[ServiceContext] Redis client initialized: addr=%s", redisAddr)

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
	taskCommentModel := task.NewTask_commentModel(mongoURL, c.Mongo.Database, "task_comment")
	attachmentCommentModel := upload.NewAttachment_commentModel(mongoURL, c.Mongo.Database, "attachment_comment")
	// 初始化数据库连接
	conn := sqlx.NewMysql(c.MySQL.DataSource)

	// 配置 MySQL 连接池（重要！防止连接泄漏和耗尽）
	// 使用 RawDB() 方法获取底层的 *sql.DB
	db, err := conn.RawDB()
	fmt.Println("mysql：", c.MySQL.DataSource)
	if err != nil {
		logx.Errorf("[ServiceContext] 获取底层数据库连接失败: %v", err)
	} else {
		db.SetMaxOpenConns(25)                  // 最大打开连接数
		db.SetMaxIdleConns(10)                  // 最大空闲连接数
		db.SetConnMaxLifetime(30 * time.Minute) // 连接最大生存时间（30分钟）
		db.SetConnMaxIdleTime(10 * time.Minute) // 空闲连接最大时间（10分钟）
		logx.Infof("[ServiceContext] MySQL 连接池配置: MaxOpenConns=25, MaxIdleConns=10, ConnMaxLifetime=30m, ConnMaxIdleTime=10m")

		// 启动连接池监控（每5分钟记录一次连接池状态）
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					stats := db.Stats()
					logx.Infof("[MySQL连接池监控] OpenConnections=%d, InUse=%d, Idle=%d, WaitCount=%d, WaitDuration=%v, MaxIdleClosed=%d, MaxLifetimeClosed=%d",
						stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount, stats.WaitDuration, stats.MaxIdleClosed, stats.MaxLifetimeClosed)
					// 如果等待连接数过多，发出警告
					if stats.WaitCount > 0 {
						logx.Errorf("[MySQL连接池警告] 有 %d 个请求在等待数据库连接，可能存在连接泄漏！", stats.WaitCount)
					}
				}
			}
		}()
	}

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

	// 管理员相关模型
	adminModelInstance := adminModel.NewAdminModel(conn)
	loginRecordModel := adminModel.NewLoginRecordModel(conn)
	systemLogModel := adminModel.NewSystemLogModel(mongoURL, c.Mongo.Database, "system_logs")

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

	// 初始化文件存储服务（仅支持COS存储）
	var fileStorageService FileStorageInterface

	// 使用COS存储
	secretId := c.FileStorage.COS.SecretId
	secretKey := c.FileStorage.COS.SecretKey
	bucket := c.FileStorage.COS.Bucket
	region := c.FileStorage.COS.Region
	urlPrefix := c.FileStorage.URLPrefix
	if urlPrefix == "" {
		urlPrefix = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region)
	}

	// 检查配置是否完整
	if secretId == "" || secretKey == "" || bucket == "" || region == "" {
		logx.Errorf("[ServiceContext] COS配置不完整: secretId=%v, secretKey=%v, bucket=%s, region=%s",
			secretId != "", secretKey != "", bucket, region)
		panic("COS配置不完整，无法启动服务。请配置环境变量 TENCENT_CLOUD_SECRET_ID 和 TENCENT_CLOUD_SECRET_KEY")
	}

	cosService, err := NewCOSStorageService(secretId, secretKey, bucket, region, urlPrefix)
	if err != nil {
		logx.Errorf("[ServiceContext] 初始化COS存储服务失败: %v", err)
		panic(fmt.Sprintf("初始化COS存储服务失败: %v", err))
	}
	fileStorageService = cosService
	logx.Infof("[ServiceContext] COS存储服务初始化成功: bucket=%s, region=%s, urlPrefix=%s", bucket, region, urlPrefix)

	// 初始化管理员认证中间件
	adminAuthMiddleware := middleware.NewAdminAuthMiddleware(jwtMiddleware, redisClient)

	s := &ServiceContext{
		Config:              c,
		JWTMiddleware:       jwtMiddleware,
		AdminAuthMiddleware: adminAuthMiddleware,
		EmailMiddleware:     emailMiddleware,
		SMSMiddleware:       smsMiddleware,

		// Redis 客户端
		RedisClient: redisClient,

		// 事务服务
		TransactionService: transactionService,
		TransactionHelper:  transactionHelper,

		// 用户相关模型
		UserModel:     userModel,
		EmployeeModel: employeeModel,

		// 管理员相关模型
		AdminModel:       adminModelInstance,
		LoginRecordModel: loginRecordModel,
		SystemLogModel:   systemLogModel,

		// 系统日志服务
		SystemLogService: NewSystemLogService(systemLogModel),

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

		// 加入公司相关
		JoinApplicationModel: user.NewJoinApplicationModel(conn),
		InviteCodeService:    NewInviteCodeService(redisClient),

		// MongoDB 相关
		MongoURL:               mongoURL,
		MongoDB:                c.Mongo.Database,
		UploadFileModel:        uploadFileModel,
		TaskProjectDetailModel: taskProjectDetailModel,
		TaskCommentModel:       taskCommentModel,
		AttachmentCommentModel: attachmentCommentModel,

		// RabbitMQ 相关
		MQClient:              mqClient,
		NotificationMQService: notificationMQService,
		EmailMQService:        emailMQService,

		// 邮件模板和服务
		EmailTemplateService: emailTemplateService,
		EmailService:         emailService,

		// 文件存储服务
		FileStorageService: fileStorageService,

		// SQL执行服务 (Docker 中路径为 ./model/sql，本地开发为 ./../model/sql)
		SQLExecutorService: NewSQLExecutorService(conn, "./model/sql"),
	}

	// 初始化GLM服务
	if c.GLM.APIKey != "" {
		s.GLMService = NewGLMService(GLMConfig{
			APIKey:  c.GLM.APIKey,
			BaseURL: c.GLM.BaseURL,
		})
		logx.Infof("[ServiceContext] GLM AI服务初始化成功")
	} else {
		logx.Infof("[ServiceContext] GLM API Key未配置，智能派发功能将不可用")
	}

	// 启动时自动执行数据库迁移
	logx.Info("[ServiceContext] 开始自动执行数据库迁移...")
	if err := s.SQLExecutorService.AutoMigrate(context.Background()); err != nil {
		logx.Errorf("[ServiceContext] 数据库迁移失败: %v", err)
	} else {
		logx.Info("[ServiceContext] 数据库迁移完成")
	}
	s.Scheduler = NewSchedulerService(s)

	// 设置Redis客户端给JWT中间件（用于Token验证）
	jwtMiddleware.SetRedisClient(redisClient)

	// 设置状态检查器给JWT中间件（用于实时检查用户/公司状态）
	statusChecker := NewStatusCheckerService(userModel, companyModel)
	jwtMiddleware.SetStatusChecker(statusChecker)

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
