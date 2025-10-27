// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package position

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPositionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取职位信息
func NewGetPositionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionLogic {
	return &GetPositionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPositionLogic) GetPosition(req *types.GetPositionRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.PositionID) {
		return utils.Response.ValidationError("职位ID不能为空"), nil
	}

	// 查询职位信息
	position, err := l.svcCtx.PositionModel.FindOne(l.ctx, req.PositionID)
	if err != nil {
		logx.Errorf("查询职位失败: %v", err)
		return utils.Response.ErrorWithKey("position_not_found"), nil
	}

	// 转换为响应格式
	converter := utils.NewConverter()
	positionInfo := converter.ToPositionInfo(position)

	return utils.Response.SuccessWithKey("position", positionInfo), nil
}
