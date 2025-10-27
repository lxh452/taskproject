package employee

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetEmployeeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取员工列表
func NewGetEmployeeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetEmployeeListLogic {
	return &GetEmployeeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetEmployeeListLogic) GetEmployeeList(req *types.EmployeeListRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	var errors []string
	req.Page, req.PageSize, errors = validator.ValidatePageParams(req.Page, req.PageSize)
	if len(errors) > 0 {
		return utils.Response.ValidationError(errors[0]), nil
	}

	// 获取当前用户信息
	if _, ok := utils.Common.GetCurrentUserID(l.ctx); !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 查询员工列表
	employees, total, err := l.svcCtx.EmployeeModel.FindByPage(l.ctx, req.Page, req.PageSize)
	if err != nil {
		logx.Errorf("查询员工列表失败: %v", err)
		return utils.Response.InternalError("查询员工列表失败"), err
	}

	// 转换为响应格式
	converter := utils.NewConverter()
	employeeList := converter.ToEmployeeInfoList(employees)

	// 构建分页响应
	pageResp := converter.ToPageResponse(employeeList, int(total), req.Page, req.PageSize)

	return utils.Response.SuccessWithKey("employees", pageResp), nil
}
