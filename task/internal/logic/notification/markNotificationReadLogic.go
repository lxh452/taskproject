package notification

import (
	"context"
	"errors"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type MarkNotificationReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标记通知为已读
func NewMarkNotificationReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkNotificationReadLogic {
	return &MarkNotificationReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkNotificationReadLogic) MarkNotificationRead(req *types.MarkNotificationReadRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.NotificationID) {
		return utils.Response.BusinessError("通知ID不能为空"), nil
	}

	// 2. 获取当前用户ID
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 3. 获取员工ID
	employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, currentUserID)
	if err != nil {
		l.Logger.Errorf("查询员工失败: %v", err)
		return utils.Response.BusinessError("用户未绑定员工信息"), nil
	}

	// 4. 查询通知是否存在
	notification, err := l.svcCtx.NotificationModel.FindOne(l.ctx, req.NotificationID)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return utils.Response.BusinessError("通知不存在"), nil
		}
		l.Logger.Errorf("查询通知失败: %v", err)
		return nil, err
	}

	// 5. 验证权限：只能标记自己的通知为已读
	if notification.EmployeeId != employee.EmployeeId {
		return utils.Response.BusinessError("无权操作其他员工的通知"), nil
	}

	// 6. 检查是否已经标记为已读
	if notification.IsRead == 1 {
		return utils.Response.Success(map[string]interface{}{
			"notificationId": req.NotificationID,
			"message":        "通知已经是已读状态",
		}), nil
	}

	// 7. 更新通知为已读状态
	err = l.svcCtx.NotificationModel.UpdateReadStatus(l.ctx, req.NotificationID, 1)
	if err != nil {
		l.Logger.Errorf("更新通知已读状态失败: %v", err)
		return nil, err
	}

	return utils.Response.Success(map[string]interface{}{
		"notificationId": req.NotificationID,
		"message":        "通知已标记为已读",
	}), nil
}
