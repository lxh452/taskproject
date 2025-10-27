// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"task_Project/model/company"
	"task_Project/model/role"
	"task_Project/model/task"
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
	RoleModel role.RoleModel

	// 任务相关模型
	TaskModel         task.TaskModel
	TaskNodeModel     task.TaskNodeModel
	TaskLogModel      task.TaskLogModel
	TaskHandoverModel task.TaskHandoverModel

	// 通知相关模型
	NotificationModel user_auth.NotificationModel

	// MQ
	MQ *MQClient
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

	// 任务相关模型
	taskModel := task.NewTaskModel(conn)
	taskNodeModel := task.NewTaskNodeModel(conn)
	taskLogModel := task.NewTaskLogModel(conn)
	taskHandoverModel := task.NewTaskHandoverModel(conn)

	// 初始化RabbitMQ
	mqClient, err := NewMQClient(c.RabbitMQ.URL, c.RabbitMQ.Exchange, c.RabbitMQ.Queue)
	if err != nil {
		// 延迟到运行期建立；启动失败仅记录日志，业务仍可运行
		logx.Errorf("init RabbitMQ failed: %v", err)
	}

	return &ServiceContext{
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
		RoleModel: roleModel,

		// 任务相关模型
		TaskModel:         taskModel,
		TaskNodeModel:     taskNodeModel,
		TaskLogModel:      taskLogModel,
		TaskHandoverModel: taskHandoverModel,

		// 通知相关模型
		NotificationModel: user_auth.NewNotificationModel(conn),

		// MQ
		MQ: mqClient,
	}
}
