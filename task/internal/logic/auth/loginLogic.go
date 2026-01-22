// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	adminModel "task_Project/model/admin"
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
			// 记录登录失败日志（用户不存在）
			l.recordLoginLog("", req.Username, 0, "用户不存在")
			return utils.Response.BusinessError("login_failed"), nil
		}
		logx.Errorf("查找用户失败: %v", err)
		return utils.Response.InternalError("查找用户失败"), nil
	}

	// 检查用户是否被封禁 (status = 2 表示封禁)
	if userInfo.Status == 2 {
		// 记录登录失败日志（用户被封禁）
		l.recordLoginLog(userInfo.Id, req.Username, 0, "用户已被封禁")
		return utils.Response.BusinessError("user_banned"), nil
	}

	// 检查用户状态
	if userInfo.Status != 1 {
		// 记录登录失败日志（用户被禁用）
		l.recordLoginLog(userInfo.Id, req.Username, 0, "用户已被禁用")
		return utils.Response.BusinessError("user_disabled"), nil
	}

	// 检查用户是否被锁定
	if userInfo.LockedUntil.Valid && userInfo.LockedUntil.Time.After(time.Now()) {
		// 计算剩余锁定时间
		remainingMinutes := int(userInfo.LockedUntil.Time.Sub(time.Now()).Minutes()) + 1
		logx.Infof("用户 %s 仍处于锁定状态，剩余 %d 分钟", userInfo.Username, remainingMinutes)
		return utils.Response.BusinessErrorWithNum(fmt.Sprintf("账户已锁定，请在 %d 分钟后重试", remainingMinutes)), nil
	}

	// 如果锁定时间已过期，自动解锁（重置失败次数和锁定状态）
	if userInfo.LockedUntil.Valid && userInfo.LockedUntil.Time.Before(time.Now()) {
		logx.Infof("用户 %s 锁定已过期，自动解锁", userInfo.Username)
		// 重置登录失败次数
		if err := l.svcCtx.UserModel.UpdateLoginFailedCount(l.ctx, userInfo.Id, 0); err != nil {
			logx.Errorf("重置登录失败次数失败: %v", err)
		}
		// 清除锁定状态
		if err := l.svcCtx.UserModel.ClearLockStatus(l.ctx, userInfo.Id); err != nil {
			logx.Errorf("清除锁定状态失败: %v", err)
		}
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

		// 如果失败次数达到5次，锁定用户10分钟
		if failedCount >= 5 {
			lockUntil := time.Now().Add(10 * time.Minute)
			lockErr := l.svcCtx.UserModel.UpdateLockStatus(l.ctx, userInfo.Id, lockUntil.Format("2006-01-02 15:04:05"))
			if lockErr != nil {
				logx.Errorf("锁定用户失败: %v", lockErr)
			}
			logx.Infof("用户 %s 登录失败次数达到5次，锁定10分钟", userInfo.Username)
			// 记录登录失败日志（账户锁定）
			l.recordLoginLog(userInfo.Id, req.Username, 0, "登录失败次数过多，账户已锁定")
			return utils.Response.BusinessError("login_too_many_attempts"), nil
		}

		remainingAttempts := 5 - failedCount
		// 记录登录失败日志（密码错误）
		l.recordLoginLog(userInfo.Id, req.Username, 0, fmt.Sprintf("密码错误，剩余%d次尝试", remainingAttempts))
		return utils.Response.BusinessErrorWithNum(fmt.Sprintf("用户名或密码错误，还剩 %d 次尝试机会", remainingAttempts)), nil
	}

	// 如果用户已加入公司，查询员工ID并检查员工状态
	var employeeID, companyID string
	if userInfo.HasJoinedCompany == 1 {
		employee, err := l.svcCtx.EmployeeModel.FindOneByUserId(l.ctx, userInfo.Id)
		if err == nil && employee != nil {
			// 检查员工是否已离职（status = 0）
			if employee.Status == 0 {
				// 记录登录失败日志（员工已离职）
				l.recordLoginLog(userInfo.Id, req.Username, 0, "员工已离职，无法登录")
				return utils.Response.BusinessError("employee_left"), nil
			}
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

	// 记录登录成功日志
	l.recordLoginLog(userInfo.Id, req.Username, 1, "登录成功")

	// 记录系统日志
	if l.svcCtx.SystemLogService != nil {
		l.svcCtx.SystemLogService.UserAction(l.ctx, "auth", "login", fmt.Sprintf("用户 %s 登录成功", req.Username), userInfo.Id, "127.0.0.1", "")
	}

	// 发送登录成功通知邮件（通过消息队列）
	go func() {
		if userInfo.Email.Valid && userInfo.Email.String != "" && l.svcCtx.EmailService != nil {
			loginTime := now.Format("2006-01-02 15:04:05")
			loginIP := "127.0.0.1"
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

// recordLoginLog 记录登录日志
func (l *LoginLogic) recordLoginLog(userID, username string, status int64, message string) {
	go func() {
		record := &adminModel.LoginRecord{
			Id:          utils.Common.GenId("lr"),
			UserId:      userID,
			UserType:    "user",
			Username:    utils.Common.ToSqlNullString(username),
			LoginTime:   time.Now(),
			LoginIp:     utils.Common.ToSqlNullString("127.0.0.1"), // TODO: 从请求中获取真实IP
			UserAgent:   utils.Common.ToSqlNullString(""),          // TODO: 从请求中获取User-Agent
			LoginStatus: status,
			FailReason:  utils.Common.ToSqlNullString(message),
		}
		_, err := l.svcCtx.LoginRecordModel.Insert(context.Background(), record)
		if err != nil {
			logx.Errorf("记录登录日志失败: %v", err)
		}
	}()
}
