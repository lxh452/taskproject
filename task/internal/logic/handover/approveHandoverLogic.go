// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handover

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApproveHandoverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 审批任务交接
func NewApproveHandoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveHandoverLogic {
	return &ApproveHandoverLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApproveHandoverLogic) ApproveHandover(req *types.ApproveHandoverRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
