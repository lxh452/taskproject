// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handover

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmHandoverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 确认交接
func NewConfirmHandoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmHandoverLogic {
	return &ConfirmHandoverLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmHandoverLogic) ConfirmHandover(req *types.ConfirmHandoverRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
