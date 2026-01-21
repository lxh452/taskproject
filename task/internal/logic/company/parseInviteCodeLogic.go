package company

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type ParseInviteCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 解析邀请码
func NewParseInviteCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ParseInviteCodeLogic {
	return &ParseInviteCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ParseInviteCodeLogic) ParseInviteCode(req *types.ParseInviteCodeRequest) (resp *types.BaseResponse, err error) {
	// 解析邀请码
	data, err := l.svcCtx.InviteCodeService.ParseInviteCode(l.ctx, req.InviteCode)
	if err != nil {
		logx.Errorf("解析邀请码失败: code=%s, err=%v", req.InviteCode, err)
		return utils.Response.BusinessError(err.Error()), nil
	}

	return utils.Response.Success(map[string]interface{}{
		"companyId":   data.CompanyID,
		"companyName": data.CompanyName,
		"maxUses":     data.MaxUses,
		"usedCount":   data.UsedCount,
	}), nil
}
