// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package position

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPositionListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取职位列表
func NewGetPositionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPositionListLogic {
	return &GetPositionListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPositionListLogic) GetPositionList(req *types.PositionListRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if _, _, errors := validator.ValidatePageParams(req.Page, req.PageSize); err != nil {
		return utils.Response.ValidationError(errors[0]), nil
	}

	// 设置默认分页参数
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	// 查询职位列表
	positions, err := l.svcCtx.PositionModel.FindByPage(l.ctx, page, pageSize, req.DepartmentID, req.Status)
	if err != nil {
		logx.Errorf("查询职位列表失败: %v", err)
		return utils.Response.InternalError("查询职位列表失败"), err
	}

	// 获取总数
	total, err := l.svcCtx.PositionModel.GetPositionCount(l.ctx, req.DepartmentID, req.Status)
	if err != nil {
		logx.Errorf("查询职位总数失败: %v", err)
		return utils.Response.InternalError("查询职位总数失败"), err
	}

	// 转换为响应格式
	converter := utils.NewConverter()
	positionList := converter.ToPositionInfoList(positions)

	// 构建分页响应
	pageResp := utils.NewConverter().ToPageResponse(positionList, total, page, pageSize)

	return utils.Response.SuccessWithKey("positions", pageResp), nil
}
