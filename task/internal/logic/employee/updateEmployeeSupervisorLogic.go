package employee

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateEmployeeSupervisorLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新员工直属上级
func NewUpdateEmployeeSupervisorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateEmployeeSupervisorLogic {
	return &UpdateEmployeeSupervisorLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateEmployeeSupervisorLogic) UpdateEmployeeSupervisor(req *types.UpdateEmployeeSupervisorRequest) (resp *types.BaseResponse, err error) {
	// 1. 参数验证
	if req.EmployeeID == "" {
		return utils.Response.ValidationError("员工ID不能为空"), nil
	}

	// 2. 验证员工是否存在
	employee, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, req.EmployeeID)
	if err != nil {
		logx.Errorf("查询员工失败: %v", err)
		return utils.Response.ErrorWithKey("employee_not_found"), nil
	}

	// 3. 如果设置了上级，验证上级是否存在且在职
	if req.SupervisorID != "" {
		// 不能设置自己为自己的上级
		if req.SupervisorID == req.EmployeeID {
			return utils.Response.ValidationError("不能设置自己为自己的上级"), nil
		}

		supervisor, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, req.SupervisorID)
		if err != nil {
			logx.Errorf("查询上级员工失败: %v", err)
			return utils.Response.ValidationError("指定的上级不存在"), nil
		}

		// 上级必须在职
		if supervisor.Status != 1 {
			return utils.Response.ValidationError("指定的上级已离职，无法设置"), nil
		}

		// 上级必须在同一公司
		if supervisor.CompanyId != employee.CompanyId {
			return utils.Response.ValidationError("上级必须在同一公司"), nil
		}

		// 检查是否会形成循环（A的上级是B，B的上级是A）
		if l.wouldCreateCycle(req.EmployeeID, req.SupervisorID) {
			return utils.Response.ValidationError("设置失败：会形成上下级循环关系"), nil
		}
	}

	// 4. 更新上级
	err = l.svcCtx.EmployeeModel.UpdateSupervisor(l.ctx, req.EmployeeID, req.SupervisorID)
	if err != nil {
		logx.Errorf("更新员工上级失败: %v", err)
		return utils.Response.InternalError("更新失败"), err
	}

	logx.Infof("员工 %s 的直属上级已更新为 %s", employee.RealName, req.SupervisorID)

	return utils.Response.Success(map[string]interface{}{
		"message":      "设置成功",
		"employeeId":   req.EmployeeID,
		"supervisorId": req.SupervisorID,
	}), nil
}

// wouldCreateCycle 检查设置上级是否会形成循环
func (l *UpdateEmployeeSupervisorLogic) wouldCreateCycle(employeeID, newSupervisorID string) bool {
	// 从新上级开始，向上遍历，看是否会遇到当前员工
	currentID := newSupervisorID
	visited := make(map[string]bool)

	for i := 0; i < 100; i++ { // 最多检查100层，防止无限循环
		if currentID == "" {
			return false
		}
		if currentID == employeeID {
			return true // 发现循环
		}
		if visited[currentID] {
			return false // 已经访问过，说明有其他循环但不涉及当前员工
		}
		visited[currentID] = true

		// 获取当前员工的上级
		emp, err := l.svcCtx.EmployeeModel.FindOne(l.ctx, currentID)
		if err != nil {
			return false
		}
		if !emp.SupervisorId.Valid || emp.SupervisorId.String == "" {
			return false
		}
		currentID = emp.SupervisorId.String
	}
	return false
}
