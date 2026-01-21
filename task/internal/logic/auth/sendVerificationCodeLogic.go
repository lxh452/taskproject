package auth

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

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
		if err == nil && existingUser != nil {
			return utils.Response.BusinessError("email_exists"), nil
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

	// 5. 存储验证码到Redis（5分钟有效期）
	codeKey := fmt.Sprintf("email_code:%s:%s", req.Type, req.Email)
	_, err = l.svcCtx.RedisClient.SetnxEx(codeKey, code, 300)
	if err != nil {
		l.Logger.Errorf("存储验证码失败: %v", err)
		return nil, errors.New("发送验证码失败")
	}

	// 6. 设置发送频率限制（60秒）
	l.svcCtx.RedisClient.SetnxEx(rateLimitKey, "1", 15)

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
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(900000) + 100000 // 生成100000-999999之间的数字
	return fmt.Sprintf("%06d", code)
}

// sendVerificationEmail 发送验证码邮件
func (l *SendVerificationCodeLogic) sendVerificationEmail(email, code, codeType string) error {
	var subject, body string

	switch codeType {
	case "register":
		subject = "注册验证码 - Task 任务管理系统"
		body = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; background: #f5f5f5; padding: 20px; }
        .container { max-width: 500px; margin: 0 auto; background: white; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); overflow: hidden; }
        .header { background: linear-gradient(135deg, #dc2626, #b91c1c); padding: 30px; text-align: center; }
        .header h1 { color: white; margin: 0; font-size: 24px; }
        .content { padding: 30px; }
        .code { font-size: 36px; font-weight: bold; color: #dc2626; letter-spacing: 8px; text-align: center; padding: 20px; background: #fef2f2; border-radius: 8px; margin: 20px 0; }
        .notice { color: #666; font-size: 14px; margin-top: 20px; }
        .footer { background: #f9fafb; padding: 20px; text-align: center; color: #9ca3af; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📋 Task 任务管理系统</h1>
        </div>
        <div class="content">
            <p>您好！</p>
            <p>您正在注册 Task 任务管理系统账号，验证码为：</p>
            <div class="code">%s</div>
            <p class="notice">⏰ 验证码有效期为 5 分钟，请尽快完成注册。</p>
            <p class="notice">🔒 如非本人操作，请忽略此邮件。</p>
        </div>
        <div class="footer">
            © %d Task 任务管理系统
        </div>
    </div>
</body>
</html>
`, code, time.Now().Year())

	case "reset_password":
		subject = "重置密码验证码 - Task 任务管理系统"
		body = fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; background: #f5f5f5; padding: 20px; }
        .container { max-width: 500px; margin: 0 auto; background: white; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); overflow: hidden; }
        .header { background: linear-gradient(135deg, #dc2626, #b91c1c); padding: 30px; text-align: center; }
        .header h1 { color: white; margin: 0; font-size: 24px; }
        .content { padding: 30px; }
        .code { font-size: 36px; font-weight: bold; color: #dc2626; letter-spacing: 8px; text-align: center; padding: 20px; background: #fef2f2; border-radius: 8px; margin: 20px 0; }
        .notice { color: #666; font-size: 14px; margin-top: 20px; }
        .footer { background: #f9fafb; padding: 20px; text-align: center; color: #9ca3af; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔑 重置密码</h1>
        </div>
        <div class="content">
            <p>您好！</p>
            <p>您正在重置 Task 任务管理系统的登录密码，验证码为：</p>
            <div class="code">%s</div>
            <p class="notice">⏰ 验证码有效期为 5 分钟。</p>
            <p class="notice">🔒 如非本人操作，请立即检查账号安全。</p>
        </div>
        <div class="footer">
            © %d Task 任务管理系统
        </div>
    </div>
</body>
</html>
`, code, time.Now().Year())

	default:
		return errors.New("无效的验证码类型")
	}

	// 使用邮件服务发送
	if l.svcCtx.EmailService != nil {
		return l.svcCtx.EmailService.SendCustomEmail(l.ctx, email, subject, body)
	}

	return errors.New("邮件服务未配置")
}
