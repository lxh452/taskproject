// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package department

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDepartmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取部门信息
func NewGetDepartmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDepartmentLogic {
	return &GetDepartmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDepartmentLogic) GetDepartment(req *types.GetDepartmentRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.ID) {
		return utils.Response.ValidationError("部门ID不能为空"), nil
	}

	// 查询部门信息
	department, err := l.svcCtx.DepartmentModel.FindOne(l.ctx, req.ID)
	if err != nil {
		logx.Errorf("查询部门失败: %v", err)
		return utils.Response.ErrorWithKey("department_not_found"), nil
	}

	// 转换为响应格式
	converter := utils.NewConverter()
	departmentInfo := converter.ToDepartmentInfo(department)

	return utils.Response.SuccessWithKey("department", departmentInfo), nil
}
