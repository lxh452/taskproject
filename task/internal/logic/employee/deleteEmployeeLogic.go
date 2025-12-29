// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package employee

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteEmployeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除员工
func NewDeleteEmployeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteEmployeeLogic {
	return &DeleteEmployeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 这里更换为hr主动提出员工离职 todo
func (l *DeleteEmployeeLogic) DeleteEmployee(req *types.DeleteEmployeeRequest) (resp *types.BaseResponse, err error) {

	// todo 既然这里是hr主动删除员工，就是离职的所以这里要发送邮件和短信，同时这个事情不能毫无准备，现在在这一块需要提供定时任务，发送到mq
	// 参数验证
	validator := utils.NewValidator()
	if validator.IsEmpty(req.EmployeeID) {
		return utils.Response.ValidationError("员工ID不能为空"), nil
	}

	// 检查员工是否存在
	if _, err = l.svcCtx.EmployeeModel.FindOne(l.ctx, req.EmployeeID); err != nil {
		logx.Errorf("查询员工失败: %v", err)
		return utils.Response.ErrorWithKey("employee_not_found"), nil
	}

	// 检查员工是否有未完成的任务
	// todo 这里可能需要重新处理找到的逻辑 目前这一块只是找了任务创建是这个id的情况 可能这人一辈子都是员工
	taskCount, err := l.svcCtx.TaskModel.GetTaskCountByCreator(l.ctx, req.EmployeeID)
	if err != nil {
		logx.Errorf("查询员工任务数量失败: %v", err)
		return utils.Response.InternalError("查询员工任务数量失败"), err
	}

	if taskCount > 0 {
		return utils.Response.BusinessError("员工还有未完成的任务，无法删除"), nil
	}

	// 软删除员工
	err = l.svcCtx.EmployeeModel.SoftDelete(l.ctx, req.EmployeeID)
	if err != nil {
		logx.Errorf("删除员工失败: %v", err)
		return utils.Response.InternalError("删除员工失败"), err
	}

	return utils.Response.Success("删除员工成功"), nil
}
