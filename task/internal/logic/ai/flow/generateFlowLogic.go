// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package flow

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateFlowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 生成流程图
func NewGenerateFlowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateFlowLogic {
	return &GenerateFlowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateFlowLogic) GenerateFlow(req *types.GenerateFlowRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
