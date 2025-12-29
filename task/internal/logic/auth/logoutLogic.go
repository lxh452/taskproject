// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"
	"fmt"
	"time"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户登出
func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogoutLogic) Logout() (resp *types.BaseResponse, err error) {
	// 从上下文中获取用户信息

	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	username, _ := utils.Common.GetCurrentUsername(l.ctx)
	realName, _ := utils.Common.GetCurrentRealName(l.ctx)

	// 从Redis删除Token，使Token失效
	tokenKey := fmt.Sprintf("%s%s", TokenKeyPrefix, userID)
	if _, err := l.svcCtx.RedisClient.Del(tokenKey); err != nil {
		logx.Errorf("从Redis删除Token失败: %v", err)
		// 不影响登出流程
	} else {
		logx.Infof("Token已从Redis删除: userId=%s, key=%s", userID, tokenKey)
	}

	// 记录登出日志
	logx.Infof("用户登出: userID=%s, username=%s, realName=%s", userID, username, realName)

	// 发送登出通知邮件（如果有邮箱）
	// 这里需要从数据库获取用户邮箱信息
	go func() {
		// 可以在这里添加发送登出通知的逻辑
		// 比如发送邮件通知用户账户已登出
		logx.Infof("用户 %s 已登出系统", username)
	}()

	return utils.Response.SuccessWithKey("logout", map[string]interface{}{
		"logoutTime": utils.Common.FormatTime(time.Now()),
		"username":   username,
	}), nil
}
