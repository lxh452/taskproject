package admin

import (
	"context"

	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type DisableCompanyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 禁用公司
func NewDisableCompanyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisableCompanyLogic {
	return &DisableCompanyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DisableCompanyLogic) DisableCompany(req *types.DisableCompanyRequest) (resp *types.BaseResponse, err error) {
	// 验证公司是否存在
	company, err := l.svcCtx.CompanyModel.FindOne(l.ctx, req.CompanyID)
	if err != nil {
		logx.Errorf("查询公司失败: %v", err)
		return utils.Response.Error(404, "公司不存在"), nil
	}

	// 检查公司是否已被禁用
	if company.Status == 0 {
		return utils.Response.Error(400, "公司已被禁用"), nil
	}

	// 禁用公司
	err = l.svcCtx.CompanyModel.UpdateStatus(l.ctx, req.CompanyID, 0)
	if err != nil {
		logx.Errorf("禁用公司失败: %v", err)
		return utils.Response.Error(500, "禁用公司失败"), nil
	}

	// 级联禁用公司下所有员工
	employees, err := l.svcCtx.EmployeeModel.FindByCompanyID(l.ctx, req.CompanyID)
	if err != nil {
		logx.Errorf("查询公司员工失败: %v", err)
	} else {
		var employeeIDs []string
		for _, emp := range employees {
			employeeIDs = append(employeeIDs, emp.Id)
		}
		if len(employeeIDs) > 0 {
			err = l.svcCtx.EmployeeModel.BatchUpdateStatus(l.ctx, employeeIDs, 0)
			if err != nil {
				logx.Errorf("批量禁用员工失败: %v", err)
			}
		}
	}

	// 级联禁用公司下所有部门
	departments, err := l.svcCtx.DepartmentModel.FindByCompanyID(l.ctx, req.CompanyID)
	if err != nil {
		logx.Errorf("查询公司部门失败: %v", err)
	} else {
		for _, dept := range departments {
			err = l.svcCtx.DepartmentModel.UpdateStatus(l.ctx, dept.Id, 0)
			if err != nil {
				logx.Errorf("禁用部门 %s 失败: %v", dept.Id, err)
			}
		}
	}

	logx.Infof("公司 %s 已被禁用，原因: %s", req.CompanyID, req.Reason)

	// 记录系统日志
	if l.svcCtx.SystemLogService != nil {
		l.svcCtx.SystemLogService.AdminAction(l.ctx, "company", "disable", "禁用公司: "+req.CompanyID+", 原因: "+req.Reason, "", "", "")
	}

	return utils.Response.Success("公司已禁用"), nil
}
