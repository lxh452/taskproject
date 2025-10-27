// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handover

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateHandoverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建任务交接
func NewCreateHandoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateHandoverLogic {
	return &CreateHandoverLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateHandoverLogic) CreateHandover(req *types.CreateHandoverRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
