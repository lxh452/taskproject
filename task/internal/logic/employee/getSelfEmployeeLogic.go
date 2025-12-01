package employee

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSelfEmployeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSelfEmployeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSelfEmployeeLogic {
	return &GetSelfEmployeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSelfEmployeeLogic) GetSelfEmployee() (resp *types.BaseResponse, err error) {
	// 从 JWT 上下文获取 userId
	userId, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok || userId == "" {
		return utils.Response.UnauthorizedError(), nil
	}

	// 通过 userId 查找员工信息
	employee, err := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, userId)
	if err != nil {
		logx.Errorf("查找员工失败: %v", err)
		return utils.Response.BusinessError("employee_not_found"), nil
	}
	if employee == nil {
		return utils.Response.BusinessError("employee_not_found"), nil
	}

	// 转换为响应格式
	converter := utils.NewConverter()
	employeeInfo := converter.ToEmployeeInfo(employee)

	return utils.Response.SuccessWithKey("employee", employeeInfo), nil
}
