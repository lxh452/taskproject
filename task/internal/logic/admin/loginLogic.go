package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	adminModel "task_Project/model/admin"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

// Admin Token在Redis中的key前缀
const (
	AdminTokenKeyPrefix = "admin:token:"
	AdminTokenExpire    = 86400 // 24小时过期
)

type AdminLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理员登录
func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLoginLogic {
	return &AdminLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminLoginLogic) AdminLogin(req *types.AdminLoginRequest, r *http.Request) (resp *types.BaseResponse, err error) {
	// 参数验证
	if utils.Validator.IsEmpty(req.Username) || utils.Validator.IsEmpty(req.Password) {
		return utils.Response.ValidationError("用户名和密码不能为空"), nil
	}

	// 获取客户端IP
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.Header.Get("X-Real-IP")
	}
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	// 获取User-Agent
	userAgent := r.Header.Get("User-Agent")

	// 查找管理员
	adminInfo, err := l.svcCtx.AdminModel.FindByUsername(l.ctx, req.Username)
	if err != nil {
		// 记录登录失败
		l.recordLoginAttempt("", "admin", req.Username, clientIP, userAgent, 0, "管理员账号不存在")
		if errors.Is(err, adminModel.ErrNotFound) {
			return utils.Response.BusinessError("用户名或密码错误"), nil
		}
		logx.Errorf("查找管理员失败: %v", err)
		return utils.Response.InternalError("查找管理员失败"), nil
	}

	// 检查管理员状态
	if adminInfo.Status != 1 {
		l.recordLoginAttempt(adminInfo.Id, "admin", req.Username, clientIP, userAgent, 0, "管理员账号已禁用")
		return utils.Response.BusinessError("管理员账号已禁用"), nil
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(adminInfo.PasswordHash), []byte(req.Password))
	if err != nil {
		l.recordLoginAttempt(adminInfo.Id, "admin", req.Username, clientIP, userAgent, 0, "密码错误")
		return utils.Response.BusinessError("用户名或密码错误"), nil
	}

	// 登录成功，生成JWT令牌
	token, err := l.svcCtx.JWTMiddleware.GenerateToken(adminInfo.Id, adminInfo.Username, adminInfo.RealName.String, "admin")
	if err != nil {
		logx.Errorf("生成JWT令牌失败: %v", err)
		return utils.Response.InternalError("生成JWT令牌失败"), nil
	}

	// 将Token存储到Redis
	tokenKey := fmt.Sprintf("%s%s", AdminTokenKeyPrefix, adminInfo.Id)
	if err := l.svcCtx.RedisClient.Setex(tokenKey, token, AdminTokenExpire); err != nil {
		logx.Errorf("存储Token到Redis失败: %v", err)
	} else {
		logx.Infof("Admin Token已存储到Redis: adminId=%s, key=%s", adminInfo.Id, tokenKey)
	}

	// 更新最后登录信息
	now := time.Now()
	updateErr := l.svcCtx.AdminModel.UpdateLastLogin(l.ctx, adminInfo.Id, now.Format("2006-01-02 15:04:05"), clientIP)
	if updateErr != nil {
		logx.Errorf("更新最后登录信息失败: %v", updateErr)
	}

	// 记录登录成功
	l.recordLoginAttempt(adminInfo.Id, "admin", req.Username, clientIP, userAgent, 1, "")

	// 记录系统日志
	if l.svcCtx.SystemLogService != nil {
		l.svcCtx.SystemLogService.AdminAction(l.ctx, "auth", "login", fmt.Sprintf("管理员 %s 登录成功", req.Username), adminInfo.Id, clientIP, userAgent)
	}

	// 返回登录响应
	loginResp := types.AdminLoginResponse{
		Token:    token,
		AdminID:  adminInfo.Id,
		Username: adminInfo.Username,
		RealName: adminInfo.RealName.String,
		Role:     adminInfo.Role,
	}

	return utils.Response.SuccessWithData(loginResp), nil
}

// recordLoginAttempt 记录登录尝试
func (l *AdminLoginLogic) recordLoginAttempt(userId, userType, username, loginIP, userAgent string, status int, failReason string) {
	record := &adminModel.LoginRecord{
		Id:          utils.Common.GenerateIDWithPrefix("login_"),
		UserId:      userId,
		UserType:    userType,
		Username:    sql.NullString{String: username, Valid: username != ""},
		LoginTime:   time.Now(),
		LoginIp:     sql.NullString{String: loginIP, Valid: loginIP != ""},
		UserAgent:   sql.NullString{String: userAgent, Valid: userAgent != ""},
		LoginStatus: int64(status),
		FailReason:  sql.NullString{String: failReason, Valid: failReason != ""},
	}

	_, err := l.svcCtx.LoginRecordModel.Insert(l.ctx, record)
	if err != nil {
		logx.Errorf("记录登录日志失败: %v", err)
	}
}
