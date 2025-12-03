// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package employee

import (
	"context"
	"database/sql"
	"time"

	"task_Project/model/user"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type JoinCompanyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 加入公司
func NewJoinCompanyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JoinCompanyLogic {
	return &JoinCompanyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *JoinCompanyLogic) JoinCompany(req *types.JoinCompanyRequest) (resp *types.BaseResponse, err error) {
	// 获取当前用户ID
	userID, ok := utils.Common.GetCurrentUserID(l.ctx)
	if !ok {
		return utils.Response.UnauthorizedError(), nil
	}

	// 获取用户信息
	userInfo, err := l.svcCtx.UserModel.FindOne(l.ctx, userID)
	if err != nil {
		logx.Errorf("查询用户失败: %v", err)
		return utils.Response.InternalError("查询用户失败"), nil
	}

	// 检查用户是否已经加入公司
	existingEmployee, _ := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, userID)
	if existingEmployee != nil {
		return utils.Response.BusinessError("user_already_in_company"), nil
	}

	// 验证公司是否存在
	companyInfo, err := l.svcCtx.CompanyModel.FindOne(l.ctx, req.CompanyID)
	if err != nil {
		logx.Errorf("查询公司失败: %v", err)
		return utils.Response.BusinessError("company_not_found"), nil
	}

	// 获取部门ID（如果未指定，使用默认部门）
	departmentID := req.DepartmentID
	if departmentID == "" {
		// 查找公司的默认部门（如人力资源部或第一个部门）
		departments, err := l.svcCtx.DepartmentModel.FindByCompanyID(l.ctx, req.CompanyID)
		if err != nil || len(departments) == 0 {
			logx.Errorf("查询公司部门失败: %v", err)
			return utils.Response.InternalError("公司尚未设置部门结构"), nil
		}
		// 优先找人力资源部，否则用第一个部门
		for _, dept := range departments {
			if dept.DepartmentCode.Valid && dept.DepartmentCode.String == "HR" {
				departmentID = dept.Id
				break
			}
		}
		if departmentID == "" {
			departmentID = departments[0].Id
		}
	}

	// 获取职位ID（如果未指定，使用默认职位）
	positionID := req.PositionID
	if positionID == "" {
		// 查找该部门的默认职位（如助理或第一个职位）
		positions, err := l.svcCtx.PositionModel.FindByDepartmentID(l.ctx, departmentID)
		if err != nil || len(positions) == 0 {
			logx.Errorf("查询部门职位失败: %v", err)
			return utils.Response.InternalError("部门尚未设置职位"), nil
		}
		// 优先找助理职位，否则用第一个职位
		for _, pos := range positions {
			if pos.PositionCode.Valid && pos.PositionCode.String == "AST" {
				positionID = pos.Id
				break
			}
		}
		if positionID == "" {
			positionID = positions[0].Id
		}
	}

	// 获取用户真实姓名
	realName := "新员工"
	if userInfo.RealName.Valid && userInfo.RealName.String != "" {
		realName = userInfo.RealName.String
	}

	// 创建员工记录
	employeeID := utils.Common.GenerateID()
	employeeInfo := &user.Employee{
		Id:           employeeID,
		UserId:       userID,
		CompanyId:    req.CompanyID,
		DepartmentId: utils.Common.ToSqlNullString(departmentID),
		PositionId:   utils.Common.ToSqlNullString(positionID),
		EmployeeId:   "", // 业务工号可以后续分配
		RealName:     realName,
		Email:        userInfo.Email,
		Phone:        userInfo.Phone,
		Skills:       sql.NullString{},
		RoleTags:     sql.NullString{},
		HireDate:     sql.NullTime{Time: time.Now(), Valid: true},
		Status:       1,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}

	_, err = l.svcCtx.EmployeeModel.Insert(l.ctx, employeeInfo)
	if err != nil {
		logx.Errorf("创建员工记录失败: %v", err)
		return utils.Response.InternalError("加入公司失败"), nil
	}

	// 更新用户的加入公司状态
	if updateErr := l.svcCtx.UserModel.UpdateHasJoinedCompany(l.ctx, userID, true); updateErr != nil {
		logx.Errorf("更新用户加入公司状态失败: %v", updateErr)
	}

	// 更新职位的当前员工数
	positionInfo, _ := l.svcCtx.PositionModel.FindOne(l.ctx, positionID)
	if positionInfo != nil {
		_ = l.svcCtx.PositionModel.UpdateCurrentEmployees(l.ctx, positionID, int(positionInfo.CurrentEmployees)+1)
	}

	logx.Infof("用户 %s 成功加入公司 %s, 员工ID: %s", userID, companyInfo.Name, employeeID)

	return utils.Response.SuccessWithKey("join", map[string]interface{}{
		"employeeId":   employeeID,
		"companyId":    req.CompanyID,
		"companyName":  companyInfo.Name,
		"departmentId": departmentID,
		"positionId":   positionID,
	}), nil
}


