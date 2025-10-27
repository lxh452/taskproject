package svc

import (
	"task_Project/model/company"
	"task_Project/model/role"
	"task_Project/model/task"
	"task_Project/model/user"
	"task_Project/model/user_auth"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// TransactionHelper 事务辅助工具
type TransactionHelper struct {
	conn sqlx.SqlConn
}

// NewTransactionHelper 创建事务辅助工具
func NewTransactionHelper(conn sqlx.SqlConn) *TransactionHelper {
	return &TransactionHelper{
		conn: conn,
	}
}

// GetUserModelWithSession 获取带会话的用户模型
func (h *TransactionHelper) GetUserModelWithSession(session sqlx.Session) user.UserModel {
	return user.NewUserModel(sqlx.NewSqlConnFromSession(session))
}

// GetEmployeeModelWithSession 获取带会话的员工模型
func (h *TransactionHelper) GetEmployeeModelWithSession(session sqlx.Session) user.EmployeeModel {
	return user.NewEmployeeModel(sqlx.NewSqlConnFromSession(session))
}

// GetCompanyModelWithSession 获取带会话的公司模型
func (h *TransactionHelper) GetCompanyModelWithSession(session sqlx.Session) company.CompanyModel {
	return company.NewCompanyModel(sqlx.NewSqlConnFromSession(session))
}

// GetDepartmentModelWithSession 获取带会话的部门模型
func (h *TransactionHelper) GetDepartmentModelWithSession(session sqlx.Session) company.DepartmentModel {
	return company.NewDepartmentModel(sqlx.NewSqlConnFromSession(session))
}

// GetPositionModelWithSession 获取带会话的职位模型
func (h *TransactionHelper) GetPositionModelWithSession(session sqlx.Session) company.PositionModel {
	return company.NewPositionModel(sqlx.NewSqlConnFromSession(session))
}

// GetRoleModelWithSession 获取带会话的角色模型
func (h *TransactionHelper) GetRoleModelWithSession(session sqlx.Session) role.RoleModel {
	return role.NewRoleModel(sqlx.NewSqlConnFromSession(session))
}

// GetTaskModelWithSession 获取带会话的任务模型
func (h *TransactionHelper) GetTaskModelWithSession(session sqlx.Session) task.TaskModel {
	return task.NewTaskModel(sqlx.NewSqlConnFromSession(session))
}

// GetTaskNodeModelWithSession 获取带会话的任务节点模型
func (h *TransactionHelper) GetTaskNodeModelWithSession(session sqlx.Session) task.TaskNodeModel {
	return task.NewTaskNodeModel(sqlx.NewSqlConnFromSession(session))
}

// GetTaskLogModelWithSession 获取带会话的任务日志模型
func (h *TransactionHelper) GetTaskLogModelWithSession(session sqlx.Session) task.TaskLogModel {
	return task.NewTaskLogModel(sqlx.NewSqlConnFromSession(session))
}

// GetTaskHandoverModelWithSession 获取带会话的任务交接模型
func (h *TransactionHelper) GetTaskHandoverModelWithSession(session sqlx.Session) task.TaskHandoverModel {
	return task.NewTaskHandoverModel(sqlx.NewSqlConnFromSession(session))
}

// GetNotificationModelWithSession 获取带会话的通知模型
func (h *TransactionHelper) GetNotificationModelWithSession(session sqlx.Session) user_auth.NotificationModel {
	return user_auth.NewNotificationModel(sqlx.NewSqlConnFromSession(session))
}
