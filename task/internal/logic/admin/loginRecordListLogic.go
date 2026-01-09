package admin

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginRecordListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取登录记录列表
func NewLoginRecordListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginRecordListLogic {
	return &LoginRecordListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginRecordListLogic) LoginRecordList(req *types.LoginRecordListRequest) (resp *types.BaseResponse, err error) {
	// 设置默认分页参数
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 处理状态参数，-1表示不筛选
	status := req.LoginStatus
	if status == 0 && req.LoginStatus != 0 {
		status = -1 // 不筛选状态
	}

	// 使用 FindByFilters 在数据库层面进行筛选
	records, total, err := l.svcCtx.LoginRecordModel.FindByFilters(
		l.ctx,
		req.UserID,
		req.UserType,
		status,
		req.StartTime,
		req.EndTime,
		page,
		pageSize,
	)
	if err != nil {
		logx.Errorf("查询登录记录失败: %v", err)
		return utils.Response.Error(500, "查询登录记录失败"), nil
	}

	var recordList []types.LoginRecordInfo
	for _, r := range records {
		recordInfo := types.LoginRecordInfo{
			ID:          r.Id,
			UserID:      r.UserId,
			Username:    r.Username.String,
			UserType:    r.UserType,
			LoginIP:     r.LoginIp.String,
			UserAgent:   r.UserAgent.String,
			LoginStatus: int(r.LoginStatus),
			FailReason:  r.FailReason.String,
			LoginTime:   r.LoginTime.Format("2006-01-02 15:04:05"),
			CreateTime:  r.CreateTime.Format("2006-01-02 15:04:05"),
		}
		recordList = append(recordList, recordInfo)
	}

	return utils.Response.SuccessWithData(map[string]interface{}{
		"list":     recordList,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}), nil
}
