package notification

import (
	"context"
	"time"

	"task_Project/model/user_auth"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateNotificationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateNotificationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateNotificationLogic {
	return &CreateNotificationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateNotification 创建通知（需要用户认证）
// 这里如果任务创建好后通知每个节点负责人
// 或者任务节点创建好后通知每个执行人
func (l *CreateNotificationLogic) CreateNotification(req *types.CreateNotificationRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.EmployeeID) {
		return utils.Response.BusinessError("员工ID不能为空"), nil
	}
	if validator.IsEmpty(req.Title) {
		return utils.Response.BusinessError("通知标题不能为空"), nil
	}
	if validator.IsEmpty(req.Content) {
		return utils.Response.BusinessError("通知内容不能为空"), nil
	}

	// 2. 获取当前用户ID（用于权限验证）
	_, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 调用内部方法创建通知
	return l.CreateNotificationInternal(req)
}

// CreateNotificationInternal 内部创建通知方法（跳过用户认证，用于系统内部调用）
func (l *CreateNotificationLogic) CreateNotificationInternal(req *types.CreateNotificationRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.EmployeeID) {
		return utils.Response.BusinessError("员工ID不能为空"), nil
	}
	if validator.IsEmpty(req.Title) {
		return utils.Response.BusinessError("通知标题不能为空"), nil
	}
	if validator.IsEmpty(req.Content) {
		return utils.Response.BusinessError("通知内容不能为空"), nil
	}

	// 2. 验证员工是否存在，并获取员工主键ID
	emp, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, req.EmployeeID)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("查询员工失败: %v %v", err, req.EmployeeID)
		return utils.Response.BusinessError("员工不存在"), nil
	}
	// 使用员工主键 Id（通知表使用员工主键存储）
	actualEmployeeID := emp.Id
	l.Logger.WithContext(l.ctx).Infof("创建通知给员工: %s (ID: %s)", emp.RealName, actualEmployeeID)

	// 3. 生成通知ID
	notificationID := utils.Common.GenId("notification")

	// 4. 创建通知
	notification := &user_auth.Notification{
		Id:          notificationID,
		EmployeeId:  actualEmployeeID,
		Title:       req.Title,
		Content:     req.Content,
		Type:        int64(req.Type),
		Priority:    int64(req.Priority),
		RelatedId:   utils.Common.ToSqlNullString(req.RelatedID),
		RelatedType: utils.Common.ToSqlNullString(req.RelatedType),
		IsRead:      int64(0), // 0表示未读
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}

	_, err = l.svcCtx.NotificationModel.Insert(l.ctx, notification)
	if err != nil {
		l.Logger.WithContext(l.ctx).Errorf("创建通知失败: %v", err)
		return nil, err
	}

	return utils.Response.Success(map[string]interface{}{
		"notificationId": notificationID,
		"message":        "通知创建成功",
	}), nil
}

// BatchCreateNotificationRequest 批量创建通知请求
type BatchCreateNotificationRequest struct {
	EmployeeIDs []string // 员工ID列表
	Title       string
	Content     string
	Type        int
	Priority    int
	RelatedID   string
	RelatedType string
}

// CreateNotificationBatch 批量创建通知（异步，用于系统内部调用）
// 在 goroutine 中为每个员工创建通知，不阻塞主进程
func (l *CreateNotificationLogic) CreateNotificationBatch(req *BatchCreateNotificationRequest) {
	logx.Infof("[BatchNotification] 开始批量创建通知: employeeIDs=%v, title=%s", req.EmployeeIDs, req.Title)

	if len(req.EmployeeIDs) == 0 {
		logx.Info("[BatchNotification] 没有员工需要通知，跳过")
		return
	}

	// 复制数据，避免闭包引用问题
	employeeIDs := make([]string, len(req.EmployeeIDs))
	copy(employeeIDs, req.EmployeeIDs)
	title := req.Title
	content := req.Content
	notificationType := req.Type
	priority := req.Priority
	relatedID := req.RelatedID
	relatedType := req.RelatedType
	svcCtx := l.svcCtx

	// 异步处理，不阻塞主进程
	go func() {
		ctx := context.Background()
		successCount := 0

		logx.Infof("[BatchNotification] goroutine 开始处理 %d 个员工", len(employeeIDs))

		for _, employeeID := range employeeIDs {
			logx.Infof("[BatchNotification] 正在处理员工: %s", employeeID)

			// 查询员工信息
			emp, err := svcCtx.EmployeeModel.FindOne(ctx, employeeID)
			if err != nil {
				logx.Errorf("[BatchNotification] 查询员工失败: employeeID=%s, err=%v", employeeID, err)
				continue
			}

			logx.Infof("[BatchNotification] 找到员工: %s (ID: %s)", emp.RealName, emp.Id)

			// 生成通知ID
			notificationID := utils.Common.GenId("notification")

			// 创建通知（使用员工主键 Id）
			notification := &user_auth.Notification{
				Id:          notificationID,
				EmployeeId:  emp.Id,
				Title:       title,
				Content:     content,
				Type:        int64(notificationType),
				Priority:    int64(priority),
				RelatedId:   utils.Common.ToSqlNullString(relatedID),
				RelatedType: utils.Common.ToSqlNullString(relatedType),
				IsRead:      int64(0),
				CreateTime:  time.Now(),
				UpdateTime:  time.Now(),
			}

			_, err = svcCtx.NotificationModel.Insert(ctx, notification)
			if err != nil {
				logx.Errorf("[BatchNotification] 创建通知失败: employeeID=%s, err=%v", employeeID, err)
				continue
			}

			logx.Infof("[BatchNotification] 成功为员工 %s (%s) 创建通知: %s", emp.RealName, emp.Id, notificationID)
			successCount++
		}

		logx.Infof("[BatchNotification] 批量创建通知完成: 成功 %d/%d", successCount, len(employeeIDs))
	}()

	logx.Info("[BatchNotification] 已启动异步处理")
}
