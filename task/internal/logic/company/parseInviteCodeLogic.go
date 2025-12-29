package company

import (
	"context"
	"time"

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
	// 验证参数
	if utils.Validator.IsEmpty(req.InviteCode) {
		return utils.Response.ValidationError("邀请码不能为空"), nil
	}

	// 解析邀请码
	data, err := l.svcCtx.InviteCodeService.ParseInviteCode(l.ctx, req.InviteCode)
	if err != nil {
		logx.Errorf("解析邀请码失败: code=%s, err=%v", req.InviteCode, err)
		return utils.Response.BusinessError(err.Error()), nil
	}

	// 验证公司是否存在
	company, err := l.svcCtx.CompanyModel.FindOne(l.ctx, data.CompanyID)
	if err != nil {
		logx.Errorf("查询公司失败: companyId=%s, err=%v", data.CompanyID, err)
		return utils.Response.BusinessError("邀请码对应的公司不存在"), nil
	}

	// 检查公司状态
	if company.Status != 1 {
		return utils.Response.BusinessError("该公司已停用"), nil
	}

	// 返回公司信息
	expireTime := time.Unix(data.ExpireAt, 0).Format("2006-01-02 15:04:05")

	return utils.Response.Success(map[string]interface{}{
		"companyId":   data.CompanyID,
		"companyName": company.Name,
		"description": company.Description.String,
		"address":     company.Address.String,
		"expireAt":    expireTime,
		"maxUses":     data.MaxUses,
		"usedCount":   data.UsedCount,
	}), nil
}
