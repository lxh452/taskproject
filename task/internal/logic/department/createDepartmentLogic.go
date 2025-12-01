// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package department

import (
	"context"
	"fmt"
	"time"

	"task_Project/model/company"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateDepartmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建部门
func NewCreateDepartmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDepartmentLogic {
	return &CreateDepartmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDepartmentLogic) CreateDepartment(req *types.CreateDepartmentRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.DepartmentName) {
		return utils.Response.ValidationError("部门名称不能为空"), nil
	}
	if validator.IsEmpty(req.CompanyID) {
		return utils.Response.ValidationError("公司ID不能为空"), nil
	}

	// 获取当前用户信息
	common := utils.NewCommon()
	if _, ok := common.GetCurrentUserID(l.ctx); !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 检查公司是否存在
	//todo 需要更改方法，防止查到已经删除的公司
	_, err = l.svcCtx.CompanyModel.FindOne(l.ctx, req.CompanyID)
	if err != nil {
		logx.Errorf("查询公司失败: %v", err)
		return utils.Response.ErrorWithKey("company_not_found"), nil
	}

	// 检查父部门是否存在（如果指定了父部门）
	//todo 需要更改方法，防止查到已经删除的部门
	if !validator.IsEmpty(req.ParentID) {
		_, err = l.svcCtx.DepartmentModel.FindOne(l.ctx, req.ParentID)
		if err != nil {
			logx.Errorf("查询父部门失败: %v", err)
			return utils.Response.ErrorWithKey("department_not_found"), nil
		}
	}

	// 生成部门ID
	departmentID := common.GenerateID()

	// 创建部门
	department := &company.Department{
		Id:             departmentID,
		CompanyId:      req.CompanyID,
		DepartmentName: req.DepartmentName,
		DepartmentCode: utils.Common.ToSqlNullString(req.DepartmentCode),
		ParentId:       utils.Common.ToSqlNullString(req.ParentID),
		ManagerId:      utils.Common.ToSqlNullString(req.ManagerID),
		Description:    utils.Common.ToSqlNullString(req.Description),
		Status:         1, // 正常状态
		CreateTime:     time.Now(),
		UpdateTime:     time.Now(),
	}

	_, err = l.svcCtx.DepartmentModel.Insert(l.ctx, department)
	if err != nil {
		logx.Errorf("创建部门失败: %v", err)
		return utils.Response.InternalError("创建部门失败"), err
	}

	// 发送通知和邮件给公司所有员工 - 创建了新的部门
	go func() {
		ctx := context.Background()
		// 查询该公司所有员工
		employees, err := l.svcCtx.EmployeeModel.FindByCompanyID(ctx, req.CompanyID)
		if err != nil {
			logx.Errorf("查询公司员工失败: %v", err)
			return
		}

		employeeIDs := make([]string, 0, len(employees))
		emails := make([]string, 0, len(employees))
		for _, emp := range employees {
			employeeIDs = append(employeeIDs, emp.Id)
			if emp.Email.Valid && emp.Email.String != "" {
				emails = append(emails, emp.Email.String)
			}
		}

		// 发布通知事件
		if l.svcCtx.NotificationMQService != nil && len(employeeIDs) > 0 {
			notificationEvent := l.svcCtx.NotificationMQService.NewNotificationEvent(
				svc.DepartmentCreated,
				employeeIDs,
				departmentID,
			)
			notificationEvent.Title = "新部门创建通知"
			notificationEvent.Content = fmt.Sprintf("公司新创建了部门：%s", req.DepartmentName)
			notificationEvent.Category = "department"
			if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, notificationEvent); err != nil {
				logx.Errorf("发布部门创建通知事件失败: %v", err)
			}
		}

		// 发布邮件事件
		if l.svcCtx.EmailMQService != nil && len(emails) > 0 {
			emailEvent := &svc.EmailEvent{
				EventType: svc.DepartmentCreated,
				To:        emails,
				Subject:   "新部门创建通知",
				Body:      fmt.Sprintf("公司新创建了部门：%s，如有需要请联系管理员了解详情。", req.DepartmentName),
				IsHTML:    false,
			}
			if err := l.svcCtx.EmailMQService.PublishEmailEvent(ctx, emailEvent); err != nil {
				logx.Errorf("发布部门创建邮件事件失败: %v", err)
			}
		}
	}()

	return utils.Response.SuccessWithKey("departmentId", departmentID), nil
}
