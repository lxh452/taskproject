package employee

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"task_Project/model/user"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ApproveJoinApplicationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 审批加入公司申请
func NewApproveJoinApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveJoinApplicationLogic {
	return &ApproveJoinApplicationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApproveJoinApplicationLogic) ApproveJoinApplication(req *types.ApproveJoinApplicationRequest) (resp *types.BaseResponse, err error) {
	// 获取当前用户ID
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 获取当前员工信息
	approverEmployee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, userID)
	if err != nil || approverEmployee == nil {
		return utils.Response.BusinessError("employee_not_in_company"), nil
	}

	// 获取申请信息
	application, err := l.svcCtx.JoinApplicationModel.FindOne(l.ctx, req.ApplicationID)
	if err != nil {
		logx.Errorf("查询申请失败: %v", err)
		return utils.Response.BusinessError("application_not_found"), nil
	}

	// 检查申请状态
	if application.Status != user.JoinApplicationStatusPending {
		return utils.Response.BusinessError("application_processed"), nil
	}

	// 检查审批人是否属于该公司
	if approverEmployee.CompanyId != application.CompanyId {
		return utils.Response.BusinessError("no_permission_approve"), nil
	}

	// 检查审批权限：创始人或人事部门或管理岗
	company, _ := l.svcCtx.CompanyModel.FindOne(l.ctx, application.CompanyId)
	isFounder := company != nil && company.Owner == userID
	isHR := false
	isManager := false

	if approverEmployee.DepartmentId.Valid {
		dept, _ := l.svcCtx.DepartmentModel.FindOne(l.ctx, approverEmployee.DepartmentId.String)
		if dept != nil && dept.DepartmentCode.Valid && dept.DepartmentCode.String == "HR" {
			isHR = true
		}
	}

	if approverEmployee.PositionId.Valid {
		pos, _ := l.svcCtx.PositionModel.FindOne(l.ctx, approverEmployee.PositionId.String)
		if pos != nil && pos.IsManagement == 1 {
			isManager = true
		}
	}

	if !isFounder && !isHR && !isManager {
		return utils.Response.BusinessError("only_admin_can_approve"), nil
	}

	// 获取申请人信息
	applicantUser, err := l.svcCtx.UserModel.FindOne(l.ctx, application.UserId)
	if err != nil {
		logx.Errorf("查询申请人信息失败: %v", err)
		return utils.Response.InternalError("查询申请人信息失败"), nil
	}

	if req.Approved {
		// 通过审批：创建员工记录（可选指定部门/职位）
		var createdEmployeeID string
		createdEmployeeID, err = l.approveAndCreateEmployee(application, applicantUser, approverEmployee.Id, req.Note, company.Name, req.DepartmentID, req.PositionID)
		if err != nil {
			logx.Errorf("审批通过处理失败: %v", err)
			return utils.Response.InternalError("处理失败"), nil
		}

		// 邮件通知申请人 - 通过
		l.sendResultEmail(applicantUser, true, company.Name, req.Note, createdEmployeeID)

		logx.Infof("审批通过: applicationId=%s, userId=%s", req.ApplicationID, application.UserId)
		return utils.Response.Success(map[string]interface{}{
			"message":    "已通过申请，员工已创建",
			"employeeId": createdEmployeeID,
			"companyId":  application.CompanyId,
		}), nil
	} else {
		// 拒绝审批
		err = l.svcCtx.JoinApplicationModel.UpdateStatus(
			l.ctx,
			req.ApplicationID,
			user.JoinApplicationStatusRejected,
			approverEmployee.Id,
			req.Note,
		)
		if err != nil {
			logx.Errorf("更新申请状态失败: %v", err)
			return utils.Response.InternalError("处理失败"), nil
		}

		// 邮件通知申请人 - 拒绝
		l.sendResultEmail(applicantUser, false, company.Name, req.Note, "")

		// 发送拒绝通知
		go l.notifyApplicant(application.UserId, company.Name, false, req.Note)

		logx.Infof("审批拒绝: applicationId=%s, userId=%s", req.ApplicationID, application.UserId)
		return utils.Response.Success(map[string]interface{}{
			"message": "已拒绝申请",
		}), nil
	}
}

// approveAndCreateEmployee 通过审批并创建员工
func (l *ApproveJoinApplicationLogic) approveAndCreateEmployee(application *user.JoinApplication, applicantUser *user.User, approverID, note, companyName, specifiedDeptID, specifiedPosID string) (string, error) {
	// 获取默认部门和职位（优先非HR部门，避免新员工全部落在人事部）
	departmentID := ""
	positionID := ""

	departments, _ := l.svcCtx.DepartmentModel.FindByCompanyID(l.ctx, application.CompanyId)

	// 如果审批人指定了部门，校验是否属于该公司
	if specifiedDeptID != "" {
		dept, _ := l.svcCtx.DepartmentModel.FindOne(l.ctx, specifiedDeptID)
		if dept == nil || dept.CompanyId != application.CompanyId {
			return "", fmt.Errorf("指定部门不存在或不属于该公司")
		}
		departmentID = specifiedDeptID
	}

	// 先挑非HR部门（如果存在），否则第一个部门
	if departmentID == "" {
		for _, dept := range departments {
			if dept.DepartmentCode.Valid && strings.EqualFold(dept.DepartmentCode.String, "HR") {
				continue
			}
			departmentID = dept.Id
			break
		}
		if departmentID == "" && len(departments) > 0 {
			departmentID = departments[0].Id
		}
	}

	// 查找该部门下的默认职位，优先助理(AST)或第一个
	if departmentID != "" {
		positions, _ := l.svcCtx.PositionModel.FindByDepartmentID(l.ctx, departmentID)

		// 如果审批人指定了职位，校验职位所属部门
		if specifiedPosID != "" {
			for _, pos := range positions {
				if pos.Id == specifiedPosID {
					positionID = specifiedPosID
					break
				}
			}
			if positionID == "" {
				return "", fmt.Errorf("指定职位不存在或不属于该部门")
			}
		}

		for _, pos := range positions {
			if pos.PositionCode.Valid && strings.EqualFold(pos.PositionCode.String, "AST") {
				positionID = pos.Id
				break
			}
		}
		if positionID == "" && len(positions) > 0 {
			positionID = positions[0].Id
		}
	}

	// 使用事务创建员工
	employeeID := utils.Common.GenId("emp")
	empCode := employeeID
	if len(employeeID) > 6 {
		empCode = "EMP-" + strings.ToUpper(employeeID[len(employeeID)-6:])
	} else {
		empCode = "EMP-" + strings.ToUpper(employeeID)
	}
	realName := "新员工"
	if applicantUser.RealName.Valid && applicantUser.RealName.String != "" {
		realName = applicantUser.RealName.String
	}

	err := l.svcCtx.TransactionService.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 更新申请状态
		if err := l.svcCtx.JoinApplicationModel.UpdateStatus(
			ctx,
			application.Id,
			user.JoinApplicationStatusApproved,
			approverID,
			note,
		); err != nil {
			return err
		}

		// 2. 创建员工记录
		employeeInfo := &user.Employee{
			Id:           employeeID,
			UserId:       application.UserId,
			CompanyId:    application.CompanyId,
			DepartmentId: utils.Common.ToSqlNullString(departmentID),
			PositionId:   utils.Common.ToSqlNullString(positionID),
			EmployeeId:   empCode,
			RealName:     realName,
			Email:        applicantUser.Email,
			Phone:        applicantUser.Phone,
			Skills:       sql.NullString{},
			RoleTags:     sql.NullString{},
			HireDate:     sql.NullTime{Time: time.Now(), Valid: true},
			Status:       1,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}

		empModelWithSession := l.svcCtx.TransactionHelper.GetEmployeeModelWithSession(session)
		if _, err := empModelWithSession.Insert(ctx, employeeInfo); err != nil {
			return err
		}

		// 3.5 自动推断并设置直属上级
		approverFinder := utils.NewApproverFinder(l.svcCtx.EmployeeModel, l.svcCtx.DepartmentModel, l.svcCtx.CompanyModel)
		supervisorID, inferErr := approverFinder.InferSupervisor(ctx, employeeID)
		if inferErr == nil && supervisorID != "" {
			if updateErr := empModelWithSession.UpdateSupervisor(ctx, employeeID, supervisorID); updateErr != nil {
				logx.Errorf("设置直属上级失败: %v", updateErr)
			} else {
				logx.Infof("员工 %s 的直属上级已自动设置为 %s", employeeID, supervisorID)
			}
		}

		// 4. 更新用户加入公司状态
		userModelWithSession := l.svcCtx.TransactionHelper.GetUserModelWithSession(session)
		if err := userModelWithSession.UpdateHasJoinedCompany(ctx, application.UserId, true); err != nil {
			return err
		}

		// 4. 更新职位员工数
		if positionID != "" {
			posModelWithSession := l.svcCtx.TransactionHelper.GetPositionModelWithSession(session)
			posInfo, _ := l.svcCtx.PositionModel.FindOne(ctx, positionID)
			if posInfo != nil {
				_ = posModelWithSession.UpdateCurrentEmployees(ctx, positionID, int(posInfo.CurrentEmployees)+1)
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// 发送通知给申请人
	go l.notifyApplicant(application.UserId, companyName, true, "")

	return empCode, nil
}

// sendResultEmail 向申请人发送审批结果邮件
func (l *ApproveJoinApplicationLogic) sendResultEmail(applicantUser *user.User, approved bool, companyName, note, employeeCode string) {
	if l.svcCtx.EmailMQService == nil || applicantUser == nil || !applicantUser.Email.Valid || applicantUser.Email.String == "" {
		return
	}

	subject := ""
	body := ""
	if approved {
		subject = "加入申请已通过"
		body = fmt.Sprintf("恭喜，您加入 %s 的申请已通过。您的员工编号：%s。", companyName, employeeCode)
	} else {
		subject = "加入申请未通过"
		body = fmt.Sprintf("抱歉，您加入 %s 的申请未通过。", companyName)
		if note != "" {
			body += fmt.Sprintf(" 原因：%s", note)
		}
	}

	emailEvent := &svc.EmailEvent{
		EventType: "join.application.result",
		To:        []string{applicantUser.Email.String},
		Subject:   subject,
		Body:      body,
		IsHTML:    false,
	}
	if err := l.svcCtx.EmailMQService.PublishEmailEvent(l.ctx, emailEvent); err != nil {
		logx.Errorf("发送加入申请结果邮件失败: %v", err)
	}
}

// notifyApplicant 通知申请人审批结果
func (l *ApproveJoinApplicationLogic) notifyApplicant(applicantUserID, companyName string, approved bool, note string) {
	ctx := context.Background()

	if approved {
		// 审批通过：获取申请人的员工信息
		employee, _ := l.svcCtx.EmployeeModel.FindByUserID(ctx, applicantUserID)
		if employee == nil {
			// 如果没有员工记录，说明创建员工失败，跳过通知
			logx.Infof("申请人尚无员工记录，跳过通知: userId=%s", applicantUserID)
			return
		}

		if l.svcCtx.NotificationMQService != nil {
			title := "加入申请已通过"
			content := "恭喜！您加入 " + companyName + " 的申请已通过，欢迎加入团队！"

			event := &svc.NotificationEvent{
				EventType:   "join.result",
				EmployeeIDs: []string{employee.Id},
				Title:       title,
				Content:     content,
				Type:        0, // 系统通知
				Category:    "join_application",
				Priority:    2, // 高优先级
				RelatedType: "join_application",
			}
			if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
				logx.Errorf("发送审批结果通知失败: %v", err)
			}
		}
	} else {
		// 审批拒绝：直接发送系统通知给用户（不需要员工记录）
		// 注意：这里需要使用用户ID而不是员工ID，因为拒绝时没有员工记录
		// 但当前的NotificationMQService只支持EmployeeIDs，所以拒绝通知只能通过邮件发送
		// 系统通知在这里无法发送，因为用户还没有员工身份
		logx.Infof("申请被拒绝，用户 %s 尚无员工记录，无法发送系统通知（已通过邮件通知）", applicantUserID)
	}
}
