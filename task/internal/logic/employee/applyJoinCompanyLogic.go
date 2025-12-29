package employee

import (
	"context"
	"database/sql"
	"time"

	"task_Project/model/user"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApplyJoinCompanyLogic struct {
	logger logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 申请加入公司
func NewApplyJoinCompanyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApplyJoinCompanyLogic {
	return &ApplyJoinCompanyLogic{
		logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApplyJoinCompanyLogic) ApplyJoinCompany(req *types.ApplyJoinCompanyRequest) (resp *types.BaseResponse, err error) {
	// 获取当前用户ID
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 检查用户是否已经加入公司
	existingEmployee, _ := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, userID)
	if existingEmployee != nil {
		return utils.Response.BusinessError("您已经加入了公司，无法再次申请"), nil
	}

	// 检查是否有待审批的申请
	pendingApp, _ := l.svcCtx.JoinApplicationModel.FindPendingByUserId(l.ctx, userID)
	if pendingApp != nil {
		return utils.Response.BusinessError("您已有待审批的申请，请等待审批结果"), nil
	}

	// 验证邀请码
	if utils.Validator.IsEmpty(req.InviteCode) {
		return utils.Response.ValidationError("邀请码不能为空"), nil
	}

	// 解析邀请码
	inviteData, err := l.svcCtx.InviteCodeService.ParseInviteCode(l.ctx, req.InviteCode)
	if err != nil {
		l.logger.WithContext(l.ctx).Errorf("解析邀请码失败: code=%s, err=%v", req.InviteCode, err)
		return utils.Response.BusinessError(err.Error()), nil
	}

	// 验证公司是否存在
	company, err := l.svcCtx.CompanyModel.FindOne(l.ctx, inviteData.CompanyID)
	if err != nil {
		l.logger.WithContext(l.ctx).Errorf("查询公司失败: companyId=%s, err=%v", inviteData.CompanyID, err)
		return utils.Response.BusinessError("邀请码对应的公司不存在"), nil
	}

	// 检查公司状态
	if company.Status != 1 {
		return utils.Response.BusinessError("该公司已停用，无法申请加入"), nil
	}

	// 创建加入申请
	applicationID := utils.Common.GenId("ja")
	application := &user.JoinApplication{
		Id:          applicationID,
		UserId:      userID,
		CompanyId:   inviteData.CompanyID,
		InviteCode:  sql.NullString{String: req.InviteCode, Valid: true},
		ApplyReason: sql.NullString{String: req.ApplyReason, Valid: req.ApplyReason != ""},
		Status:      user.JoinApplicationStatusPending,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	_, err = l.svcCtx.JoinApplicationModel.Insert(l.ctx, application)
	if err != nil {
		l.logger.WithContext(l.ctx).Errorf("创建加入申请失败: %v", err)
		return utils.Response.InternalError("提交申请失败"), nil
	}

	// 使用邀请码（增加使用计数）
	if err := l.svcCtx.InviteCodeService.UseInviteCode(l.ctx, req.InviteCode); err != nil {
		l.logger.WithContext(l.ctx).Errorf("更新邀请码使用次数失败: %v", err)
		// 不影响主流程
	}

	// 发送通知给审批人（创始人或人事部门）
	go l.notifyApprovers(inviteData.CompanyID, company.Name, userID, applicationID)

	l.logger.WithContext(l.ctx).Infof("用户 %s 申请加入公司 %s, 申请ID: %s", userID, company.Name, applicationID)

	return utils.Response.Success(map[string]interface{}{
		"applicationId": applicationID,
		"companyId":     inviteData.CompanyID,
		"companyName":   company.Name,
		"status":        user.JoinApplicationStatusPending,
		"message":       "申请已提交，请等待审批",
	}), nil
}

// notifyApprovers 通知审批人
func (l *ApplyJoinCompanyLogic) notifyApprovers(companyID, companyName, applicantUserID, applicationID string) {
	ctx := context.Background()

	// 获取申请人信息
	applicant, _ := l.svcCtx.UserModel.FindOne(ctx, applicantUserID)
	applicantName := "新用户"
	if applicant != nil && applicant.RealName.Valid {
		applicantName = applicant.RealName.String
	}

	// 查找审批人：优先人事部门，否则创始人
	var approverEmployeeIDs []string

	// 1. 查找人事部门员工
	departments, _ := l.svcCtx.DepartmentModel.FindByCompanyID(ctx, companyID)
	for _, dept := range departments {
		if dept.DepartmentCode.Valid && dept.DepartmentCode.String == "HR" {
			// 找到人事部门，获取该部门所有员工
			hrEmployees, _ := l.svcCtx.EmployeeModel.FindByDepartmentID(ctx, dept.Id)
			for _, emp := range hrEmployees {
				approverEmployeeIDs = append(approverEmployeeIDs, emp.Id)
			}
			break
		}
	}

	// 2. 如果没有人事部门，通知创始人
	if len(approverEmployeeIDs) == 0 {
		company, _ := l.svcCtx.CompanyModel.FindOne(ctx, companyID)
		if company != nil {
			founderEmployee, _ := l.svcCtx.EmployeeModel.FindByUserID(ctx, company.Owner)
			if founderEmployee != nil {
				approverEmployeeIDs = append(approverEmployeeIDs, founderEmployee.Id)
			}
		}
	}

	// 发送通知
	if len(approverEmployeeIDs) > 0 && l.svcCtx.NotificationMQService != nil {
		event := &svc.NotificationEvent{
			EventType:   "join.application",
			EmployeeIDs: approverEmployeeIDs,
			Title:       "新员工申请加入",
			Content:     applicantName + " 申请加入 " + companyName + "，请及时审批",
			Type:        0, // 系统通知
			Category:    "join_application",
			Priority:    1, // 中优先级
			RelatedID:   applicationID,
			RelatedType: "join_application",
		}
		if err := l.svcCtx.NotificationMQService.PublishNotificationEvent(ctx, event); err != nil {
			l.logger.WithContext(l.ctx).Errorf("发送加入申请通知失败: %v", err)
		} else {
			l.logger.WithContext(l.ctx).Infof("已发送加入申请通知: applicationId=%s, approvers=%v", applicationID, approverEmployeeIDs)
		}
	}
}
