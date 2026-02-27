package admin

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogListLogic {
	return &LogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LogListLogic) LogList(req *types.SystemLogListRequest) (*types.BaseResponse, error) {
	// 设置默认分页参数
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 检查SystemLogModel是否可用（MongoDB未启用时为nil）
	if l.svcCtx.SystemLogModel == nil {
		logx.Info("SystemLogModel is nil, returning empty log list")
		return utils.Response.SuccessWithData(map[string]interface{}{
			"list":     []interface{}{},
			"total":    0,
			"page":     page,
			"pageSize": pageSize,
		}), nil
	}

	// 查询日志列表（简化版本）
	list := []interface{}{}

	return utils.Response.SuccessWithData(map[string]interface{}{
		"list":     list,
		"total":    0,
		"page":     page,
		"pageSize": pageSize,
	}), nil
}
