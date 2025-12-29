// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package position

import (
	"context"

	"task_Project/model/company"
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
	var errors []string
	if req.Page, req.PageSize, errors = validator.ValidatePageParams(req.Page, req.PageSize); len(errors) > 0 {
		return utils.Response.ValidationError(errors[0]), nil
	}

	var positions []*company.Position
	var total int64

	// 根据条件查询职位列表
	if req.DepartmentID != "" && req.Name != "" {
		// 按部门ID和名称搜索
		positions, total, err = l.svcCtx.PositionModel.SearchPositionsByDepartment(l.ctx, req.DepartmentID, req.Name, req.Page, req.PageSize)
	} else if req.DepartmentID != "" {
		// 按部门ID查询
		positions, total, err = l.svcCtx.PositionModel.FindByDepartmentPage(l.ctx, req.DepartmentID, req.Page, req.PageSize)
	} else if req.Name != "" {
		// 按名称搜索
		positions, total, err = l.svcCtx.PositionModel.SearchPositions(l.ctx, req.Name, req.Page, req.PageSize)
	} else {
		// 查询所有职位列表
		positions, total, err = l.svcCtx.PositionModel.FindByPage(l.ctx, req.Page, req.PageSize)
	}

	if err != nil {
		logx.Errorf("查询职位列表失败: %v", err)
		return utils.Response.InternalError("查询职位列表失败"), err
	}

	// 转换为响应格式
	converter := utils.NewConverter()
	positionList := converter.ToPositionInfoList(positions)

	// 构建分页响应
	pageResp := utils.NewConverter().ToPageResponse(positionList, int(total), req.Page, req.PageSize)

	return utils.Response.SuccessWithKey("list", pageResp), nil
}
