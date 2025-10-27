// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"
	"errors"
	"time"

	"task_Project/model/user"
	"task_Project/task/internal/middleware"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户注册
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	requiredFields := map[string]string{
		"用户名":  req.Username,
		"密码":   req.Password,
		"真实姓名": req.RealName,
	}

	if errors := utils.Validator.ValidateRequired(requiredFields); len(errors) > 0 {
		return utils.Response.ValidationError(errors[0]), nil
	}

	// 验证密码强度
	if passwordErr := utils.Validator.ValidatePassword(req.Password); passwordErr != "" {
		return utils.Response.ValidationError(passwordErr), nil
	}

	// 验证邮箱格式
	if req.Email != "" {
		if emailErr := utils.Validator.ValidateEmail(req.Email); emailErr != "" {
			return utils.Response.ValidationError(emailErr), nil
		}
	}

	// 验证手机号格式
	if req.Phone != "" {
		if phoneErr := utils.Validator.ValidatePhone(req.Phone); phoneErr != "" {
			return utils.Response.ValidationError(phoneErr), nil
		}
	}

	// 检查用户名是否已存在
	_, err = l.svcCtx.UserModel.FindByUsername(l.ctx, req.Username)
	if err == nil {
		return utils.Response.BusinessError("username_exists"), nil
	}
	if !errors.Is(err, user.ErrNotFound) {
		logx.Errorf("检查用户名失败: %v", err)
		return utils.Response.InternalError("检查用户名失败"), nil
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		_, err = l.svcCtx.UserModel.FindByEmail(l.ctx, req.Email)
		if err == nil {
			return utils.Response.BusinessError("email_exists"), nil
		}
		if !errors.Is(err, user.ErrNotFound) {
			logx.Errorf("检查邮箱失败: %v", err)
			return utils.Response.InternalError("检查邮箱失败"), nil
		}
	}

	// 检查手机号是否已存在
	if req.Phone != "" {
		_, err = l.svcCtx.UserModel.FindByPhone(l.ctx, req.Phone)
		if err == nil {
			return utils.Response.BusinessError("phone_exists"), nil
		}
		if !errors.Is(err, user.ErrNotFound) {
			logx.Errorf("检查手机号失败: %v", err)
			return utils.Response.InternalError("检查手机号失败"), nil
		}
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logx.Errorf("密码加密失败: %v", err)
		return utils.Response.InternalError("密码加密失败"), nil
	}

	// 生成用户ID
	common := utils.NewCommon()
	userID := common.GenerateID()

	// 创建用户
	userInfo := &user.User{
		Id:           userID,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		Email:        utils.Common.ToSqlNullString(req.Email),
		Phone:        utils.Common.ToSqlNullString(req.Phone),
		RealName:     utils.Common.ToSqlNullString(req.RealName),
		Status:       1, // 正常状态
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}

	// 使用事务创建用户
	err = l.svcCtx.TransactionService.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 创建带会话的用户模型
		userModelWithSession := l.svcCtx.TransactionHelper.GetUserModelWithSession(session)
		_, err := userModelWithSession.Insert(ctx, userInfo)
		return err
	})

	if err != nil {
		logx.Errorf("创建用户失败: %v", err)
		return utils.Response.InternalError("注册失败"), nil
	}

	// 发送注册成功邮件
	go func() {
		if req.Email != "" {
			emailMsg := middleware.EmailMessage{
				To:      []string{req.Email},
				Subject: "注册成功通知",
				Body:    "欢迎注册企业任务系统！您的账户已成功创建，用户名：" + req.Username,
				IsHTML:  false,
			}
			if err := l.svcCtx.EmailMiddleware.SendEmail(context.Background(), emailMsg); err != nil {
				logx.Errorf("发送注册成功邮件失败: %v", err)
			}
		}
	}()

	// 发送注册成功短信
	go func() {
		if req.Phone != "" {
			if err := l.svcCtx.SMSMiddleware.SendNotificationSMS(context.Background(), req.Phone, "欢迎注册企业任务系统！您的账户已成功创建。"); err != nil {
				logx.Errorf("发送注册成功短信失败: %v", err)
			}
		}
	}()

	return utils.Response.SuccessWithKey("register", map[string]interface{}{
		"userId":   userID,
		"username": req.Username,
		"realName": req.RealName,
	}), nil
}
