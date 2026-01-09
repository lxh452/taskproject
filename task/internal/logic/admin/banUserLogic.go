package admin

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type BanUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 封禁用户
func NewBanUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BanUserLogic {
	return &BanUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BanUserLogic) BanUser(req *types.BanUserRequest) (resp *types.BaseResponse, err error) {
	// 验证用户是否存在
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, req.UserID)
	if err != nil {
		logx.Errorf("查询用户失败: %v", err)
		return utils.Response.Error(404, "用户不存在"), nil
	}

	// 检查用户是否已被封禁 (status = 2 表示封禁)
	if user.Status == 2 {
		return utils.Response.Error(400, "用户已被封禁"), nil
	}

	// 封禁用户 (status = 2)
	err = l.svcCtx.UserModel.UpdateStatus(l.ctx, req.UserID, 2)
	if err != nil {
		logx.Errorf("封禁用户失败: %v", err)
		return utils.Response.Error(500, "封禁用户失败"), nil
	}

	// 清除用户的登录Token（强制下线）
	tokenKey := "user:token:" + req.UserID
	_, _ = l.svcCtx.RedisClient.Del(tokenKey)

	logx.Infof("用户 %s 已被封禁，原因: %s", req.UserID, req.BanReason)

	// 记录系统日志
	if l.svcCtx.SystemLogService != nil {
		l.svcCtx.SystemLogService.AdminAction(l.ctx, "user", "ban", "封禁用户: "+req.UserID+", 原因: "+req.BanReason, "", "", "")
	}

	return utils.Response.Success("用户已封禁"), nil
}
