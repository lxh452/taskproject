// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AutoDispatchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 任务自动派发
func NewAutoDispatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AutoDispatchLogic {
	return &AutoDispatchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AutoDispatchLogic) AutoDispatch(req *types.AutoDispatchRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
