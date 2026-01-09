package admin

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type RestoreUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 恢复已删除用户
func NewRestoreUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RestoreUserLogic {
	return &RestoreUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RestoreUserLogic) RestoreUser(req *types.RestoreUserRequest) (resp *types.BaseResponse, err error) {
	// 直接尝试恢复用户（清除delete_time）
	// 注意：这里需要在UserModel中添加RestoreUser方法
	err = l.svcCtx.UserModel.RestoreUser(l.ctx, req.UserID)
	if err != nil {
		logx.Errorf("恢复用户失败: %v", err)
		return utils.Response.Error(500, "恢复用户失败"), nil
	}

	logx.Infof("用户 %s 已被恢复", req.UserID)

	// 记录系统日志
	if l.svcCtx.SystemLogService != nil {
		l.svcCtx.SystemLogService.AdminAction(l.ctx, "user", "restore", "恢复用户: "+req.UserID, "", "", "")
	}

	return utils.Response.Success("用户已恢复"), nil
}
