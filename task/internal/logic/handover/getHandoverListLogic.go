// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handover

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHandoverListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取交接列表
func NewGetHandoverListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHandoverListLogic {
	return &GetHandoverListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetHandoverListLogic) GetHandoverList(req *types.HandoverListRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
