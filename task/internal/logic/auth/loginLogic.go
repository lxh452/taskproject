// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"task_Project/model/user"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

// Token在Redis中的key前缀
const (
	TokenKeyPrefix = "auth:token:"
	TokenExpire    = 86400 // 24小时过期
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	if utils.Validator.IsEmpty(req.Username) || utils.Validator.IsEmpty(req.Password) {
		return utils.Response.ValidationError("用户名和密码不能为空"), nil
	}

	// 查找用户
	userInfo, err := l.svcCtx.UserModel.FindByUsername(l.ctx, req.Username)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return utils.Response.BusinessError("login_failed"), nil
		}
		logx.Errorf("查找用户失败: %v", err)
		return utils.Response.InternalError("查找用户失败"), nil
	}

	// 检查用户状态
	if userInfo.Status != 1 {
		return utils.Response.BusinessError("user_disabled"), nil
	}

	// 检查用户是否被锁定
	if userInfo.LockedUntil.Valid && userInfo.LockedUntil.Time.After(time.Now()) {
		return utils.Response.BusinessError("user_locked"), nil
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(userInfo.PasswordHash), []byte(req.Password))
	if err != nil {
		// 密码错误，增加失败次数
		failedCount := userInfo.LoginFailedCount + 1
		updateErr := l.svcCtx.UserModel.UpdateLoginFailedCount(l.ctx, userInfo.Id, int(failedCount))
		if updateErr != nil {
			logx.Errorf("更新登录失败次数失败: %v", updateErr)
		}

		// 如果失败次数达到5次，锁定用户1小时
		if failedCount >= 5 {
			lockUntil := time.Now().Add(time.Hour)
			lockErr := l.svcCtx.UserModel.UpdateLockStatus(l.ctx, userInfo.Id, lockUntil.Format("2006-01-02 15:04:05"))
			if lockErr != nil {
				logx.Errorf("锁定用户失败: %v", lockErr)
			}
			return utils.Response.BusinessError("login_failed_too_many"), nil
		}

		return utils.Response.BusinessError("login_failed"), nil
	}

	// 如果用户已加入公司，查询员工ID
	var employeeID, companyID string
	if userInfo.HasJoinedCompany == 1 {
		employee, err := l.svcCtx.EmployeeModel.FindOneByUserId(l.ctx, userInfo.Id)
		if err == nil && employee != nil {
			employeeID = employee.Id
			companyID = employee.CompanyId
		}
	}
	fmt.Println("员工id", employeeID)

	// 登录成功，生成JWT令牌（包含员工信息）
	token, err := l.svcCtx.JWTMiddleware.GenerateTokenWithEmployee(userInfo.Id, userInfo.Username, userInfo.RealName.String, "user", employeeID, companyID)
	if err != nil {
		logx.Errorf("生成JWT令牌失败: %v", err)
		return utils.Response.InternalError("生成JWT令牌失败"), nil
	}

	// 将Token存储到Redis，用于后续验证
	tokenKey := fmt.Sprintf("%s%s", TokenKeyPrefix, userInfo.Id)
	if err := l.svcCtx.RedisClient.Setex(tokenKey, token, TokenExpire); err != nil {
		logx.Errorf("存储Token到Redis失败: %v", err)
		// 不影响登录流程，但记录错误
	} else {
		logx.Infof("Token已存储到Redis: userId=%s, key=%s", userInfo.Id, tokenKey)
	}

	// 更新最后登录信息
	now := time.Now()
	updateErr := l.svcCtx.UserModel.UpdateLastLogin(l.ctx, userInfo.Id, now.Format("2006-01-02 15:04:05"), "127.0.0.1")
	if updateErr != nil {
		logx.Errorf("更新最后登录信息失败: %v", updateErr)
	}

	// 发送登录成功通知邮件（通过消息队列）
	go func() {
		if userInfo.Email.Valid && userInfo.Email.String != "" && l.svcCtx.EmailService != nil {
			loginTime := now.Format("2006-01-02 15:04:05")
			loginIP := "127.0.0.1" // TODO: 从请求中获取真实IP
			if err := l.svcCtx.EmailService.SendLoginSuccessEmail(context.Background(), userInfo.Email.String, userInfo.Username, loginTime, loginIP); err != nil {
				logx.Errorf("发送登录通知邮件失败: %v", err)
			}
		}
	}()

	// 返回登录响应
	loginResp := types.LoginResponse{
		Token:            token,
		UserID:           userInfo.Id,
		Username:         userInfo.Username,
		RealName:         userInfo.RealName.String,
		HasJoinedCompany: userInfo.HasJoinedCompany == 1,
	}

	return utils.Response.SuccessWithKey("login", loginResp), nil
}
