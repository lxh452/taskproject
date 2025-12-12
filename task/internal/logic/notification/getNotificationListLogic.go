package notification

import (
	"context"
	"fmt"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNotificationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNotificationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNotificationListLogic {
	return &GetNotificationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNotificationListLogic) GetNotificationList(req *types.NotificationListRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	validator := utils.NewValidator()
	page, pageSize, errs := validator.ValidatePageParams(req.Page, req.PageSize)
	if len(errs) > 0 {
		return utils.Response.ValidationError(errs[0]), nil
	}

	// 2. 获取当前用户ID
	currentUserID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 3. 获取员工ID
	employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, currentUserID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询员工失败: %v", err)
		return utils.Response.BusinessError("用户未绑定员工信息"), nil
	}

	// 4. 确定查询的员工ID（优先使用请求中的employeeId，否则使用当前员工ID）
	employeeID := req.EmployeeID
	if employeeID == "" {
		employeeID = employee.Id
	}

	// 5. 验证权限：如果查询的不是当前员工的通知，需要管理员权限
	if employeeID != employee.Id {
		// TODO: 可以在这里添加管理员权限验证
		// 暂时不允许查询其他员工的通知
		return utils.Response.BusinessError("无权查询其他员工的通知"), nil
	}

	// 6. 构建过滤条件
	// isRead 约定：
	//   0 -> 未读
	//   1 -> 已读
	//  其他（例如 -1）-> 不按已读状态过滤（全部）
	var isRead *int
	if req.IsRead == 0 || req.IsRead == 1 {
		isRead = &req.IsRead
	}

	var category *string
	if req.Category > 0 {
		// 将 category int 转换为 string（根据业务需求调整）
		categoryStr := fmt.Sprintf("%d", req.Category)
		category = &categoryStr
	}

	// 7. 查询通知列表
	notifications, total, err := l.svcCtx.NotificationModel.FindByEmployee(l.ctx, employeeID, isRead, category, page, pageSize)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询通知列表失败: %v", err)
		return nil, err
	}

	// 8. 转换为响应格式
	converter := utils.NewConverter()
	notificationInfos := converter.ToNotificationInfoList(notifications)

	// 9. 构建分页响应
	pageResponse := converter.ToPageResponse(notificationInfos, int(total), page, pageSize)

	return utils.Response.Success(pageResponse), nil
}
