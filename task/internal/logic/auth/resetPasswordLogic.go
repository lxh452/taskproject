package auth

import (
	"context"
	"errors"
	"fmt"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type ResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetPasswordLogic) ResetPassword(req *types.ResetPasswordRequest) (resp *types.BaseResponse, err error) {
	// 1. 验证邮箱格式
	if !utils.Common.IsValidEmail(req.Email) {
		return utils.Response.BusinessError("email_format_invalid"), nil
	}

	// 2. 验证密码强度
	if !utils.Common.IsValidPassword(req.NewPassword) {
		return utils.Response.BusinessError("password_format_invalid"), nil
	}

	// 3. 验证验证码
	codeKey := fmt.Sprintf("email_code:reset_password:%s", req.Email)
	storedCode, err := l.svcCtx.RedisClient.Get(codeKey)
	if err != nil || storedCode == "" {
		return utils.Response.BusinessError("code_error"), nil
	}
	if storedCode != req.VerificationCode {
		return utils.Response.BusinessError("code_error"), nil
	}

	// 4. 查找用户
	user, err := l.svcCtx.UserModel.FindByEmail(l.ctx, req.Email)
	if err != nil || user == nil {
		return utils.Response.BusinessError("email_not_found"), nil
	}

	// 5. 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		l.Logger.Errorf("加密密码失败: %v", err)
		return nil, errors.New("系统错误")
	}

	// 6. 更新密码
	user.PasswordHash = string(hashedPassword)
	err = l.svcCtx.UserModel.Update(l.ctx, user)
	if err != nil {
		l.Logger.Errorf("更新密码失败: %v", err)
		return nil, errors.New("重置密码失败")
	}

	// 7. 删除已使用的验证码
	l.svcCtx.RedisClient.Del(codeKey)

	// 8. 清除用户的所有登录token（强制重新登录）
	tokenPattern := fmt.Sprintf("user_token:%s:*", user.Id)
	keys, _ := l.svcCtx.RedisClient.Keys(tokenPattern)
	for _, key := range keys {
		l.svcCtx.RedisClient.Del(key)
	}

	return utils.Response.Success(map[string]interface{}{
		"message": "密码重置成功，请使用新密码登录",
	}), nil
}
