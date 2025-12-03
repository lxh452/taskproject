// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"context"
	"database/sql"
	"time"

	companyModel "task_Project/model/company"
	roleModel "task_Project/model/role"
	"task_Project/model/user"
	"task_Project/task/internal/svc"
	"task_Project/task/internal/types"
	"task_Project/task/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateCompanyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 创建公司
func NewCreateCompanyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCompanyLogic {
	return &CreateCompanyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCompanyLogic) CreateCompany(req *types.CreateCompanyRequest) (resp *types.BaseResponse, err error) {
	// 参数验证
	if utils.Validator.IsEmpty(req.Name) {
		return utils.Response.BusinessError("company_name_required"), nil
	}

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

	// 检查用户是否已经加入其他公司
	existingEmployee, _ := l.svcCtx.EmployeeModel.FindByUserID(l.ctx, userID)
	if existingEmployee != nil {
		return utils.Response.BusinessError("user_already_in_company"), nil
	}

	// 检查公司名称是否已存在
	existingCompanies, err := l.svcCtx.CompanyModel.FindByOwner(l.ctx, userID)
	if err != nil {
		logx.Errorf("查询用户公司失败: %v", err)
		return utils.Response.InternalError("查询用户公司失败"), nil
	}

	for _, existingCompany := range existingCompanies {
		if existingCompany.Name == req.Name {
			return utils.Response.BusinessError("company_name_exists"), nil
		}
	}

	// 生成公司ID
	companyID := utils.Common.GenerateID()
	employeeID := utils.Common.GenerateID()

	// 创建公司
	companyInfo := &companyModel.Company{
		Id:                companyID,
		Name:              req.Name,
		CompanyAttributes: int64(req.CompanyAttributes),
		CompanyBusiness:   int64(req.CompanyBusiness),
		Owner:             userID,
		Description:       utils.Common.ToSqlNullString(req.Description),
		Address:           utils.Common.ToSqlNullString(req.Address),
		Phone:             utils.Common.ToSqlNullString(req.Phone),
		Email:             utils.Common.ToSqlNullString(req.Email),
		Status:            1, // 正常状态
		CreateTime:        time.Now(),
		UpdateTime:        time.Now(),
	}

	// 使用事务创建公司、部门、职位、员工和权限
	var founderDeptID, founderPosID, founderRoleID string

	err = l.svcCtx.TransactionService.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		// 1. 插入公司
		companyModelWithSession := l.svcCtx.TransactionHelper.GetCompanyModelWithSession(session)
		if _, err := companyModelWithSession.Insert(ctx, companyInfo); err != nil {
			return err
		}

		// 2. 创建创始人专属部门（管理层/总裁办）
		depModel := l.svcCtx.TransactionHelper.GetDepartmentModelWithSession(session)
		founderDeptID = utils.Common.GenId("dept")
		founderDept := &companyModel.Department{
			Id:                 founderDeptID,
			CompanyId:          companyID,
			ParentId:           sql.NullString{},
			DepartmentName:     "总裁办",
			DepartmentCode:     utils.Common.ToSqlNullString("CEO"),
			DepartmentPriority: 100, // 最高优先级
			ManagerId:          sql.NullString{String: employeeID, Valid: true},
			Description:        sql.NullString{String: "公司创始人与最高管理层", Valid: true},
			Status:             1,
			CreateTime:         time.Now(),
			UpdateTime:         time.Now(),
		}
		if _, err := depModel.Insert(ctx, founderDept); err != nil {
			return err
		}

		// 3. 创建创始人职位（CEO/创始人）
		posModel := l.svcCtx.TransactionHelper.GetPositionModelWithSession(session)
		founderPosID = utils.Common.GenId("pos")
		founderPos := &companyModel.Position{
			Id:               founderPosID,
			DepartmentId:     founderDeptID,
			PositionName:     "创始人",
			PositionCode:     utils.Common.ToSqlNullString("FOUNDER"),
			JobType:          2,  // 管理类
			PositionLevel:    10, // 最高级别
			RequiredSkills:   sql.NullString{},
			JobDescription:   sql.NullString{String: "公司创始人，拥有最高权限", Valid: true},
			Responsibilities: sql.NullString{},
			Requirements:     sql.NullString{},
			SalaryRangeMin:   sql.NullFloat64{},
			SalaryRangeMax:   sql.NullFloat64{},
			IsManagement:     1, // 是管理岗
			MaxEmployees:     1,
			CurrentEmployees: 1,
			Status:           1,
			CreateTime:       time.Now(),
			UpdateTime:       time.Now(),
		}
		if _, err := posModel.Insert(ctx, founderPos); err != nil {
			return err
		}

		// 4. 创建创始人员工记录
		empModel := l.svcCtx.TransactionHelper.GetEmployeeModelWithSession(session)
		realName := "创始人"
		if userInfo.RealName.Valid && userInfo.RealName.String != "" {
			realName = userInfo.RealName.String
		}
		founderEmp := &user.Employee{
			Id:           employeeID,
			UserId:       userID,
			CompanyId:    companyID,
			DepartmentId: utils.Common.ToSqlNullString(founderDeptID),
			PositionId:   utils.Common.ToSqlNullString(founderPosID),
			EmployeeId:   "FOUNDER-001",
			RealName:     realName,
			Email:        userInfo.Email,
			Phone:        userInfo.Phone,
			Skills:       sql.NullString{},
			RoleTags:     utils.Common.ToSqlNullString("创始人,管理员"),
			HireDate:     sql.NullTime{Time: time.Now(), Valid: true},
			Status:       1,
			CreateTime:   time.Now(),
			UpdateTime:   time.Now(),
		}
		if _, err := empModel.Insert(ctx, founderEmp); err != nil {
			return err
		}

		// 5. 创建超级管理员角色（拥有所有权限）
		roleModelWithSession := l.svcCtx.TransactionHelper.GetRoleModelWithSession(session)
		founderRoleID = utils.Common.GenId("role")
		// 所有权限码（根据 authz.go 中定义的权限）
		allPermissions := "[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50]"
		founderRole := &roleModel.Role{
			Id:              founderRoleID,
			CompanyId:       companyID,
			RoleName:        "超级管理员",
			RoleCode:        "SUPER_ADMIN",
			RoleDescription: utils.Common.ToSqlNullString("公司创始人专属角色，拥有所有权限"),
			IsSystem:        1, // 系统角色
			Permissions:     utils.Common.ToSqlNullString(allPermissions),
			Status:          1,
			CreateTime:      time.Now(),
			UpdateTime:      time.Now(),
		}
		if _, err := roleModelWithSession.Insert(ctx, founderRole); err != nil {
			return err
		}

		// 6. 将角色绑定到职位
		positionRoleModelWithSession := l.svcCtx.TransactionHelper.GetPositionRoleModelWithSession(session)
		positionRoleID := utils.Common.GenId("pr")
		positionRole := &roleModel.PositionRole{
			Id:         positionRoleID,
			PositionId: founderPosID,
			RoleId:     founderRoleID,
			CreateTime: time.Now(),
		}
		if _, err := positionRoleModelWithSession.Insert(ctx, positionRole); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logx.Errorf("创建公司失败: %v", err)
		return utils.Response.InternalError("创建公司失败"), nil
	}

	// 复制系统默认部门与职位（其他部门）
	if err := l.svcCtx.ApplyDefaultOrgStructure(l.ctx, companyID); err != nil {
		logx.Errorf("初始化公司组织结构失败: %v", err)
		// 不影响主流程，创始人部门和职位已经创建
	}

	// 更新用户的加入公司状态
	if updateErr := l.svcCtx.UserModel.UpdateHasJoinedCompany(l.ctx, userID, true); updateErr != nil {
		logx.Errorf("更新用户加入公司状态失败: %v", updateErr)
		// 不影响主流程，继续执行
	}

	// 发送创建成功通知邮件
	go func() {
		logx.Infof("用户 %s 创建公司成功: %s, 员工ID: %s", userID, req.Name, employeeID)
	}()

	return utils.Response.SuccessWithKey("create", map[string]interface{}{
		"companyId":    companyID,
		"employeeId":   employeeID,
		"departmentId": founderDeptID,
		"positionId":   founderPosID,
		"roleId":       founderRoleID,
		"name":         req.Name,
		"bootstraped":  true,
	}), nil
}
