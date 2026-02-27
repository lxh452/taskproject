// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package aiagent

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 执行 Agent
func NewExecuteAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteAgentLogic {
	return &ExecuteAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExecuteAgentLogic) ExecuteAgent(req *types.ExecuteAgentRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
