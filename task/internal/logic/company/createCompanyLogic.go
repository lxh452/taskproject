// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package company

import (
	"context"
	"database/sql"
	"fmt"
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
	companyID := utils.Common.GenId("cp")
	employeeID := utils.Common.GenId("emp")

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
	// 注意：模板初始化改为异步处理，不阻塞主事务，避免高并发时数据库压力过大
	txCtx, cancel := context.WithTimeout(l.ctx, 30*time.Second)
	defer cancel()

	var founderDeptID, founderPosID, founderRoleID string

	err = l.svcCtx.TransactionService.TransactCtx(txCtx, func(ctx context.Context, session sqlx.Session) error {
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

		// 4.1. 更新职位的当前员工数（创始人已占用该职位）
		positionModelWithSession := l.svcCtx.TransactionHelper.GetPositionModelWithSession(session)
		if err := positionModelWithSession.UpdateCurrentEmployees(ctx, founderPosID, 1); err != nil {
			logx.Errorf("更新职位当前员工数失败: %v", err)
			return err
		}

		// 5. 创建超级管理员角色（拥有所有权限）
		roleModelWithSession := l.svcCtx.TransactionHelper.GetRoleModelWithSession(session)
		founderRoleID = utils.Common.GenId("role")
		// 所有权限码（根据 permdefs.go 中定义的权限）
		// 任务(1-5), 任务节点(10-13), 交接(20-23), 通知(30-32)
		// 公司(40-43), 部门(45-48), 职位(50-53), 角色(60-65), 员工(70-74)
		allPermissions := "[1,2,3,4,5,10,11,12,13,20,21,22,23,30,31,32,40,41,42,43,45,46,47,48,50,51,52,53,60,61,62,63,64,65,70,71,72,73,74]"
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
			GrantBy:    sql.NullString{String: userID, Valid: true},
			GrantTime:  time.Now(),
			ExpireTime: sql.NullTime{}, // 不过期
			Status:     1,              // 正常状态
			CreateTime: time.Now(),
		}
		if _, err := positionRoleModelWithSession.Insert(ctx, positionRole); err != nil {
			return err
		}

		// 注意：权限验证直接通过职位->角色->权限查询，无需同步到user_permission表

		// 7. 更新用户的加入公司状态（在同一事务中）
		userModelWithSession := l.svcCtx.TransactionHelper.GetUserModelWithSession(session)
		if err := userModelWithSession.UpdateHasJoinedCompany(ctx, userID, true); err != nil {
			logx.Errorf("更新用户加入公司状态失败: %v", err)
			return err
		}

		return nil
	})

	if err != nil {
		logx.Errorf("创建公司失败: %v", err)
		return utils.Response.InternalError("创建公司失败"), nil
	}

	// 如果用户选择了模板，异步初始化组织结构（不阻塞用户响应）
	if req.UseTemplate {
		go func() {
			// 使用独立的 context，避免超时
			templateCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

			logx.Infof("[Async] 开始异步初始化公司组织结构: companyID=%s", companyID)
			if err := l.svcCtx.ApplyDefaultOrgStructure(templateCtx, companyID); err != nil {
				logx.Errorf("[Async] 异步初始化公司组织结构失败: companyID=%s, error=%v", companyID, err)
				// 异步失败不影响主流程，可以后续手动触发或重试
			} else {
				logx.Infof("[Async] 异步初始化公司组织结构成功: companyID=%s", companyID)
			}
		}()
		logx.Infof("已启动异步初始化组织结构任务: companyID=%s", companyID)
	} else {
		logx.Infof("用户选择不使用模板，公司 %s 仅创建基础结构", companyID)
	}

	// 生成新的JWT令牌（包含员工信息），用于更新前端token
	// 这样用户就不需要重新登录了
	newToken := ""
	tokenErr := error(nil)
	newToken, tokenErr = l.svcCtx.JWTMiddleware.GenerateTokenWithEmployee(
		userID,
		userInfo.Username,
		userInfo.RealName.String,
		"user",
		employeeID,
		companyID,
	)
	if tokenErr != nil {
		logx.Errorf("生成新Token失败: %v", tokenErr)
		newToken = "" // 如果生成失败，返回空字符串，前端需要重新登录
	} else {
		// 更新Redis中的Token
		tokenKey := fmt.Sprintf("auth:token:%s", userID)
		if err := l.svcCtx.RedisClient.Setex(tokenKey, newToken, 86400); err != nil {
			logx.Errorf("更新Redis Token失败: %v", err)
		} else {
			logx.Infof("已更新Token: userId=%s, employeeId=%s, companyId=%s", userID, employeeID, companyID)
		}
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
		"useTemplate":  req.UseTemplate,
		"token":        newToken, // 返回新token，前端需要更新
	}), nil
}
