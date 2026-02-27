// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package task

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SuggestAssigneeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 推荐负责人
func NewSuggestAssigneeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SuggestAssigneeLogic {
	return &SuggestAssigneeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SuggestAssigneeLogic) SuggestAssignee(req *types.SuggestAssigneeRequest) (resp *types.BaseResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
