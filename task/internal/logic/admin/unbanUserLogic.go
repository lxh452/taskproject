package admin

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnbanUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 解封用户
func NewUnbanUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnbanUserLogic {
	return &UnbanUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnbanUserLogic) UnbanUser(req *types.UnbanUserRequest) (resp *types.BaseResponse, err error) {
	// 验证用户是否存在
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, req.UserID)
	if err != nil {
		logx.Errorf("查询用户失败: %v", err)
		return utils.Response.Error(404, "用户不存在"), nil
	}

	// 检查用户是否处于封禁状态 (status = 2 表示封禁)
	if user.Status != 2 {
		return utils.Response.Error(400, "用户未被封禁"), nil
	}

	// 解封用户 (status = 1 表示正常)
	err = l.svcCtx.UserModel.UpdateStatus(l.ctx, req.UserID, 1)
	if err != nil {
		logx.Errorf("解封用户失败: %v", err)
		return utils.Response.Error(500, "解封用户失败"), nil
	}

	logx.Infof("用户 %s 已被解封", req.UserID)

	// 记录系统日志
	if l.svcCtx.SystemLogService != nil {
		l.svcCtx.SystemLogService.AdminAction(l.ctx, "user", "unban", "解封用户: "+req.UserID, "", "", "")
	}

	return utils.Response.Success("用户已解封"), nil
}
