package admin

import (
	"context"
	"fmt"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminLogoutLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理员登出
func NewAdminLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLogoutLogic {
	return &AdminLogoutLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminLogoutLogic) AdminLogout(req *types.AdminLogoutRequest) (resp *types.BaseResponse, err error) {
	// 从上下文获取当前管理员ID
	adminID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok || adminID == "" {
		return utils.Response.UnauthorizedError(), nil
	}

	// 从Redis中删除Token
	tokenKey := fmt.Sprintf("%s%s", AdminTokenKeyPrefix, adminID)
	_, err = l.svcCtx.RedisClient.Del(tokenKey)
	if err != nil {
		logx.Errorf("从Redis删除Token失败: %v", err)
		// 不影响登出流程
	} else {
		logx.Infof("Admin Token已从Redis删除: adminId=%s, key=%s", adminID, tokenKey)
	}

	return utils.Response.SuccessWithKey("logout", nil), nil
}
