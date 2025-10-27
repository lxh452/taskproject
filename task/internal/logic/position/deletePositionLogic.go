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

type DeletePositionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除职位
func NewDeletePositionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePositionLogic {
	return &DeletePositionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePositionLogic) DeletePosition(req *types.DeletePositionRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.PositionID) {
		return utils.Response.ValidationError("职位ID不能为空"), nil
	}

	// 检查职位是否存在
	if _, err = l.svcCtx.PositionModel.FindOne(l.ctx, req.PositionID); err != nil {
		logx.Errorf("查询职位失败: %v", err)
		return utils.Response.ErrorWithKey("position_not_found"), nil
	}

	// 检查职位是否有员工
	employeeCount, err := l.svcCtx.EmployeeModel.GetEmployeeCountByPosition(l.ctx, req.PositionID)
	if err != nil {
		logx.Errorf("查询职位员工数量失败: %v", err)
		return utils.Response.InternalError("查询职位员工数量失败"), err
	}

	if employeeCount > 0 {
		return utils.Response.BusinessError("职位还有员工，无法删除"), nil
	}

	// 软删除职位
	err = l.svcCtx.PositionModel.SoftDelete(l.ctx, req.PositionID)
	if err != nil {
		logx.Errorf("删除职位失败: %v", err)
		return utils.Response.InternalError("删除职位失败"), err
	}

	return utils.Response.Success("删除职位成功"), nil
}
