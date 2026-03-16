package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"task_Project/model/user"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendVerificationCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendVerificationCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendVerificationCodeLogic {
	return &SendVerificationCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendVerificationCodeLogic) SendVerificationCode(req *types.SendVerificationCodeRequest) (resp *types.BaseResponse, err error) {
	// 1. 验证邮箱格式
	if !utils.Common.IsValidEmail(req.Email) {
		return utils.Response.BusinessError("email_format_invalid"), nil
	}

	// 2. 检查邮箱是否已被注册（如果是注册场景）
	if req.Type == "register" {
		existingUser, err := l.svcCtx.UserModel.FindByEmail(l.ctx, req.Email)
		// 如果查询成功且用户存在，说明邮箱已被注册
		if err == nil && existingUser != nil {
			return utils.Response.BusinessError("email_exists"), nil
		}
		// 如果是用户不存在的错误（ErrNotFound），这是正常的新用户注册场景，直接返回
		// 导入 user 包的 ErrNotFound 用于判断
		if err != nil && !errors.Is(err, user.ErrNotFound) {
			// 其他数据库错误，记录日志但不阻止发送验证码（降级处理）
			l.Logger.Errorf("查询用户邮箱时出现数据库错误: %v", err)
		}
	}

	// 3. 检查发送频率限制（1分钟内只能发送一次）
	rateLimitKey := fmt.Sprintf("email_code_rate:%s", req.Email)
	exists, _ := l.svcCtx.RedisClient.Exists(rateLimitKey)
	if exists {
		return utils.Response.BusinessError("send_to_fast"), nil
	}

	// 4. 生成6位验证码
	code := generateVerificationCode()

	// 5. 存储验证码到Redis（10分钟有效期）
	codeKey := fmt.Sprintf("email_code:%s:%s", req.Type, req.Email)
	_, err = l.svcCtx.RedisClient.SetnxEx(codeKey, code, 600)
	if err != nil {
		l.Logger.Errorf("存储验证码失败: %v", err)
		return nil, errors.New("发送验证码失败")
	}

	// 6. 设置发送频率限制（60秒）
	l.svcCtx.RedisClient.SetnxEx(rateLimitKey, "1", 60)

	// 7. 发送验证码邮件
	err = l.sendVerificationEmail(req.Email, code, req.Type)
	if err != nil {
		l.Logger.Errorf("发送验证码邮件失败: %v", err)
		// 删除已存储的验证码
		l.svcCtx.RedisClient.Del(codeKey)
		return nil, errors.New("发送验证码失败")
	}

	return utils.Response.Success(map[string]interface{}{
		"message": "验证码已发送",
	}), nil
}

// generateVerificationCode 生成6位随机数字验证码
func generateVerificationCode() string {
	// 使用独立的随机数生成器，避免并发问题
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := r.Intn(900000) + 100000 // 生成100000-999999之间的数字
	return fmt.Sprintf("%06d", code)
}

// sendVerificationEmail 发送验证码邮件（使用统一邮件模板服务）
func (l *SendVerificationCodeLogic) sendVerificationEmail(email, code, codeType string) error {
	var subject, purpose string
	var theme string

	switch codeType {
	case "register":
		subject = "注册验证码 - Task 任务管理系统"
		purpose = "注册 Task 任务管理系统账号"
		theme = "info"
	case "reset_password":
		subject = "重置密码验证码 - Task 任务管理系统"
		purpose = "重置 Task 任务管理系统登录密码"
		theme = "warning"
	default:
		return errors.New("无效的验证码类型")
	}

	return l.svcCtx.UnifiedEmailService.SendEmail(l.ctx, svc.EmailRequest{
		TemplateName: "email/auth/verification_code",
		To:           []string{email},
		Subject:      subject,
		Data: map[string]interface{}{
			"code":    code,
			"purpose": purpose,
		},
		Style: svc.TemplateStyle{
			Theme: theme,
		},
	})
}
