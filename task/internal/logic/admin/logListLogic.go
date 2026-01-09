package admin

import (
	"context"
	"time"

	adminModel "task_Project/model/admin"
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

	// 构建过滤条件
	filter := adminModel.SystemLogFilter{
		Level:    req.Level,
		Module:   req.Module,
		Keyword:  req.Keyword,
		UserID:   req.UserID,
		UserType: req.UserType, // 用户类型筛选
	}

	// 如果没有指定用户类型，默认显示用户端日志
	if filter.UserType == "" {
		filter.UserType = "user"
	}

	// 解析时间范围
	if req.StartTime != "" {
		startTime, err := time.Parse("2006-01-02 15:04:05", req.StartTime)
		if err == nil {
			filter.StartTime = startTime
		}
	}
	if req.EndTime != "" {
		endTime, err := time.Parse("2006-01-02 15:04:05", req.EndTime)
		if err == nil {
			filter.EndTime = endTime
		}
	}

	// 查询日志列表
	logs, total, err := l.svcCtx.SystemLogModel.FindList(l.ctx, filter, int64(page), int64(pageSize))
	if err != nil {
		l.Errorf("查询系统日志失败: %v", err)
		return utils.Response.Error(500, "获取日志列表失败"), nil
	}

	// 转换为响应格式
	list := make([]types.SystemLogInfo, 0, len(logs))
	for _, log := range logs {
		list = append(list, types.SystemLogInfo{
			LogID:      log.LogID,
			Level:      log.Level,
			Module:     log.Module,
			Action:     log.Action,
			Message:    log.Message,
			Detail:     log.Detail,
			UserID:     log.UserID,
			UserType:   log.UserType,
			IP:         log.IP,
			UserAgent:  log.UserAgent,
			RequestID:  log.RequestID,
			TraceID:    log.TraceID,
			StackTrace: log.StackTrace,
			CreateTime: log.CreateAt.Format("2006-01-02 15:04:05"),
		})
	}

	return utils.Response.SuccessWithData(map[string]interface{}{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}), nil
}
