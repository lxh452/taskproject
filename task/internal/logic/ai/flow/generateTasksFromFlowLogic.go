// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package flow

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateTasksFromFlowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 根据流程生成任务
func NewGenerateTasksFromFlowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateTasksFromFlowLogic {
	return &GenerateTasksFromFlowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateTasksFromFlowLogic) GenerateTasksFromFlow(req *types.GenerateTasksFromFlowRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
