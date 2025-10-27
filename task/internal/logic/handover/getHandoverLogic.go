// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handover

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHandoverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取交接信息
func NewGetHandoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHandoverLogic {
	return &GetHandoverLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHandoverLogic) GetHandover(req *types.GetHandoverRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
