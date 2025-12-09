// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package employee

import (
	"context"
	"errors"
	"fmt"
	"time"

	"task_Project/model/company"
	"task_Project/model/user"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateEmployeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建员工
func NewCreateEmployeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateEmployeeLogic {
	return &CreateEmployeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateEmployeeLogic) CreateEmployee(req *types.CreateEmployeeRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	requiredFields := map[string]string{
		"用户ID": req.UserID,
		"公司ID": req.CompanyID,
		"部门ID": req.DepartmentID,
		"职位ID": req.PositionID,
		"真实姓名": req.RealName,
	}
	if errors := utils.Validator.ValidateRequired(requiredFields); len(errors) > 0 {
		return utils.Response.BusinessError("employee_required_fields"), nil
	}

	// 获取当前用户ID
	if _, ok := utils.Common.GetCurrentUserID(l.ctx); !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 验证用户是否存在
	if _, err = l.svcCtx.UserModel.FindOne(l.ctx, req.UserID); err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return utils.Response.BusinessError("user_not_found"), nil
		}
		logx.Errorf("查询用户失败: %v", err)
		return utils.Response.InternalError("查询用户失败"), nil
	}

	// 验证公司是否存在
	companyInfo, err := l.svcCtx.CompanyModel.FindOne(l.ctx, req.CompanyID)
	if err != nil {
		if errors.Is(err, company.ErrNotFound) {
			return utils.Response.BusinessError("company_not_found"), nil
		}
		logx.Errorf("查询公司失败: %v", err)
		return utils.Response.InternalError("查询公司失败"), nil
	}

	// 验证部门是否存在
	departmentInfo, err := l.svcCtx.DepartmentModel.FindOne(l.ctx, req.DepartmentID)
	if err != nil {
		if errors.Is(err, company.ErrNotFound) {
			return utils.Response.BusinessError("department_not_found"), nil
		}
		logx.Errorf("查询部门失败: %v", err)
		return utils.Response.InternalError("查询部门失败"), nil
	}

	// 验证职位是否存在
	positionInfo, err := l.svcCtx.PositionModel.FindOne(l.ctx, req.PositionID)
	if err != nil {
		if errors.Is(err, company.ErrNotFound) {
			return utils.Response.BusinessError("position_not_found"), nil
		}
		logx.Errorf("查询职位失败: %v", err)
		return utils.Response.InternalError("查询职位失败"), nil
	}

	// 检查用户是否已经是该公司的员工
	existingEmployee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, req.UserID)
	if err == nil && existingEmployee.CompanyId == req.CompanyID {
		return utils.Response.BusinessError("employee_already_exists"), nil
	}

	// 检查员工编号是否已存在
	if req.EmployeeID != "" {
		_, err = l.svcCtx.EmployeeModel.FindByEmployeeID(l.ctx, req.EmployeeID)
		if err == nil {
			return utils.Response.BusinessError("employee_id_exists"), nil
		}
	}

	// 生成员工ID
	employeeID := utils.Common.GenerateID()

	// 创建员工
	employeeInfo := &user.Employee{
		Id:           employeeID,
		UserId:       req.UserID,
		CompanyId:    req.CompanyID,
		DepartmentId: utils.Common.ToSqlNullString(req.DepartmentID),
		PositionId:   utils.Common.ToSqlNullString(req.PositionID),
		EmployeeId:   req.EmployeeID,
		RealName:     req.RealName,
		Email:        utils.Common.ToSqlNullString(req.WorkEmail),
		Phone:        utils.Common.ToSqlNullString(req.WorkPhone),
		Skills:       utils.Common.ToSqlNullString(req.Skills),
		RoleTags:     utils.Common.ToSqlNullString(req.RoleTags),
		HireDate:     utils.Common.ToSqlNullTime(req.HireDate),
		Status:       1, // 正常状态
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}

	// 使用事务创建员工
	err = l.svcCtx.TransactionService.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 创建带会话的模型
		employeeModelWithSession := l.svcCtx.TransactionHelper.GetEmployeeModelWithSession(session)
		positionModelWithSession := l.svcCtx.TransactionHelper.GetPositionModelWithSession(session)

		_, err := employeeModelWithSession.Insert(ctx, employeeInfo)
		if err != nil {
			return err
		}

		// 更新职位的当前员工数
		err = positionModelWithSession.UpdateCurrentEmployees(ctx, req.PositionID, int(positionInfo.CurrentEmployees)+1)
		if err != nil {
			return err
		}

		// 注意：权限验证直接通过职位->角色->权限查询，无需同步到user_permission表

		return nil
	})

	if err != nil {
		logx.Errorf("创建员工失败: %v", err)
		return utils.Response.InternalError("创建员工失败"), nil
	}

	// 更新用户的加入公司状态
	if updateErr := l.svcCtx.UserModel.UpdateHasJoinedCompany(l.ctx, req.UserID, true); updateErr != nil {
		logx.Errorf("更新用户加入公司状态失败: %v", updateErr)
		// 不影响主流程，继续执行
	}

	// 发送入职通知邮件（通过消息队列）
	go func() {
		ctx := context.Background()
		// 发送入职邮件给新员工
		if req.WorkEmail != "" && l.svcCtx.EmailService != nil {
			onboardingTime := time.Now().Format("2006-01-02 15:04:05")
			if err := l.svcCtx.EmailService.SendOnboardingEmail(ctx, req.WorkEmail, req.RealName, companyInfo.Name, departmentInfo.DepartmentName, positionInfo.PositionName, onboardingTime); err != nil {
				logx.Errorf("发送入职通知邮件失败: %v", err)
			}
		}

		// 查询该公司所有员工，发送新员工入职通知
		employees, err := l.svcCtx.EmployeeModel.FindByCompanyID(ctx, req.CompanyID)
		if err != nil {
			logx.Errorf("查询公司员工失败: %v", err)
			return
		}

		employeeIDs := make([]string, 0, len(employees))
		emails := make([]string, 0, len(employees))
		for _, emp := range employees {
			// 排除新入职的员工本人
			if emp.Id == employeeID {
				continue
			}
			employeeIDs = append(employeeIDs, emp.Id)
			if emp.Email.Valid && emp.Email.String != "" {
				emails = append(emails, emp.Email.String)
			}
		}

		// 发布通知事件 - 通知公司其他员工有新同事入职
		if l.svcCtx.NotificationMQService != nil && len(employeeIDs) > 0 {
			notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
				svc.EmployeeCreated,
				employeeIDs,
				employeeID,
			)
			notificationEvent.Title = "新员工入职通知"
			notificationEvent.Content = fmt.Sprintf("欢迎新同事 %s 加入 %s 部门，职位：%s", req.RealName, departmentInfo.DepartmentName, positionInfo.PositionName)
			notificationEvent.Category = "employee"
			if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, notificationEvent); err != nil {
				logx.Errorf("发布新员工入职通知事件失败: %v", err)
			}
		}

		// 发布邮件事件 - 通知公司其他员工有新同事入职
		if l.svcCtx.EmailMQService != nil && len(emails) > 0 {
			emailEvent := &svc.EmailEvent{
				EventType: svc.EmployeeCreated,
				To:        emails,
				Subject:   "新员工入职通知",
				Body:      fmt.Sprintf("欢迎新同事 %s 加入 %s 部门，职位：%s。请各位同事多多关照！", req.RealName, departmentInfo.DepartmentName, positionInfo.PositionName),
				IsHTML:    false,
			}
			if err := l.svcCtx.EmailMQService.PublishEmailEvent(ctx, emailEvent); err != nil {
				logx.Errorf("发布新员工入职邮件事件失败: %v", err)
			}
		}
	}()

	return utils.Response.SuccessWithKey("create", map[string]interface{}{
		"employeeId": employeeID,
		"realName":   req.RealName,
		"companyId":  req.CompanyID,
	}), nil
}
